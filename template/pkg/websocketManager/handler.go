package websocketManager

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocket 連接升級器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允許所有來源的跨域請求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHandler 提供 WebSocket 連接處理
type WebSocketHandler struct {
	manager *Manager
}

// NewWebSocketHandler 創建新的 WebSocket 處理程序
func NewWebSocketHandler(manager *Manager) *WebSocketHandler {
	return &WebSocketHandler{
		manager: manager,
	}
}

// HandleConnection 處理 WebSocket 連接請求
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	// 升級 HTTP 連接為 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket Handler: Failed to upgrade connection: %v\n", err)
		return
	}

	// 生成客戶端 ID
	clientID := uuid.New().String()

	// 創建新的客戶端
	client := &Client{
		ID:           clientID,
		Conn:         conn,
		Send:         make(chan []byte, 256), // 緩沖區設置為 256 條訊息
		manager:      h.manager,
		LastActivity: time.Now(),
		IsAuthed:     false,
	}

	// 註冊客戶端
	h.manager.register <- client

	// 啟動訊息讀取和寫入
	go client.ReadPump()
	go client.WritePump()

	log.Printf("WebSocket Handler: New connection established for client %s\n", clientID)
}

// 在 Core 模組中提供 WebSocketManager
func ProvideWebSocketManager(auth func(token string) (uint, error)) *Manager {
	return NewManager(auth)
}
