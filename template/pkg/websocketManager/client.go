package websocketManager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// 客戶端常量
const (
	// 重連設置
	maxClientReconnectAttempts = 5
	baseClientReconnectDelay   = 2 * time.Second
	maxClientReconnectDelay    = 30 * time.Second

	// 客戶端心跳間隔 (15秒)
	clientHeartbeatInterval = 15 * time.Second

	// 客戶端超時設置
	clientReadTimeout  = 35 * time.Second
	clientWriteTimeout = 10 * time.Second
	clientWaitTimeout  = 60 * time.Second

	// 消息緩衝區大小
	clientSendBufferSize = 100
	clientRecvBufferSize = 100
)

// WebSocketClient 提供一個易用的WebSocket客戶端實現
type WebSocketClient struct {
	// 連接信息
	URL     string
	Conn    *websocket.Conn
	Token   string
	Headers map[string]string

	// 狀態控制
	State           ClientState
	reconnectCount  int
	isConnected     bool
	shouldReconnect bool

	// 通道
	Send    chan []byte
	Receive chan []byte
	done    chan struct{}

	// 定時器
	heartbeatTicker *time.Ticker
	reconnectTicker *time.Ticker

	// 事件處理器
	OnConnect   func()
	OnMessage   func(data []byte)
	OnError     func(err error)
	OnClose     func()
	OnReconnect func(attempt int)
}

// 創建新的WebSocket客戶端
func NewWebSocketClient(url string, token string) *WebSocketClient {
	return &WebSocketClient{
		URL:             url,
		Token:           token,
		Headers:         make(map[string]string),
		State:           StateDisconnected,
		reconnectCount:  0,
		shouldReconnect: true,
		Send:            make(chan []byte, clientSendBufferSize),
		Receive:         make(chan []byte, clientRecvBufferSize),
		done:            make(chan struct{}),
	}
}

// 連接WebSocket服務器
func (c *WebSocketClient) Connect() error {
	if c.isConnected {
		return nil
	}

	c.State = StateConnecting

	u, err := url.Parse(c.URL)
	if err != nil {
		if c.OnError != nil {
			c.OnError(fmt.Errorf("無效的URL: %v", err))
		}
		c.State = StateFailed
		return err
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	// 準備請求頭
	header := make(map[string][]string)
	for k, v := range c.Headers {
		header[k] = []string{v}
	}

	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		if c.OnError != nil {
			c.OnError(fmt.Errorf("連接失敗: %v", err))
		}
		c.State = StateFailed

		// 如果設置了自動重連，啟動重連
		if c.shouldReconnect {
			go c.startReconnecting()
		}

		return err
	}

	c.Conn = conn
	c.isConnected = true
	c.reconnectCount = 0
	c.State = StateConnected

	// 如果提供了令牌，立即進行身份驗證
	if c.Token != "" {
		c.authenticate()
	}

	// 啟動收發協程
	go c.readPump()
	go c.writePump()

	// 啟動心跳
	c.startHeartbeat()

	// 觸發連接回調
	if c.OnConnect != nil {
		c.OnConnect()
	}

	return nil
}

// 身份驗證
func (c *WebSocketClient) authenticate() {
	authMsg := Message{
		Type:    "auth",
		Content: c.Token,
	}
	msgBytes, _ := json.Marshal(authMsg)
	c.Send <- msgBytes
}

// 讀取協程
func (c *WebSocketClient) readPump() {
	defer func() {
		c.Close()
	}()

	c.Conn.SetReadLimit(4096)
	c.Conn.SetReadDeadline(time.Now().Add(clientReadTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(clientReadTimeout))
		return nil
	})

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				if c.OnError != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					c.OnError(fmt.Errorf("讀取錯誤: %v", err))
				}
				return
			}

			// 重置讀取超時
			c.Conn.SetReadDeadline(time.Now().Add(clientReadTimeout))

			// 處理收到的消息
			var msg Message
			if err := json.Unmarshal(message, &msg); err == nil {
				// 處理心跳響應，無需通知應用層
				if msg.Type == "heartbeat" {
					continue
				}
			}

			// 將消息放入接收通道
			select {
			case c.Receive <- message:
			default:
				log.Println("WebSocketClient: 接收通道已滿，丟棄消息")
			}

			// 調用消息回調
			if c.OnMessage != nil {
				c.OnMessage(message)
			}
		}
	}
}

