package websocketManager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// 服務端示例代碼
func ServerExample() {
	// 創建一個簡單的身份驗證函數
	authFunc := func(token string) (uint, error) {
		// 這裡只是示例，實際應用中應該驗證令牌的有效性
		if token == "valid-token" {
			return 1, nil // 返回用戶ID
		}
		return 0, fmt.Errorf("無效的令牌")
	}

	// 創建WebSocket管理器
	manager := NewManager(authFunc)

	// 創建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 啟動WebSocket管理器
	go manager.Start(ctx)

	// 配置WebSocket升級器
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 在生產環境中，應該檢查來源
		},
	}

	// 設置WebSocket處理器
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("升級連接失敗: %v", err)
			return
		}

		// 創建客戶端並分配唯一ID
		client := &Client{
			ID:           fmt.Sprintf("%p", conn),
			Conn:         conn,
			Send:         make(chan []byte, 256),
			manager:      manager,
			LastActivity: time.Now(),
			IsAuthed:     false,
		}

		// 將客戶端註冊到管理器
		manager.register <- client

		// 啟動客戶端讀寫協程
		go client.ReadPump()
		go client.WritePump()
	})

	// 每5秒向所有客戶端廣播一條消息
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 創建廣播消息
				message := Message{
					Type:    "broadcast",
					Content: fmt.Sprintf("服務器時間: %s", time.Now().Format(time.RFC3339)),
				}

				// 序列化消息
				msgBytes, err := json.Marshal(message)
				if err != nil {
					log.Printf("消息序列化失敗: %v", err)
					continue
				}

				// 廣播消息
				manager.broadcast <- msgBytes
			}
		}
	}()

	// 啟動HTTP服務器
	log.Println("啟動WebSocket服務器在 :8080...")
	http.ListenAndServe(":8080", nil)
}

// 客戶端示例代碼
func ClientExample() {
	// 創建WebSocket客戶端
	client := NewWebSocketClient("ws://localhost:8080/ws", "valid-token")

	// 設置事件處理器
	client.OnConnect = func() {
		fmt.Println("已連接到服務器!")
	}

	client.OnMessage = func(data []byte) {
		fmt.Printf("收到消息: %s\n", string(data))
	}

	client.OnError = func(err error) {
		fmt.Printf("錯誤: %v\n", err)
	}

	client.OnClose = func() {
		fmt.Println("連接已關閉")
	}

	client.OnReconnect = func(attempt int) {
		fmt.Printf("嘗試重連 #%d...\n", attempt)
	}

	// 連接到服務器
	err := client.Connect()
	if err != nil {
		fmt.Printf("連接失敗: %v\n", err)
		return
	}

	// 發送自定義消息
	go func() {
		// 等待連接建立
		time.Sleep(1 * time.Second)

		// 發送消息
		err := client.SendMessage("chat", "你好，WebSocket服務器!")
		if err != nil {
			fmt.Printf("發送消息失敗: %v\n", err)
		}
	}()

	// 持續運行60秒
	time.Sleep(60 * time.Second)

	// 關閉連接
	client.Close()
}

// 實用的WebSocket連接配置函數
func ConfigureWebSocketUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			// 在生產環境中，應檢查Origin
			return true
		},
		// 設置60秒的握手超時
		HandshakeTimeout: 60 * time.Second,
	}
}

// 實用的WebSocket連接配置函數
func ConfigureWebSocketConnection(conn *websocket.Conn) {
	// 設置讀取限制為4MB
	conn.SetReadLimit(4 * 1024 * 1024)

	// 設置讀取超時為30秒
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// 設置寫入超時為10秒
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// 設置Ping/Pong處理器
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		return nil
	})

	conn.SetPingHandler(func(message string) error {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err := conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(5*time.Second))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err
	})
}
