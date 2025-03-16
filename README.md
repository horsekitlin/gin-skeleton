# Go 微服務專案框架

這個專案是一個基於 Go 語言的微服務框架，採用了現代化的架構設計和主流的依賴套件，讓開發者可以快速搭建高性能、可擴展的微服務。

## 專案特色

- **模組化架構**：使用 `go.uber.org/fx` 實現依賴注入和模組化設計
- **RESTful API**：基於 Gin 框架實現的 REST API
- **資料庫整合**：支援 PostgreSQL 和自動化資料庫遷移
- **服務發現**：集成 Nacos 服務註冊與發現
- **文件生成**：自動生成 Swagger API 文件
- **認證功能**：內建 JWT 認證方式
- **快取支援**：整合 Redis 快取
- **WebSocket**：支援實時通訊
- **結構化日誌**：使用 zap 提供高效能的日誌記錄

## 環境需求

- Go 1.18+
- PostgreSQL
- Redis (選用)
- Nacos (選用)

## 快速開始

### 使用 setup.sh 初始化專案

1. 複製專案模板

```bash
git clone https://github.com/yourusername/go-microservice-template.git my-service
cd my-service
```

2. 執行初始化腳本

```bash
chmod +x setup.sh
./setup.sh
```

3. 輸入你的專案名稱（如：github.com/yourusername/my-service）

4. 設定完成後，你的專案就準備好了

### 手動配置

1. 修改 `.env` 檔案，根據你的環境設定資料庫、Redis 等配置

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=your_db_name
REDIS_ADDR=localhost:6379
```

2. 執行遷移腳本初始化資料庫

```bash
# 請確保你已經創建了相應的資料庫
psql -U postgres -c "CREATE DATABASE your_db_name;"

# 執行初始化 SQL
psql -U postgres -d your_db_name -f migrations/games-users.sql
psql -U postgres -d your_db_name -f migrations/games-orders.sql
```

3. 運行專案

```bash
go run main.go
```

## 專案結構說明

```
.
├── internal/              # 內部代碼，不對外公開
│   ├── config/            # 配置相關
│   │   ├── init.go        # 配置初始化
│   │   ├── models.go      # 配置模型
│   │   ├── nacos.go       # Nacos 配置
│   │   └── provider.go    # 配置提供者
│   ├── domain/            # 領域模型
│   │   ├── entity/        # 實體定義
│   │   ├── models/        # 數據模型
│   │   └── symbol.go      # 符號定義
│   ├── handler/           # HTTP 處理器
│   ├── interfaces/        # 接口定義
│   ├── middleware/        # 中間件
│   └── service/           # 業務邏輯
├── migrations/            # 資料庫遷移腳本
│   └── seeds/             # 種子數據
├── pkg/                   # 公共包
│   ├── core/              # 核心模組
│   ├── databaseManager/   # 資料庫管理
│   ├── httpClient/        # HTTP 客戶端
│   ├── logger/            # 日誌工具
│   ├── nacosManager/      # Nacos 管理器
│   ├── redisManager/      # Redis 管理器
│   ├── utils/             # 通用工具
│   └── websocketManager/  # WebSocket 管理器
├── docs/                  # Swagger 文檔
├── .env                   # 環境變數
├── .env.example           # 環境變數範例
├── go.mod                 # Go 模組定義
├── go.sum                 # Go 模組版本鎖定
└── main.go                # 主程式入口
```

## 模塊說明

### 配置管理 (internal/config)

配置管理模組負責讀取和管理應用的配置信息，支援從環境變數、配置文件和 Nacos 讀取配置。

#### 使用方法

```go
import "github.com/yourusername/project/internal/config"

func UseConfig(cfg *config.Config) {
    // 使用配置
    dbConfig := cfg.Database
    redisConfig := cfg.Redis
}
```

### 服務層 (internal/service)

服務層包含業務邏輯實現，每個服務都實現了對應的業務功能。

#### 新增服務

1. 在 `internal/interfaces` 添加接口定義

```go
package interfaces

type NewService interface {
    DoSomething() error
}
```

2. 在 `internal/service` 實現接口

```go
package service

