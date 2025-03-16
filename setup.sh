#!/bin/bash

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_NAME=""
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
TEMPLATE_DIR="${SCRIPT_DIR}/template"

print_step() {
    echo -e "${BLUE}==>${NC} ${GREEN}$1${NC}"
}

print_info() {
    echo -e "${BLUE}  ->${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

print_error() {
    echo -e "${RED}Error:${NC} $1"
}

print_step "檢查模板目錄..."
if [ ! -d "$TEMPLATE_DIR" ]; then
    print_error "模板目錄不存在: $TEMPLATE_DIR"
    exit 1
fi

print_step "設定專案名稱..."
echo -n "請輸入專案名稱 (例如: github.com/username/project)："
read PROJECT_NAME

if [ -z "$PROJECT_NAME" ]; then
    print_error "專案名稱不能為空"
    exit 1
fi

print_info "專案名稱設定為: $PROJECT_NAME"

print_step "複製檔案..."
find "$TEMPLATE_DIR" -type f -not -path "*/.git*" | while read file; do
    rel_path="${file#$TEMPLATE_DIR/}"
    target_dir="$(dirname "$rel_path")"
    
    if [ ! -d "$target_dir" ]; then
        mkdir -p "$target_dir"
    fi
    
    if [[ "$rel_path" == "README.md" && -f "README.md" ]]; then
        print_info "  已跳過: $rel_path (避免覆蓋)"
    else
        cp "$file" "$rel_path"
        print_info "  已複製: $rel_path"
    fi
done

print_step "替換匯入路徑..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    find . -type f -name "*.go" -not -path "./template/*" | while read file; do
        sed -i '' "s|github.com/yourusername/project|$PROJECT_NAME|g" "$file"
        sed -i '' "s|github.com/yourusername/shoppingcart|$PROJECT_NAME|g" "$file"
        sed -i '' "s|\"github.com/yourusername|\"$PROJECT_NAME|g" "$file"
    done
else
    find . -type f -name "*.go" -not -path "./template/*" | while read file; do
        sed -i "s|github.com/yourusername/project|$PROJECT_NAME|g" "$file"
        sed -i "s|github.com/yourusername/shoppingcart|$PROJECT_NAME|g" "$file"
        sed -i "s|\"github.com/yourusername|\"$PROJECT_NAME|g" "$file"
    done
fi

print_step "初始化 Go 模組..."
if [ -f go.mod ]; then
    rm go.mod
fi
go mod init $PROJECT_NAME

print_step "安裝依賴套件..."
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
go mod tidy

print_step "建立 docs 目錄..."
mkdir -p docs
cat > docs/docs.go << EOF
package docs

import "github.com/swaggo/swag"

var SwaggerInfo = &swag.Spec{
	Version: "1.0",
	Host: "localhost:8080",
	BasePath: "/api/v1",
	Schemes: []string{"http", "https"},
	Title: "$PROJECT_NAME API",
	Description: "API for $PROJECT_NAME",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
EOF

print_step "清理重複檔案..."
find . -name "*''" -delete

print_step "清理模板..."
read -p "是否刪除 template 目錄和 setup.sh？(y/n): " answer
if [[ "$answer" == "y" || "$answer" == "Y" ]]; then
    rm -rf "$TEMPLATE_DIR"
    rm "$SCRIPT_DIR/setup.sh"
    print_info "清理完成"
fi

print_step "專案設定完成！"
print_info "專案名稱: $PROJECT_NAME"
print_info "您可以使用命令 'go run ./main.go' 運行專案"