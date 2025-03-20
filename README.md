# Golang Monorepo 項目結構說明

## 目錄
- [簡介](#簡介)
- [整體結構](#整體結構)
- [資料夾說明](#資料夾說明)
- [模組管理](#模組管理)
- [規範與最佳實踐](#規範與最佳實踐)
- [工具配置](#工具配置)
- [開發流程](#開發流程)
- [常見問題](#常見問題)

## 簡介

本 monorepo 採用集中式儲存庫管理方法，使用 Golang 作為主要開發語言，將所有相關專案、套件和應用程式集中在單一代碼儲存庫中。使用 monorepo 架構的主要目標是促進代碼共享、簡化依賴管理並確保整個生態系統的一致性。

### 什麼是 Monorepo?

Monorepo（單體儲存庫）是一種將多個專案的代碼儲存在同一個版本控制儲存庫中的開發策略。這與傳統的多儲存庫（multi-repo）方法相反，後者為每個專案維護單獨的儲存庫。

### 我們使用 Monorepo 的原因

- **代碼共享和重用**：更容易共享通用功能和套件
- **統一版本控制**：所有專案共享同一個版本歷史
- **簡化依賴管理**：避免複雜的外部依賴關係
- **原子性變更**：可以跨多個專案進行原子性變更
- **統一的工具與流程**：所有專案使用相同的構建工具、測試框架和開發流程

## 整體結構

```
monorepo-root/
├── .github/                  # GitHub 相關配置
├── .golangci.yml             # Golangci-lint 配置
├── Makefile                  # 全域 Makefile
├── go.mod                    # 主要 Go 模組定義
├── go.sum                    # 相依性校驗檔案
├── cmd/                      # 應用程式入口點
│   ├── api/                  # API 服務入口
│   ├── worker/               # 背景工作處理器入口
│   ├── cli/                  # 命令列工具入口
│   └── admin/                # 管理工具入口
├── internal/                 # 僅供本倉庫使用的程式碼
│   ├── app/                  # 應用邏輯
│   │   ├── api/              # API 服務
│   │   ├── worker/           # 背景工作處理器
│   │   └── admin/            # 管理應用
│   ├── auth/                 # 認證相關
│   ├── config/               # 配置管理
│   ├── db/                   # 資料庫連接和模型
│   ├── middleware/           # HTTP 中間件
│   └── server/               # 服務器設定
├── pkg/                      # 可被外部專案匯入的套件
│   ├── logger/               # 日誌套件
│   ├── validator/            # 資料驗證
│   ├── errors/               # 錯誤處理
│   └── utils/                # 通用工具函數
├── api/                      # API 定義
│   ├── proto/                # Protocol Buffers 定義
│   ├── swagger/              # Swagger/OpenAPI 規範
│   └── graphql/              # GraphQL 結構定義
├── web/                      # Web 前端應用
│   ├── admin/                # 管理界面
│   └── client/               # 用户界面
├── deployments/              # 部署配置
│   ├── docker/               # Docker 配置
│   ├── kubernetes/           # Kubernetes 配置
│   └── terraform/            # 基礎設施即代碼
├── migrations/               # 資料庫遷移檔案
├── scripts/                  # 開發和運維腳本
│   ├── db/                   # 資料庫相關腳本
│   ├── ci/                   # CI 腳本
│   └── dev/                  # 開發輔助腳本
├── docs/                     # 文檔
│   ├── architecture/         # 架構文檔
│   ├── guides/               # 開發指南
│   └── api/                  # API 文檔
├── tools/                    # 開發工具
│   ├── generators/           # 代碼生成器
│   └── build/                # 構建工具
└── README.md                 # 項目說明
```

## 資料夾說明

### .github/

**用途**：存放 GitHub 相關配置文件。
**內容**：
- `workflows/`: GitHub Actions 工作流程定義
- `ISSUE_TEMPLATE/`: Issue 模板
- `PULL_REQUEST_TEMPLATE.md`: PR 模板

**規範**：
- 所有 CI/CD 流程應定義在 `workflows/` 中
- 保持工作流程模組化，避免過於複雜的單一工作流程
- 使用標準的 issue 和 PR 模板以確保提交質量

### cmd/

**用途**：包含各個應用程式的入口點 (main 函數)。
**內容**：

#### api/
**用途**：API 服務的入口點。
**結構**：
```
api/
├── main.go                   # 主函數
├── app.go                    # 應用設定與初始化
└── flags.go                  # 命令列參數定義
```

#### worker/
**用途**：背景工作處理器的入口點。
**結構**：
```
worker/
├── main.go                   # 主函數
├── app.go                    # 應用設定與初始化
└── flags.go                  # 命令列參數定義
```

#### cli/
**用途**：命令列工具的入口點。
**結構**：
```
cli/
├── main.go                   # 主函數
├── commands/                 # 命令定義
└── flags.go                  # 命令列參數定義
```

**規範**：
- 每個入口點應盡量保持精簡，主要負責初始化和啟動應用
- 業務邏輯應放在 `internal/app/` 或 `pkg/` 中
- 使用 `flag` 或 `cobra` 處理命令列參數
- 遵循 [Go 命令模式](https://github.com/golang-standards/project-layout/tree/master/cmd)

### internal/

**用途**：僅供本倉庫使用的程式碼，無法被外部專案匯入。
**內容**：

#### app/
**用途**：應用邏輯，按照不同應用類型組織。
**結構**：
```
app/
├── api/                      # API 服務邏輯
│   ├── handlers/             # HTTP 處理器
│   ├── routes/               # 路由定義
│   └── services/             # 業務服務
├── worker/                   # 背景工作處理器邏輯
│   ├── jobs/                 # 工作定義
│   ├── queue/                # 佇列管理
│   └── scheduler/            # 排程器
└── admin/                    # 管理應用邏輯
    ├── handlers/             # HTTP 處理器
    ├── routes/               # 路由定義
    └── services/             # 業務服務
```

#### auth/
**用途**：認證和授權相關邏輯。
**結構**：
```
auth/
├── jwt/                      # JWT 處理
├── oauth/                    # OAuth 處理
└── rbac/                     # 基於角色的訪問控制
```

#### config/
**用途**：配置管理。
**結構**：
```
config/
├── loader.go                 # 配置載入器
├── validator.go              # 配置驗證
└── defaults.go               # 默認配置
```

#### db/
**用途**：資料庫連接和模型。
**結構**：
```
db/
├── connection.go             # 資料庫連接管理
├── models/                   # 資料模型定義
├── migrations/               # 資料庫遷移管理
└── repositories/             # 資料庫操作封裝
```

**規範**：
- 使用 `internal` 套件保護不應被外部使用的代碼
- 按照功能和領域組織代碼，而非技術層次
- 使用依賴注入模式管理服務間依賴
- 遵循介面隔離原則，定義清晰的接口

### pkg/

**用途**：可被外部專案匯入的套件，應具有穩定的 API。
**內容**：

#### logger/
**用途**：日誌處理套件。
**結構**：
```
logger/
├── logger.go                 # 日誌介面定義
├── zap.go                    # Zap 實現
└── logrus.go                 # Logrus 實現
```

#### validator/
**用途**：資料驗證套件。
**結構**：
```
validator/
├── validator.go              # 驗證器介面
├── rules/                    # 驗證規則
└── errors.go                 # 驗證錯誤
```

#### errors/
**用途**：錯誤處理套件。
**結構**：
```
errors/
├── errors.go                 # 錯誤類型定義
├── codes.go                  # 錯誤碼定義
└── handler.go                # 錯誤處理器
```

**規範**：
- 每個套件應有明確的責任和範圍
- 提供完整的文檔和單元測試
- 遵循 Go 套件設計最佳實踐
- 版本控制遵循語義化版本規範
- 使用介面實現依賴倒置

### api/

**用途**：API 定義和協議。
**內容**：

#### proto/
**用途**：Protocol Buffers 定義檔案，用於 gRPC 服務。
**結構**：
```
proto/
├── service.proto             # 服務定義
├── models.proto              # 資料模型定義
└── generated/                # 生成的代碼
```

#### swagger/
**用途**：Swagger/OpenAPI 規範，用於 REST API 文檔。
**結構**：
```
swagger/
├── api.yaml                  # API 規範
└── generated/                # 生成的文檔
```

**規範**：
- 使用版本控制管理 API 變更
- 先定義 API，再實現功能
- 遵循 API 設計最佳實踐
- 使用自動化工具生成文檔和客戶端

### web/

**用途**：Web 前端應用（如使用 React、Vue 等）。
**內容**：

#### admin/
**用途**：管理界面前端。
**結構**：
```
admin/
├── src/                      # 源代碼
├── public/                   # 靜態資源
├── package.json              # 套件配置
└── README.md                 # 專案說明
```

#### client/
**用途**：用户界面前端。
**結構**：
```
client/
├── src/                      # 源代碼
├── public/                   # 靜態資源
├── package.json              # 套件配置
└── README.md                 # 專案說明
```

**規範**：
- 前端代碼遵循各自框架的最佳實踐
- 使用 API 客戶端與後端通信
- 遵循一致的代碼風格
- 考慮使用靜態類型（如 TypeScript）

### deployments/

**用途**：部署相關配置和腳本。
**內容**：

#### docker/
**用途**：Docker 配置檔案。
**結構**：
```
docker/
├── Dockerfile.api            # API 服務的 Dockerfile
├── Dockerfile.worker         # Worker 的 Dockerfile
└── docker-compose.yml        # 開發用 Docker Compose 配置
```

#### kubernetes/
**用途**：Kubernetes 部署配置。
**結構**：
```
kubernetes/
├── api/                      # API 服務部署
├── worker/                   # Worker 部署
└── common/                   # 共享資源
```

**規範**：
- 按環境（開發、測試、生產）組織配置
- 使用環境變數進行配置注入
- 遵循基礎設施即代碼原則
- 使用 Helm 或類似工具管理複雜配置

### migrations/

**用途**：資料庫遷移檔案。
**結構**：
```
migrations/
├── 001_initial_schema.sql    # 初始架構
├── 002_add_users_table.sql   # 增加使用者表
└── README.md                 # 遷移說明
```

**規範**：
- 遷移檔案應按序號排列
- 每個遷移應專注於一個變更
- 每個遷移應包含向前和向後的操作
- 遷移應是冪等的

### scripts/

**用途**：開發和運維腳本。
**內容**：

#### db/
**用途**：資料庫相關腳本。
**結構**：
```
db/
├── seed.go                   # 資料填充
├── backup.sh                 # 資料庫備份
└── restore.sh                # 資料庫恢復
```

#### ci/
**用途**：CI 腳本。
**結構**：
```
ci/
├── build.sh                  # 構建腳本
├── test.sh                   # 測試腳本
└── deploy.sh                 # 部署腳本
```

**規範**：
- 腳本應有明確的文檔說明
- 支持不同的執行環境
- 錯誤處理和日誌輸出
- 參數化配置

### docs/

**用途**：項目文檔。
**內容**：

#### architecture/
**用途**：架構文檔。
**結構**：
```
architecture/
├── overview.md               # 系統概覽
├── diagrams/                 # 架構圖
└── decisions/                # 架構決策記錄 (ADRs)
```

#### guides/
**用途**：開發指南。
**結構**：
```
guides/
├── getting-started.md        # 入門指南
├── contribution.md           # 貢獻指南
└── style-guide.md            # 代碼風格指南
```

**規範**：
- 使用 Markdown 格式
- 文檔應保持更新
- 包含程式碼範例
- 使用圖表輔助說明

### tools/

**用途**：開發工具。
**內容**：

#### generators/
**用途**：代碼生成器。
**結構**：
```
generators/
├── templates/                # 模板文件
├── model.go                  # 模型生成器
└── handler.go                # 處理器生成器
```

**規範**：
- 工具應簡化重複性工作
- 提供完整的使用說明
- 注重可擴展性
- 與專案代碼風格一致

## 模組管理

Go 模組系統是 Golang 的官方依賴管理方案。在 monorepo 中，我們使用單一的主要 `go.mod` 文件來管理依賴。

### go.mod 結構

```
module github.com/organization/monorepo

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/go-sql-driver/mysql v1.7.1
    // 其他依賴
)
```

### 匯入路徑

在本 monorepo 中，匯入路徑應遵循以下規範：

- 內部套件匯入：
  ```go
  import "github.com/organization/monorepo/internal/app/api"
  ```

- 公開套件匯入：
  ```go
  import "github.com/organization/monorepo/pkg/logger"
  ```

### 版本控制

雖然 monorepo 內部套件共享同一個版本，但我們仍然需要為整個倉庫設置版本標籤，特別是為了公開的 `pkg` 套件：

```bash
git tag v1.0.0
git push origin v1.0.0
```

## 規範與最佳實踐

### 代碼風格

- 遵循 [Effective Go](https://golang.org/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 或 `goimports` 格式化代碼
- 使用 `golangci-lint` 進行靜態分析
- 遵循 Go 項目標準布局

### 版本控制

- 遵循 [Conventional Commits](https://www.conventionalcommits.org/) 提交規範
- 使用語義化版本控制（[SemVer](https://semver.org/)）
- 使用 feature 分支進行功能開發
- 使用 pull request 進行代碼審查

### 錯誤處理

- 使用明確的錯誤類型和錯誤碼
- 遵循 Go 的錯誤處理慣例，使用多返回值
- 避免過度使用 `panic`
- 使用 `context` 進行請求範圍控制和取消

### 測試策略

- 使用 Go 的內建測試框架
- 編寫單元測試、集成測試和基準測試
- 使用 testify 等工具簡化測試
- 使用 `go test -race` 檢測競態條件
- 使用模擬（mock）隔離外部依賴

### 日誌和監控

- 使用結構化日誌
- 按照嚴重性級別分類日誌
- 實現分散式追蹤
- 設置健康檢查和度量指標端點

## 工具配置

### 開發工具

我們使用以下工具進行開發：

- **Go 1.21+**：主要開發語言
- **golangci-lint**：代碼質量檢查
- **goimports**：代碼格式化和匯入管理
- **gomock**：模擬生成
- **Docker**：容器化環境
- **Swagger**：API 文檔
- **Goose/Migrate**：資料庫遷移
- **dlv**：Go 語言調試器

### 編輯器配置

建議使用以下編輯器插件：

- **VSCode Go 插件**
- **GoLand IDE**
- **Vim-go**

### Makefile 命令

我們使用 Makefile 簡化常用命令：

```makefile
.PHONY: build test lint run migrate

build:
    go build -o bin/ ./cmd/...

test:
    go test -race ./...

lint:
    golangci-lint run

run:
    go run ./cmd/api/main.go

migrate:
    go run ./cmd/migration/main.go up
```

## 開發流程

### 設置開發環境

```bash
# 克隆儲存庫
git clone https://github.com/organization/monorepo.git
cd monorepo

# 安裝依賴
go mod download

# 安裝開發工具
make setup-dev

# 啟動開發環境
docker-compose up -d
```

### 常用命令

```bash
# 運行所有測試
make test

# 執行特定套件的測試
go test ./pkg/logger

# 構建所有應用
make build

# 運行特定應用
make run-api

# 執行 lint 檢查
make lint

# 建立新的遷移
make migration name=add_users_table

# 生成 API 文檔
make swagger
```

### 分支策略

- `main`：主要開發分支，所有功能開發完成後合併至此
- `production`：生產環境分支，僅接受從 `main` 合併的穩定版本
- `feature/*`：功能開發分支
- `bugfix/*`：錯誤修復分支
- `release/*`：發布準備分支

### 發布流程

1. 更新版本號和變更日誌
2. 創建發布分支 `release/vX.Y.Z`
3. 執行完整測試和品質檢查
4. 合併至 `main` 和 `production` 分支
5. 創建版本標籤
6. 構建和發布二進制文件或容器映像

## 常見問題

### Q: 如何在 monorepo 中管理多個應用程式？

A: 我們將每個應用程式的入口點放在 `cmd/` 目錄下，並將業務邏輯放在 `internal/app/` 中相應的子目錄下。共享代碼放在 `pkg/` 或 `internal/` 的其他目錄中。使用 Makefile 或其他構建工具簡化多應用程式的構建和運行。

### Q: 如何處理跨應用程式的共享代碼？

A: 共享代碼應放在 `pkg/` 目錄（如果可以被外部專案使用）或 `internal/` 目錄（如果僅供本倉庫使用）中。通過清晰的套件組織和接口設計，確保共享代碼的一致性和可測試性。

### Q: 如何管理不同環境的配置？

A: 使用環境變數和配置文件的組合。將配置結構定義在 `internal/config/` 中，並使用 `-config` 命令列參數指定配置文件路徑。敏感配置（如密碼、API 金鑰）應使用環境變數注入，而不是儲存在代碼中。

### Q: 如何確保代碼質量？

A: 使用以下策略：
- 強制代碼審查
- 自動化測試（單元測試、集成測試、端到端測試）
- 使用 `golangci-lint` 進行靜態分析
- 使用 CI/CD 管道自動執行測試和檢查
- 定期代碼重構和改進

### Q: 如何處理資料庫遷移？

A: 使用 migration 工具（如 Goose、golang-migrate）管理資料庫遷移。遷移檔案儲存在 `migrations/` 目錄中，並按照順序執行。在 CI/CD 管道中自動應用遷移，確保資料庫結構與代碼同步。

## 結論

本 Golang monorepo 結構旨在促進代碼共享、簡化依賴管理並確保一致性。通過遵循這些指南和最佳實踐，我們可以有效地管理複雜的 Go 專案並提高開發效率。

如有問題或建議，請參考文檔或聯繫架構團隊。