// 寫入協程
func (c *WebSocketClient) writePump() {
	defer func() {
		if c.heartbeatTicker != nil {
			c.heartbeatTicker.Stop()
		}
	}()

	for {
		select {
		case <-c.done:
			return
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(clientWriteTimeout))
			if !ok {
				// 通道已關閉
				c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				if c.OnError != nil {
					c.OnError(fmt.Errorf("創建寫入器失敗: %v", err))
				}
				return
			}

			w.Write(message)

			// 添加隊列中其他消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				if c.OnError != nil {
					c.OnError(fmt.Errorf("關閉寫入器失敗: %v", err))
				}
				return
			}
		}
	}
}

// 發送消息
func (c *WebSocketClient) SendMessage(msgType string, content interface{}) error {
	if !c.isConnected {
		return fmt.Errorf("客戶端未連接")
	}

	msg := Message{
		Type:    msgType,
		Content: content,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("消息序列化失敗: %v", err)
	}

	select {
	case c.Send <- msgBytes:
		return nil
	default:
		return fmt.Errorf("發送通道已滿")
	}
}

// 啟動心跳
func (c *WebSocketClient) startHeartbeat() {
	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
	}

	c.heartbeatTicker = time.NewTicker(clientHeartbeatInterval)

	go func() {
		for {
			select {
			case <-c.done:
				return
			case <-c.heartbeatTicker.C:
				// 發送心跳消息
				heartbeat := HeartbeatMessage{
					Type:      "heartbeat",
					Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
				}
				heartbeatBytes, _ := json.Marshal(heartbeat)

				// 嘗試發送心跳
				err := c.Conn.WriteControl(
					websocket.PingMessage,
					[]byte{},
					time.Now().Add(clientWriteTimeout),
				)

				if err != nil {
					if c.OnError != nil {
						c.OnError(fmt.Errorf("心跳發送失敗: %v", err))
					}
					c.Close()
					return
				}

				// 發送應用層心跳
				select {
				case c.Send <- heartbeatBytes:
				default:
					log.Println("WebSocketClient: 發送通道已滿，無法發送心跳")
				}
			}
		}
	}()
}

// 啟動重連
func (c *WebSocketClient) startReconnecting() {
	if c.reconnectTicker != nil {
		c.reconnectTicker.Stop()
	}

	// 初始重連延遲
	delay := baseClientReconnectDelay
	c.State = StateConnecting

	for c.shouldReconnect && c.reconnectCount < maxClientReconnectAttempts {
		c.reconnectCount++

		// 指數退避
		if c.reconnectCount > 1 {
			delay = time.Duration(1<<uint(c.reconnectCount-1)) * baseClientReconnectDelay
			if delay > maxClientReconnectDelay {
				delay = maxClientReconnectDelay
			}
		}

		log.Printf("WebSocketClient: 嘗試重連 #%d，等待 %v\n", c.reconnectCount, delay)

		// 觸發重連回調
		if c.OnReconnect != nil {
			c.OnReconnect(c.reconnectCount)
		}

		time.Sleep(delay)

		// 嘗試重新連接
		err := c.Connect()
		if err == nil {
			log.Printf("WebSocketClient: 重連成功\n")
			return
		}

		log.Printf("WebSocketClient: 重連 #%d 失敗: %v\n", c.reconnectCount, err)
	}

	if c.reconnectCount >= maxClientReconnectAttempts {
		log.Printf("WebSocketClient: 達到最大重連次數 %d，放棄重連\n", maxClientReconnectAttempts)
		c.State = StateFailed

		// 連接失敗，通知上層
		if c.OnError != nil {
			c.OnError(fmt.Errorf("達到最大重連次數 %d，放棄重連", maxClientReconnectAttempts))
		}

		// 觸發關閉事件
		if c.OnClose != nil {
			c.OnClose()
		}
	}
}

// 關閉連接
func (c *WebSocketClient) Close() {
	if !c.isConnected {
		return
	}

	c.shouldReconnect = false
	c.isConnected = false
	c.State = StateDisconnected

	// 發送關閉信號
	close(c.done)

	// 停止所有定時器
	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
	}

	if c.reconnectTicker != nil {
		c.reconnectTicker.Stop()
	}

	// 關閉WebSocket連接
	if c.Conn != nil {
		c.Conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "客戶端關閉連接"),
		)
		c.Conn.Close()
	}

	// 觸發關閉回調
	if c.OnClose != nil {
		c.OnClose()
	}
}

// 設置是否應該自動重連
func (c *WebSocketClient) SetAutoReconnect(autoReconnect bool) {
	c.shouldReconnect = autoReconnect
}

// 獲取當前連接狀態
func (c *WebSocketClient) GetState() ClientState {
	return c.State
}

// 添加自定義請求頭
func (c *WebSocketClient) AddHeader(key, value string) {
	c.Headers[key] = value
}
