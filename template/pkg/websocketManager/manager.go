package websocketManager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// 客戶端結構體，代表一個 WebSocket 連接
type Client struct {
	ID           string          // 客戶端唯一標識
	UserID       uint            // 用戶 ID
	Conn         *websocket.Conn // WebSocket 連接
	Send         chan []byte     // 發送訊息的通道
	manager      *Manager        // 所屬的管理器
	LastActivity time.Time       // 最後活動時間
	IsAuthed     bool            // 是否已認證
}

// 訊息結構體
type Message struct {
	Type    string      `json:"type"`              // 訊息類型
	Content interface{} `json:"content,omitempty"` // 訊息內容
	From    string      `json:"from,omitempty"`    // 發送方
	To      string      `json:"to,omitempty"`      // 接收方
}

// WebSocket 管理器結構體
type Manager struct {
	clients     map[string]*Client               // 所有已連接的客戶端
	userClients map[uint]map[string]bool         // 用戶ID對應的客戶端map
	broadcast   chan []byte                      // 廣播訊息的通道
	register    chan *Client                     // 註冊客戶端的通道
	unregister  chan *Client                     // 註銷客戶端的通道
	shutdown    chan struct{}                    // 關閉的通道
	auth        func(token string) (uint, error) // 認證函數
	mutex       sync.RWMutex                     // 讀寫鎖
}

// 創建新的 WebSocket 管理器
func NewManager(authFunc func(token string) (uint, error)) *Manager {
	return &Manager{
		clients:     make(map[string]*Client),
		userClients: make(map[uint]map[string]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		shutdown:    make(chan struct{}),
		auth:        authFunc,
		mutex:       sync.RWMutex{},
	}
}

// 啟動 WebSocket 管理器
func (manager *Manager) Start(ctx context.Context) {
	log.Println("WebSocket Manager: Starting...")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("WebSocket Manager: Shutting down...")
			return

		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client.ID] = client
			manager.mutex.Unlock()
			log.Printf("WebSocket Manager: Client %s registered\n", client.ID)

		case client := <-manager.unregister:
			if _, ok := manager.clients[client.ID]; ok {
				manager.mutex.Lock()

				// 從用戶-客戶端映射中移除
				if client.IsAuthed {
					delete(manager.userClients[client.UserID], client.ID)
					// 如果用戶沒有其他客戶端連接，則刪除用戶映射
					if len(manager.userClients[client.UserID]) == 0 {
						delete(manager.userClients, client.UserID)
					}
				}

				// 關閉發送通道
				close(client.Send)
				// 從客戶端列表中刪除
				delete(manager.clients, client.ID)

				manager.mutex.Unlock()
				log.Printf("WebSocket Manager: Client %s unregistered\n", client.ID)
			}

		case message := <-manager.broadcast:
			// 廣播消息給所有客戶端
			manager.mutex.RLock()
			for id, client := range manager.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					manager.mutex.RUnlock()
					manager.mutex.Lock()

					// 從用戶-客戶端映射中移除
					if client.IsAuthed {
						delete(manager.userClients[client.UserID], client.ID)
						if len(manager.userClients[client.UserID]) == 0 {
							delete(manager.userClients, client.UserID)
						}
					}

					delete(manager.clients, id)
					manager.mutex.Unlock()
					manager.mutex.RLock()
					log.Printf("WebSocket Manager: Client %s removed (failed to send)\n", id)
				}
			}
			manager.mutex.RUnlock()

		case <-ticker.C:
			// 定期檢查非活躍連接，超過 5 分鐘未活動則斷開
			threshold := time.Now().Add(-5 * time.Minute)
			manager.mutex.Lock()
			for id, client := range manager.clients {
				if client.LastActivity.Before(threshold) {
					close(client.Send)

					// 從用戶-客戶端映射中移除
					if client.IsAuthed {
						delete(manager.userClients[client.UserID], client.ID)
						if len(manager.userClients[client.UserID]) == 0 {
							delete(manager.userClients, client.UserID)
						}
					}

					delete(manager.clients, id)
					log.Printf("WebSocket Manager: Client %s removed due to inactivity\n", id)
				}
			}
			manager.mutex.Unlock()
		}
	}
}

