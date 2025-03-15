#!/bin/bash

# 顏色設定
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 印出帶顏色的訊息
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

# 檢測操作系統類型
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        OS_TYPE="macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS_TYPE="linux"
    else
        OS_TYPE="other"
    fi
    print_info "檢測到操作系統: $OS_TYPE"
}

# 檢查必要的命令是否存在
check_requirements() {
    print_step "檢查必要的命令..."
    
    if ! command -v go &> /dev/null; then
        print_error "go 命令不存在，請安裝 Go"
        exit 1
    fi
    
    if ! command -v find &> /dev/null; then
        print_error "find 命令不存在"
        exit 1
    fi
    
    if ! command -v grep &> /dev/null; then
        print_error "grep 命令不存在"
        exit 1
    fi
    
    print_info "所有必要的命令都存在"
}

# 輸入專案名稱
read_project_name() {
    print_step "請輸入專案名稱 (例如: github.com/username/project)："
    read project_name
    
    if [ -z "$project_name" ]; then
        print_error "專案名稱不能為空"
        exit 1
    fi
    
    print_info "專案名稱設定為: $project_name"
}

# 將 template 目錄下的檔案移到上一層
move_template_files() {
    print_step "將 template 目錄下的檔案移動到上一層..."
    
    if [ ! -d "template" ]; then
        print_error "template 目錄不存在"
        exit 1
    fi
    
    # 確保目標目錄存在
    mkdir -p cmd internal pkg deployments
    
    # 移動檔案
    cp -r template/cmd/* cmd/
    cp -r template/internal/* internal/
    cp -r template/pkg/* pkg/
    cp -r template/deployments/* deployments/
    
    # 移動根目錄下其他檔案
    find template -maxdepth 1 -type f -exec cp {} . \;
    
    print_info "檔案移動完成"
}

# 初始化 Go 模組
init_go_module() {
    print_step "初始化 Go 模組為 $project_name..."
    
    go mod init $project_name
    
    if [ $? -ne 0 ]; then
        print_error "Go 模組初始化失敗"
        exit 1
    fi
    
    print_info "Go 模組初始化完成"
}

# 替換所有檔案中的路徑
replace_import_paths() {
    print_step "替換所有檔案中的匯入路徑..."
    
    # 處理所有 .go 文件
    find . -type f -name "*.go" | while read file; do
        if [ "$OS_TYPE" == "macos" ]; then
            # macOS 使用 BSD sed
            sed -i '' "s|github.com/yourusername/project|$project_name|g" "$file"
            sed -i '' "s|github.com/yourusername/shoppingcart|$project_name|g" "$file"
        else
            # Linux 使用 GNU sed
            sed -i "s|github.com/yourusername/project|$project_name|g" "$file"
            sed -i "s|github.com/yourusername/shoppingcart|$project_name|g" "$file"
        fi
        print_info "已處理: $file"
    done
    
    # 處理所有 Dockerfile 文件
    find ./deployments -type f -name "*.Dockerfile" | while read file; do
        if [ "$OS_TYPE" == "macos" ]; then
            sed -i '' "s|github.com/yourusername/project|$project_name|g" "$file"
            sed -i '' "s|github.com/yourusername/shoppingcart|$project_name|g" "$file"
        else
            sed -i "s|github.com/yourusername/project|$project_name|g" "$file"
            sed -i "s|github.com/yourusername/shoppingcart|$project_name|g" "$file"
        fi
        print_info "已處理: $file"
    done
    
    # 處理 docker-compose.yaml
    if [ -f "./deployments/docker/docker-compose.yaml" ]; then
        if [ "$OS_TYPE" == "macos" ]; then
            sed -i '' "s|github.com/yourusername/project|$project_name|g" "./deployments/docker/docker-compose.yaml"
            sed -i '' "s|github.com/yourusername/shoppingcart|$project_name|g" "./deployments/docker/docker-compose.yaml"
        else
            sed -i "s|github.com/yourusername/project|$project_name|g" "./deployments/docker/docker-compose.yaml"
            sed -i "s|github.com/yourusername/shoppingcart|$project_name|g" "./deployments/docker/docker-compose.yaml"
        fi
        print_info "已處理: ./deployments/docker/docker-compose.yaml"
    fi
    
    print_info "匯入路徑替換完成"
}

# 檢查是否有未替換的匯入路徑
check_import_paths() {
    print_step "檢查是否有未替換的匯入路徑..."
    
    # 檢查是否有未替換的 import 路徑
    UNFIXED_IMPORTS=$(grep -r "github.com/yourusername" --include="*.go" . || true)
    
    if [ -n "$UNFIXED_IMPORTS" ]; then
        print_warning "存在未替換的匯入路徑，嘗試再次替換..."
        
        # 再次嘗試替換
        find . -type f -name "*.go" | while read file; do
            if [ "$OS_TYPE" == "macos" ]; then
                sed -i '' "s|github.com/yourusername/project|$project_name|g" "$file"
                sed -i '' "s|github.com/yourusername/shoppingcart|$project_name|g" "$file"
                # 添加更多可能的模式以防萬一
                sed -i '' "s|\"github.com/yourusername|\"$project_name|g" "$file"
            else
                sed -i "s|github.com/yourusername/project|$project_name|g" "$file"
                sed -i "s|github.com/yourusername/shoppingcart|$project_name|g" "$file"
                # 添加更多可能的模式以防萬一
                sed -i "s|\"github.com/yourusername|\"$project_name|g" "$file"
            fi
        done
        
        # 再次檢查
        UNFIXED_IMPORTS=$(grep -r "github.com/yourusername" --include="*.go" . || true)
        if [ -n "$UNFIXED_IMPORTS" ]; then
            print_warning "仍有部分匯入路徑未替換，可能需要手動檢查："
            echo "$UNFIXED_IMPORTS"
        else
            print_info "所有匯入路徑已替換成功"
        fi
    else
        print_info "所有匯入路徑已替換成功"
    fi
}

# 安裝依賴套件
install_dependencies() {
    print_step "安裝依賴套件..."
    
    go mod tidy
    
    if [ $? -ne 0 ]; then
        print_warning "部分依賴套件可能安裝失敗，請手動檢查"
    else
        print_info "依賴套件安裝完成"
    fi
}

# 顯示完成訊息
show_completion() {
    print_step "專案設定完成！"
    print_info "專案名稱: $project_name"
    print_info "專案結構已建立，可以開始開發了"
    print_info "您可以使用以下命令運行 API 網關:"
    print_info "  go run ./cmd/api-gateway/main.go"
    print_info "以及用戶服務:"
    print_info "  go run ./cmd/user-service/main.go"
}

# 清理 template 目錄
cleanup() {
    print_step "清理 template 目錄..."
    
    rm -rf template
    
    print_info "清理完成"
}

# 主程序
main() {
    echo -e "${GREEN}=================================${NC}"
    echo -e "${GREEN}   Go 微服務專案設定嚮導   ${NC}"
    echo -e "${GREEN}=================================${NC}"
    echo ""
    
    detect_os
    check_requirements
    read_project_name
    move_template_files
    init_go_module
    replace_import_paths
    check_import_paths
    install_dependencies
    cleanup
    show_completion
}

# 執行主程序
main