type NewServiceImpl struct {
    // 依賴項
    db *gorm.DB
}

func NewNewService(db *gorm.DB) interfaces.NewService {
    return &NewServiceImpl{
        db: db,
    }
}

func (s *NewServiceImpl) DoSomething() error {
    // 實現邏輯
    return nil
}
```

3. 在 `internal/service/module.go` 註冊服務

```go
func Module() fx.Option {
    return fx.Options(
        // 其他服務
        fx.Provide(NewNewService),
    )
}
```

### 資料庫管理

專案使用 GORM 作為 ORM 框架，支援 PostgreSQL。

#### 定義模型

```go
package entity

import (
    "time"
    "github.com/yourusername/project/pkg/utils"
)

type User struct {
    ID        string    `gorm:"primaryKey" json:"id"`
    Username  string    `gorm:"uniqueIndex" json:"username"`
    Email     string    `gorm:"uniqueIndex" json:"email"`
    Password  string    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate() error {
    u.ID = utils.GenerateID()
    return nil
}
```

#### 資料庫操作

```go
func (s *UserServiceImpl) CreateUser(params CreateUserParams) (*entity.User, error) {
    user := &entity.User{
        Username: params.Username,
        Email:    params.Email,
        Password: params.Password,
    }
    
    if err := s.db.Create(user).Error; err != nil {
        return nil, err
    }
    
    return user, nil
}
```

### HTTP 處理器 (internal/handler)

HTTP 處理器負責處理 HTTP 請求並返回響應。

#### 新增處理器

1. 在 `internal/handler` 添加處理器文件

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/project/internal/interfaces"
)

type NewHandler struct {
    newService interfaces.NewService
}

func NewNewHandler(newService interfaces.NewService) *NewHandler {
    return &NewHandler{
        newService: newService,
    }
}

func (h *NewHandler) Handle(c *gin.Context) {
    // 處理請求
    err := h.newService.DoSomething()
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"message": "success"})
}
```

2. 在 `internal/handler/module.go` 註冊處理器

```go
func Module() fx.Option {
    return fx.Options(
        // 其他處理器
        fx.Provide(NewNewHandler),
    )
}
```

3. 在 `internal/handler/router.go` 註冊路由

```go
func RegisterRoutes(r *gin.Engine, newHandler *NewHandler) {
    api := r.Group("/api/v1")
    {
        // 其他路由
        api.POST("/new", newHandler.Handle)
    }
}
```

### 認證中間件

專案使用 JWT 進行認證。

#### 使用認證中間件

```go
func RegisterRoutes(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
    api := r.Group("/api/v1")
    {
        // 公開路由
        api.POST("/login", authHandler.Login)
        
        // 需要認證的路由
        authenticated := api.Group("").Use(authMiddleware.Authenticate())
        {
            authenticated.GET("/profile", userHandler.GetProfile)
        }
    }
}
```

### 使用 Redis 缓存

專案支援 Redis 快取，適合在需要高性能讀取的場景使用。

```go
import "github.com/yourusername/project/pkg/redisManager"

func (s *AuthServiceImpl) GetUserToken(userId string) (string, error) {
    // 嘗試從 Redis 獲取
    token, err := s.redis.Get(context.Background(), "token:"+userId)
    if err == nil && token != "" {
        return token, nil
    }
    
    // 生成新 token
    token, err = s.generateToken(userId)
    if err != nil {
        return "", err
    }
    
    // 保存到 Redis
    s.redis.Set(context.Background(), "token:"+userId, token, time.Hour*24)
    
    return token, nil
}
```

### WebSocket 支援

專案支援 WebSocket 實時通訊。

```go
import "github.com/yourusername/project/pkg/websocketManager"

func (h *MessageHandler) HandleWebSocket(c *gin.Context) {
    // 升級 HTTP 連接為 WebSocket
    conn, err := h.wsManager.Upgrade(c.Writer, c.Request)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // 註冊客戶端
    client := &websocketManager.Client{
        ID:   c.GetString("userId"),
        Conn: conn,
        Send: make(chan []byte, 256),
    }
    
    h.wsManager.Register(client)
    
    // 處理連接
    go client.WritePump()
    go client.ReadPump()
}
```

## API 文檔

專案使用 Swagger 生成 API 文檔。

### 生成 Swagger 文檔

```bash
swag init -g main.go
```

### 訪問 Swagger UI

啟動服務後，訪問 `http://localhost:8080/swagger/index.html` 查看 API 文檔。

## 環境變數設定

專案使用 .env 文件和環境變數配置。以下是主要的配置項：

| 變數名 | 說明 | 預設值 |
|--------|------|--------|
| SERVER_PORT | 服務器監聽端口 | 8080 |
| DB_HOST | 資料庫主機 | localhost |
| DB_PORT | 資料庫端口 | 5432 |
| DB_USER | 資料庫用戶名 | postgres |
| DB_PASSWORD | 資料庫密碼 | |
| DB_NAME | 資料庫名稱 | |
| REDIS_ADDR | Redis 地址 | localhost:6379 |
| REDIS_PASSWORD | Redis 密碼 | |
| JWT_SECRET | JWT 簽名密鑰 | |
| NACOS_ADDR | Nacos 服務地址 | |
| LOG_LEVEL | 日誌級別 | info |

## 部署指南

### Docker 部署

1. 建立 Docker 映像

```bash
docker build -t your-service-name .
```

2. 運行容器

```bash
docker run -p 8080:8080 --env-file .env your-service-name
```

### Kubernetes 部署

提供 Kubernetes 部署配置示例：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: your-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: your-service
  template:
    metadata:
      labels:
        app: your-service
    spec:
      containers:
      - name: your-service
        image: your-service-name:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: your-service-config
---
apiVersion: v1
kind: Service
metadata:
  name: your-service
spec:
  selector:
    app: your-service
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

## 常見問題排解

### 無法連接資料庫

- 檢查 `.env` 文件中的資料庫配置是否正確
- 確認資料庫服務是否正在運行
- 檢查網絡連接和防火牆設定

### Swagger 文檔生成失敗

- 確認已安裝 swag 工具：`go install github.com/swaggo/swag/cmd/swag@latest`
- 檢查註解格式是否正確
- 運行 `swag init -g main.go` 重新生成

### 依賴問題

如果遇到依賴相關問題，嘗試以下步驟：

```bash
go mod tidy
go clean -modcache
go mod download
```

## 性能優化

### 資料庫效能優化

- 使用適當的索引
- 優化查詢語句
- 考慮分頁查詢大量數據

### Redis 快取策略

- 對熱點數據使用 Redis 快取
- 設定合理的過期時間
- 使用 Redis Pipeline 減少網絡往返

### 日誌級別控制

在生產環境中，建議將日誌級別設為 `info` 或 `warn`，以減少日誌量並提高性能：

```env
LOG_LEVEL=info
```

## 開發指南

### 目錄結構規範

- `internal/` - 內部程式碼，不對外公開
- `pkg/` - 公共程式碼，可被外部專案引用
- `migrations/` - 資料庫遷移腳本
- `docs/` - 文檔和 Swagger 檔案

### 編碼規範

- 遵循 Go 官方規範 https://golang.org/doc/effective_go
- 使用 gofmt 格式化代碼
- 添加適當的註釋
- 將業務邏輯放在 service 層，避免在 handler 層實現

### 單元測試

為主要功能編寫單元測試，特別是業務邏輯和工具函數：

```bash
go test ./...
```

## 貢獻指南

1. Fork 專案
2. 創建功能分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add some amazing feature'`
4. 推送到分支：`git push origin feature/amazing-feature`
5. 創建 Pull Request

## 參考資源

- [Gin 框架文檔](https://gin-gonic.com/docs/)
- [GORM 文檔](https://gorm.io/docs/)
- [Uber FX 文檔](https://uber-go.github.io/fx/)
- [Swagger 註解指南](https://github.com/swaggo/swag#declarative-comments-format)
- [Nacos 使用指南](https://nacos.io/en-us/docs/quick-start.html)

## 授權協議

[MIT License](https://opensource.org/licenses/MIT)