// 驗證客戶端
func (manager *Manager) AuthenticateClient(client *Client, token string) error {
	userID, err := manager.auth(token)
	if err != nil {
		return err
	}

	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	client.UserID = userID
	client.IsAuthed = true

	// 將客戶端添加到用戶-客戶端映射
	if _, exists := manager.userClients[userID]; !exists {
		manager.userClients[userID] = make(map[string]bool)
	}
	manager.userClients[userID][client.ID] = true

	log.Printf("WebSocket Manager: Client %s authenticated for user %d\n", client.ID, userID)
	return nil
}

// 廣播訊息給所有已認證的客戶端
func (manager *Manager) BroadcastToAll(message interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	manager.broadcast <- msgBytes
	return nil
}

// 向指定用戶發送訊息
func (manager *Manager) SendToUser(userID uint, message interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	// 檢查用戶是否有連接的客戶端
	clientMap, exists := manager.userClients[userID]
	if !exists || len(clientMap) == 0 {
		return fmt.Errorf("no connected clients for user %d", userID)
	}

	// 發送訊息給用戶的所有客戶端
	for clientID := range clientMap {
		if client, ok := manager.clients[clientID]; ok && client.IsAuthed {
			select {
			case client.Send <- msgBytes:
				// 訊息已送入通道
			default:
				// 發送通道已滿或已關閉，移除客戶端
				close(client.Send)
				delete(manager.userClients[userID], clientID)
				delete(manager.clients, clientID)
				log.Printf("WebSocket Manager: Client %s removed (failed to send to user)\n", clientID)
			}
		}
	}

	return nil
}

// 客戶端讀取訊息
func (client *Client) ReadPump() {
	defer func() {
		client.manager.unregister <- client
		client.Conn.Close()
	}()

	// 設置讀取參數
	client.Conn.SetReadLimit(4096) // 4KB
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastActivity = time.Now()
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket Manager: Client %s unexpected close: %v\n", client.ID, err)
			}
			break
		}

		client.LastActivity = time.Now()

		// 處理接收到的訊息
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("WebSocket Manager: Client %s sent invalid message: %v\n", client.ID, err)
			continue
		}

		// 處理身份驗證
		if msg.Type == "auth" {
			token, ok := msg.Content.(string)
			if !ok {
				log.Printf("WebSocket Manager: Client %s sent invalid auth token\n", client.ID)
				continue
			}

			if err := client.manager.AuthenticateClient(client, token); err != nil {
				errMsg := Message{Type: "error", Content: "Authentication failed: " + err.Error()}
				errBytes, _ := json.Marshal(errMsg)
				client.Send <- errBytes
				log.Printf("WebSocket Manager: Client %s auth failed: %v\n", client.ID, err)
				continue
			}

			// 發送認證成功訊息
			successMsg := Message{Type: "auth_success", Content: "Authentication successful"}
			successBytes, _ := json.Marshal(successMsg)
			client.Send <- successBytes
			continue
		}

		// 檢查是否已認證
		if !client.IsAuthed {
			errMsg := Message{Type: "error", Content: "Not authenticated"}
			errBytes, _ := json.Marshal(errMsg)
			client.Send <- errBytes
			continue
		}

		// 處理其他類型的訊息...
		// 這裡可以根據 msg.Type 處理不同類型的訊息
	}
}

// 客戶端寫入訊息
func (client *Client) WritePump() {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// 通道已關閉
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 添加排隊的訊息
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 關閉 WebSocket 管理器
func (manager *Manager) Shutdown() {
	log.Println("WebSocket Manager: Shutting down...")
	close(manager.shutdown)
}
