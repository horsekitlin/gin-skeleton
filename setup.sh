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
    for cmd in go find grep; do
        if ! command -v $cmd &> /dev/null; then
            print_error "$cmd 命令不存在"
            exit 1
        fi
    done
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

# 複製模板檔案
copy_template_files() {
    print_step "複製模板檔案..."
    
    if [ ! -d "template" ]; then
        print_error "template 目錄不存在"
        exit 1
    fi
    
    find template -type d -not -path "template" -not -path "template/.git*" | while read dir; do
        target_dir="${dir#template/}"
        if [ ! -z "$target_dir" ]; then
            mkdir -p "$target_dir"
        fi
    done
    
    find template -type f -not -path "template/.git*" | while read file; do
        target_file="${file#template/}"
        if [ ! -z "$target_file" ]; then
            cp "$file" "$target_file"
        else
            cp "$file" ./
        fi
    done
    
    print_info "檔案複製完成"
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
    
    local sed_cmd
    if [ "$OS_TYPE" == "macos" ]; then
        sed_cmd="sed -i ''"
    else
        sed_cmd="sed -i"
    fi
    
    local patterns=(
        "github.com/yourusername/project"
        "github.com/yourusername/shoppingcart"
        "\"github.com/yourusername"
    )
    
    local replacements=(
        "$project_name"
        "$project_name"
        "\"$project_name"
    )
    
    local file_types=(
        "*.go"
        "*.Dockerfile"
        "docker-compose.yaml"
        "*.yml"
        "*.yaml"
    )
    
    for i in "${!patterns[@]}"; do
        for ext in "${file_types[@]}"; do
            find . -type f -name "$ext" -not -path "./template/*" | while read file; do
                $sed_cmd "s|${patterns[$i]}|${replacements[$i]}|g" "$file"
            done
        done
    done
    
    print_info "匯入路徑替換完成"
}

# 檢查是否有未替換的匯入路徑
check_import_paths() {
    print_step "檢查是否有未替換的匯入路徑..."
    
    UNFIXED_IMPORTS=$(grep -r "github.com/yourusername" --include="*.go" . --exclude-dir=template || true)
    
    if [ -n "$UNFIXED_IMPORTS" ]; then
        print_warning "仍有部分匯入路徑未替換，可能需要手動檢查："
        echo "$UNFIXED_IMPORTS"
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
    print_info "您可以使用命令 'go run ./main.go' 運行專案"
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
    copy_template_files
    init_go_module
    replace_import_paths
    check_import_paths
    install_dependencies
    cleanup
    show_completion
}

# 執行主程序
main