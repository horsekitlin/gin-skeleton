# Shopping Cart Microservices

基於 Golang 的購物車微服務系統，使用 Gin、GORM、WebSocket、Redis 和 PostgreSQL 技術。

## 架構概述

本項目採用微服務架構，包含以下服務：

- **API Gateway**: 處理外部請求，路由到相應的微服務
- **User Service**: 用戶管理和認證
- **Product Service**: 產品管理
- **Shopping Cart Service**: 購物車管理
- **Order Service**: 訂單處理
- **Merchant Service**: 商家管理

## 技術棧

- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **Cache**: Redis
- **Real-time Communication**: WebSocket
- **Service Registration & Discovery**: Nacos
- **Configuration Management**: Nacos
- **Dependency Injection**: Uber FX
- **Container**: Docker
- **CI/CD**: GitHub Actions

## 本地開發

### 先決條件

- Go 1.18+
- Docker & Docker Compose
- PostgreSQL
- Redis

### 設置環境

1. 克隆倉庫:
   ```bash
   git clone https://github.com/yourusername/project.git
   cd shoppingcart
   ```

2. 創建 `.env` 文件:
   ```bash
   cp .env.example .env
   ```
   
3. 修改 `.env` 文件中的設置以匹配您的環境

### 運行服務

使用 Docker Compose 啟動所有服務：

```bash
docker-compose -f deployments/docker/docker-compose.yaml up -d
```

## 部署

本項目使用 GitHub Actions 進行 CI/CD，當推送到 main 分支或創建新標籤時，會自動構建 Docker 映像並部署。

### 所需的 GitHub Secrets

- `DOCKERHUB_USERNAME`: Docker Hub 用戶名
- `DOCKERHUB_TOKEN`: Docker Hub 訪問令牌
- `SERVER_HOST`: 部署服務器主機
- `SERVER_USERNAME`: 部署服務器用戶名
- `SERVER_SSH_KEY`: 部署服務器 SSH 密鑰

### 服務器部署步驟

1. 在服務器上創建部署目錄：
   ```bash
   mkdir -p /path/to/deployment
   ```

2. 複製 `docker-compose.yaml` 和 `.env` 文件到部署目錄：
   ```bash
   cp deployments/docker/docker-compose.yaml /path/to/deployment/
   cp deployments/docker/.env.example /path/to/deployment/.env
   ```

3. 編輯 `.env` 文件，設置正確的環境變量

4. 啟動服務：
   ```bash
   cd /path/to/deployment
   docker-compose up -d
   ```

## 項目結構

```
shoppingcart/
├── api/               # API定義
├── cmd/               # 各服務入口
├── internal/          # 內部實現
├── pkg/               # 公共庫
├── deployments/       # 部署配置
├── scripts/           # 輔助腳本
├── .github/           # GitHub Actions工作流
├── .env               # 環境變量
├── go.mod             # Go模塊定義
└── README.md          # 項目說明
```

## 貢獻

1. Fork 本倉庫
2. 創建您的特性分支：`git checkout -b my-new-feature`
3. 提交您的更改：`git commit -am 'Add some feature'`
4. 推送到分支：`git push origin my-new-feature`
5. 提交拉取請求

## 許可證

[MIT](LICENSE)