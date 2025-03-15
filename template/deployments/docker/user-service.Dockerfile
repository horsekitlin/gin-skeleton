FROM golang:1.19-alpine AS builder

# 安裝必要的工具
RUN apk add --no-cache ca-certificates git

# 設置工作目錄
WORKDIR /app

# 下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製源碼
COPY . .

# 編譯應用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o /app/bin/user-service ./cmd/user-service

# 創建最終映像
FROM alpine:3.17

# 安裝 CA 證書和時區數據
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# 創建非 root 用戶
RUN adduser -D -H -h /app appuser

# 複製編譯好的二進制文件
COPY --from=builder /app/bin/user-service /app/user-service

# 複製 .env 檔案
COPY .env /app/.env

# 創建必要的目錄
RUN mkdir -p /tmp/nacos/log /tmp/nacos/cache && \
    chown -R appuser:appuser /app /tmp/nacos

# 切換到非 root 用戶
USER appuser

# 設置工作目錄
WORKDIR /app

# 設置入口點
ENTRYPOINT ["/app/user-service"]

# 開放端口
EXPOSE 9000

# 健康檢查 (gRPC 服務可能需要不同的健康檢查方式)
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD nc -z localhost 9000 || exit 1