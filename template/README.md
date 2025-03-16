# 微服務架構項目

這是一個基於 Go 的微服務架構項目，使用 fx 依賴注入框架和模組化設計，提供遊戲和認證相關服務。

## 目錄

- [架構概述](#架構概述)
- [技術堆疊](#技術堆疊)
- [項目結構](#項目結構)
- [模組說明](#模組說明)
- [開始使用](#開始使用)
- [環境配置](#環境配置)
- [開發指南](#開發指南)
- [API 文檔](#api-文檔)
- [WebSocket 使用](#websocket-使用)
- [依賴注入機制](#依賴注入機制)
- [貢獻指南](#貢獻指南)

## 架構概述

本項目採用模組化的微服務架構，包含以下主要組件：

1. **核心基礎設施** - 提供通用功能如數據庫連接、Redis 緩存、Nacos 配置等
2. **業務服務** - 提供業務邏輯，如用戶認證、遊戲邏輯等
3. **API 層** - 提供 RESTful API 和 WebSocket 接口
4. **配置管理** - 支持本地和 Nacos 配置中心

系統採用微服務架構，目前包含：
- **auth-service** - 處理用戶認證和用戶管理
- **slot-game1** - 處理遊戲業務邏輯

## 技術堆疊

- **Go 1.21+** - 核心編程語言
- **Gin** - Web 框架
- **GORM** - ORM 數據庫操作
- **fx** - 依賴注入框架
- **PostgreSQL** - 關係型數據庫
- **Redis** - 緩存和會話管理
- **Nacos** - 配置中心和服務發現
- **WebSocket** - 實時通訊
- **Swagger** - API 文檔
- **JWT** - 用戶認證

## 項目結構

```
├── main.go                 # 主程序入口
├── .air.toml               # Air 熱重載配置
├── .env                    # 環境變量
├── .env.auth-service       # 認證服務環境變量
├── .env.slot-game1         # 遊戲服務環境變量
├── go.mod                  # Go 模組定義
├── src/                    # 應用程序源碼
│   ├── config/             # 配置管理
│   ├── handler/            # HTTP 處理程序
│   ├── middleware/         # 中間件
│   ├── service/            # 業務服務
│   └── interfaces/         # 接口定義
├── pkg/                    # 核心基礎設施包
│   ├── core/               # 核心模組整合
│   ├── databaseManager/    # 數據庫管理
│   ├── redisManager/       # Redis 管理
│   ├── nacosManager/       # Nacos 管理
│   ├── websocketManager/   # WebSocket 管理
│   ├── logger/             # 日誌管理
│   └── utils/              # 通用工具
├── docs/                   # Swagger 文檔
└── migrations/             # 數據庫遷移腳本
```

## 模組說明

### 核心模組 (pkg/core)

核心模組整合所有基礎設施，提供一致的訪問接口。包含：

- **DatabaseModule** - 數據庫連接和操作
- **RedisModule** - Redis 緩存和會話管理
- **WebSocketModule** - WebSocket 實時通訊
- **LoggerModule** - 日誌管理

### 配置模組 (src/config)

處理應用程序配置，支持從環境變量和 Nacos 加載配置。

### 服務模組 (src/service)

提供業務邏輯實現：

- **UserService** - 用戶管理
- **AuthService** - 認證服務

### 處理程序模組 (src/handler)

處理 HTTP 請求和路由：

- **AuthHandler** - 處理認證相關請求
- **UserHandler** - 處理用戶相關請求
- **WebSocketHandler** - 處理 WebSocket 連接

## 開始使用

### 前置要求

- Go 1.21+
- PostgreSQL
- Redis
- Nacos (可選)

### 安裝

1. 複製代碼庫：

```bash
git clone https://github.com/your-username/passontw-slot-game.git
cd passontw-slot-game
```

2. 安裝依賴：

```bash
go mod download
```

3. 安裝 air 工具 (用於熱重載)：

```bash
go install github.com/cosmtrek/air@latest
```

### 啟動服務

啟動認證服務：

```bash
air auth-service
```

啟動遊戲服務：

```bash
air slot-game1
```

## 環境配置

專案使用 `.env` 文件進行環境配置，各服務使用獨立的環境配置：

- `.env.auth-service` - 認證服務配置
- `.env.slot-game1` - 遊戲服務配置

主要配置項：

```
NACOS_HOST=172.237.27.51
NACOS_PORT=8848
NACOS_NAMESPACE=test_golang
NACOS_GROUP=TEST_GOLANG_ENVS
NACOS_USERNAME=nacos
NACOS_PASSWORD=xup6jo3fup6
NACOS_DATAID=auth_service_config
NACOS_SERVICE_NAME=auth-service1
ENABLE_NACOS=true 
```

## 開發指南

### 添加新的 API 端點

1. 在適當的處理程序文件中添加處理函數：

```go
// 處理程序文件 (src/handler/example_handler.go)
func (h *ExampleHandler) HandleSomething(c *gin.Context) {
    // 處理邏輯
    c.JSON(http.StatusOK, SuccessResponse{Message: "成功"})
}
```

2. 在路由器中註冊端點：

```go
// 路由文件 (src/handler/router.go)
func configurePublicRoutes(api *gin.RouterGroup, ..., exampleHandler *ExampleHandler) {
    // 添加新路由
    api.GET("/examples", exampleHandler.HandleSomething)
}
```

### 添加新的服務

1. 創建服務接口：

```go
// 服務接口 (src/service/example_service.go)
type ExampleService interface {
    DoSomething() error
}
```

2. 實現服務：

```go
// 服務實現 (src/service/example_service.go)
type exampleService struct {
    db *gorm.DB
}

func NewExampleService(db *gorm.DB) ExampleService {
    return &exampleService{db: db}
}

func (s *exampleService) DoSomething() error {
    // 實現邏輯
    return nil
}
```

3. 將服務添加到模組：

```go
// 服務模組 (src/service/module.go)
var Module = fx.Options(
    fx.Provide(
        // 已有服務...
        NewExampleService,
    ),
)
```

## WebSocket 使用

### 客戶端連接 WebSocket

使用標準 WebSocket API 連接到服務器：

```javascript
// 創建 WebSocket 連接
const socket = new WebSocket('ws://localhost:3000/ws');

// 連接成功
socket.onopen = function(event) {
    console.log('WebSocket 連接已開啟');
    
    // 發送認證訊息
    socket.send(JSON.stringify({
        type: 'auth',
        content: 'your-jwt-token'
    }));
};

// 接收訊息
socket.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('收到訊息:', message);
    
    // 處理不同類型的訊息
    switch(message.type) {
        case 'auth_success':
            console.log('認證成功');
            break;
        case 'error':
            console.log('錯誤:', message.content);
            break;
        // 其他訊息類型...
    }
};

// 連接關閉
socket.onclose = function(event) {
    console.log('WebSocket 連接已關閉');
};

// 發送訊息
function sendMessage(type, content) {
    socket.send(JSON.stringify({
        type: type,
        content: content
    }));
}
```

### WebSocket 認證流程

1. 客戶端與 `/ws` 端點建立 WebSocket 連接
2. 客戶端發送 JWT 令牌進行認證：

```json
{
    "type": "auth",
    "content": "your-jwt-token"
}
```

3. 服務器驗證令牌並回應：

```json
{
    "type": "auth_success",
    "content": "Authentication successful"
}
```

或者回應錯誤：

```json
{
    "type": "error",
    "content": "Authentication failed: invalid token"
}
```

4. 認證成功後，客戶端可以發送和接收訊息

### 處理 WebSocket 消息

在服務器端處理 WebSocket 消息的範例：

```go
// 在 WebSocket 管理器的 ReadPump 方法中處理消息
func (client *Client) ReadPump() {
    // 設置連接參數...
    
    for {
        _, message, err := client.Conn.ReadMessage()
        if err != nil {
            // 處理錯誤...
            break
        }
        
        // 解析消息
        var msg Message
        if err := json.Unmarshal(message, &msg); err != nil {
            // 處理錯誤...
            continue
        }
        
        // 根據消息類型處理
        switch msg.Type {
        case "auth":
            // 處理認證...
        case "chat":
            // 處理聊天消息...
        case "game_action":
            // 處理遊戲動作...
        default:
            // 未知消息類型...
        }
    }
}
```

## 依賴注入機制

項目使用 `fx` 框架實現依賴注入，主要優點：

1. **模組化** - 將相關功能分組為模組
2. **生命週期管理** - 處理組件啟動和關閉
3. **依賴管理** - 自動注入所需的依賴項

範例：

```go
// main.go
func main() {
    app := fx.New(
        // 基礎設施
        core.Module,
        
        // 應用模組
        config.Module,
        service.Module,
        handler.Module,
    )
    
    app.Run()
}

// service/module.go
var Module = fx.Options(
    fx.Provide(
        ProvideGormDB,
        NewUserService,
        NewAuthService,
    ),
)

// handler/module.go
var Module = fx.Options(
    fx.Provide(
        NewUserHandler,
        NewAuthHandler,
        NewRouter,
    ),
    fx.Invoke(StartServer),
)
```

## API 文檔

項目使用 Swagger 自動生成 API 文檔，可通過以下 URL 訪問：

- http://localhost:3000/api-docs/index.html (遊戲服務)
- http://localhost:3001/api-docs/index.html (認證服務)

## 貢獻指南

1. Fork 代碼庫
2. 創建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 開啟 Pull Request

## 許可證

此項目根據 [MIT 許可證](LICENSE) 授權。

## 聯繫我們

如有問題或建議，請聯繫：support@passontw.com 