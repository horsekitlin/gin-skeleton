#!/bin/bash

# 顏色設定
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 全域變數
PROJECT_NAME=""
OS_TYPE=""
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
TEMPLATE_DIR="${SCRIPT_DIR}/template"
TMP_DIR="${SCRIPT_DIR}/.tmp_template"

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
    print_step "檢測操作系統類型..."
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
    local required_cmds=("go" "find" "grep" "sed")
    local missing=0
    
    for cmd in "${required_cmds[@]}"; do
        if ! command -v $cmd &> /dev/null; then
            print_error "$cmd 命令不存在"
            missing=1
        fi
    done
    
    if [ $missing -eq 1 ]; then
        print_error "請先安裝缺少的命令，然後再試一次"
        exit 1
    fi
    
    print_info "所有必要的命令都存在"
}

# 清理函數 - 用於退出前清理臨時文件
cleanup_on_exit() {
    if [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}

# 捕捉退出信號，確保清理臨時文件
trap cleanup_on_exit EXIT

# 驗證 template 目錄是否存在
check_template_dir() {
    print_step "檢查模板目錄..."
    
    if [ ! -d "$TEMPLATE_DIR" ]; then
        print_error "模板目錄不存在: $TEMPLATE_DIR"
        print_info "請確認您在正確的目錄下執行此腳本"
        print_info "目前腳本位置: $SCRIPT_DIR"
        exit 1
    fi
    
    print_info "模板目錄存在: $TEMPLATE_DIR"
    print_info "模板目錄內容:"
    ls -la "$TEMPLATE_DIR" | head -10
    if [ "$(ls -A "$TEMPLATE_DIR" | wc -l)" -gt 10 ]; then
        print_info "... (僅顯示部分內容)"
    fi
}

# 輸入專案名稱
read_project_name() {
    print_step "設定專案名稱..."
    
    echo -n "請輸入專案名稱 (例如: github.com/username/project)："
    read PROJECT_NAME
    
    if [ -z "$PROJECT_NAME" ]; then
        print_error "專案名稱不能為空"
        exit 1
    fi
    
    # 驗證專案名稱格式
    if [[ ! "$PROJECT_NAME" =~ ^[a-zA-Z0-9_\-\.\/]+$ ]]; then
        print_warning "專案名稱包含非標準字符，可能會導致問題"
    fi
    
    print_info "專案名稱設定為: $PROJECT_NAME"
}

# 準備模板文件 - 複製到臨時目錄並且進行替換
prepare_template_files() {
    print_step "準備模板檔案..."
    
    # 創建臨時目錄
    if [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
    mkdir -p "$TMP_DIR"
    
    print_info "正在複製檔案到臨時目錄並進行替換..."
    
    # 複製文件到臨時目錄
    cp -r "$TEMPLATE_DIR"/* "$TMP_DIR"/ 2>/dev/null || true
    
    # 定義替換的模式和替換值
    local patterns=(
        "github.com/yourusername/project"
        "github.com/yourusername/shoppingcart"
        "github.com/yourusername"
    )
    
    local replacements=(
        "$PROJECT_NAME"
        "$PROJECT_NAME"
        "$PROJECT_NAME"
    )
    
    # 定義文件類型
    local file_types=(
        "*.go"
        "*.md"
        "*.Dockerfile"
        "docker-compose.yaml"
        "*.yml"
        "*.yaml"
    )
    
    # 在臨時目錄中進行替換
    local total_replacements=0
    
    for i in "${!patterns[@]}"; do
        for ext in "${file_types[@]}"; do
            local sed_cmd
            if [ "$OS_TYPE" == "macos" ]; then
                sed_cmd="sed -i '' -e"
            else
                sed_cmd="sed -i -e"
            fi
            
            find "$TMP_DIR" -type f -name "$ext" | while read file; do
                # 檢查文件是否包含模式
                if grep -q "${patterns[$i]}" "$file"; then
                    $sed_cmd "s|${patterns[$i]}|${replacements[$i]}|g" "$file"
                    let total_replacements++
                    print_info "  已處理: $(basename "$file")"
                fi
            done
        done
    done
    
    print_info "模板準備完成，進行了 $total_replacements 次替換"
}

# 複製處理過的模板文件到目標目錄
copy_prepared_files() {
    print_step "複製準備好的檔案到專案目錄..."
    
    # 目標是當前目錄
    local target_dir="$SCRIPT_DIR"
    
    # 複製文件前檢查是否有衝突
    local important_files=(
        "main.go"
        "go.mod"
        "internal"
        "pkg"
    )
    
    local existing_files=0
    for file in "${important_files[@]}"; do
        if [ -e "$target_dir/$file" ] && [ "$file" != "README.md" ]; then
            let existing_files++
        fi
    done
    
    if [ $existing_files -gt 0 ]; then
        echo -n "當前目錄已包含一些重要文件，複製操作可能會覆蓋它們。是否繼續？(y/n): "
        read answer
        if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
            print_info "操作取消"
            exit 0
        fi
    fi
    
    # 複製文件到目標目錄
    find "$TMP_DIR" -type f | while read file; do
        rel_path="${file#$TMP_DIR/}"
        target="$target_dir/$rel_path"
        
        # 確保目標目錄存在
        mkdir -p "$(dirname "$target")"
        
        # 複製文件，但避免覆蓋 README.md 如果它已存在
        if [[ "$rel_path" == "README.md" && -f "$target" ]]; then
            print_info "  已跳過: $rel_path (避免覆蓋)"
        else
            cp "$file" "$target"
            print_info "  已複製: $rel_path"
        fi
    done
    
    print_info "檔案複製完成"
}

# 初始化 Go 模組
init_go_module() {
    print_step "初始化 Go 模組為 $PROJECT_NAME..."
    
    # 如果已有go.mod，則移除它
    if [ -f go.mod ]; then
        rm go.mod
    fi
    
    # 初始化新的模組
    go mod init $PROJECT_NAME
    
    if [ $? -ne 0 ]; then
        print_error "Go 模組初始化失敗"
        exit 1
    fi
    
    print_info "Go 模組初始化完成"
}

# 檢查是否有未替換的匯入路徑
check_import_paths() {
    print_step "檢查是否有未替換的匯入路徑..."
    
    local patterns=(
        "github.com/yourusername"
        "yourusername/project"
    )
    
    local unfixed_found=0
    
    for pattern in "${patterns[@]}"; do
        UNFIXED_IMPORTS=$(grep -r "$pattern" --include="*.go" --include="*.yaml" --include="*.yml" . || true)
        
        if [ -n "$UNFIXED_IMPORTS" ]; then
            print_warning "找到未替換的匯入路徑 '$pattern'："
            echo "$UNFIXED_IMPORTS"
            unfixed_found=1
        fi
    done
    
    if [ $unfixed_found -eq 0 ]; then
        print_info "所有匯入路徑已替換成功"
    else
        print_warning "部分匯入路徑可能需要手動檢查和替換"
    fi
}

# 安裝依賴套件
install_dependencies() {
    print_step "安裝依賴套件..."
    
    # 先移除可能存在的 go.sum
    if [ -f go.sum ]; then
        rm go.sum
    fi
    
    go mod tidy
    
    if [ $? -ne 0 ]; then
        print_warning "部分依賴套件可能安裝失敗，請手動檢查"
    else
        print_info "依賴套件安裝完成"
        print_info "已添加的套件:"
        go list -m all | grep -v "^$PROJECT_NAME"
    fi
}

# 初始化 Swagger 文檔（可選）
init_swagger() {
    print_step "檢查是否初始化 Swagger 文檔..."
    
    if command -v swag &> /dev/null; then
        echo -n "是否立即初始化 Swagger 文檔？(y/n): "
        read answer
        
        if [[ "$answer" == "y" || "$answer" == "Y" ]]; then
            print_info "正在初始化 Swagger 文檔..."
            
            # 檢查 swaggo/swag 版本，以修復可能的版本兼容性問題
            go get -u github.com/swaggo/swag/cmd/swag
            go get -u github.com/swaggo/gin-swagger
            go get -u github.com/swaggo/files
            
            # 重新整理 go.mod
            go mod tidy
            
            # 初始化 swagger 文檔
            swag init -g main.go
            
            if [ $? -ne 0 ]; then
                print_warning "Swagger 文檔初始化失敗，請稍後手動執行 'swag init -g main.go'"
            else
                print_info "Swagger 文檔初始化成功"
            fi
        else
            print_info "跳過 Swagger 初始化，您可以稍後手動執行 'swag init -g main.go'"
        fi
    else
        print_warning "未找到 swag 命令，如需使用 Swagger 文檔，請先安裝："
        print_info "go install github.com/swaggo/swag/cmd/swag@latest"
    fi
}

# 顯示完成訊息
show_completion() {
    print_step "專案設定完成！"
    print_info "專案名稱: $PROJECT_NAME"
    print_info "專案位置: $(pwd)"
    print_info "專案結構:"
    
    # 顯示目錄結構 (排除隱藏目錄和 template 目錄)
    find . -type d -not -path "*/\.*" -not -path "./template*" | sort | while read dir; do
        if [ "$dir" != "." ]; then
            depth=$(echo "$dir" | tr -cd '/' | wc -c)
            indent=$(printf '%*s' $((depth*2)) '')
            echo "  $indent${dir##*/}"
        fi
    done
    
    print_info "專案已建立，可以開始開發了"
    print_info "您可以使用命令 'go run ./main.go' 運行專案"
}

# 清理腳本和模板目錄
cleanup_script_and_template() {
    print_step "清理腳本和模板目錄..."
    
    echo -n "設定已完成，是否刪除 setup.sh 和 template 目錄？(y/n): "
    read answer
    
    local delete_cmd=""
    
    if [[ "$answer" == "y" || "$answer" == "Y" ]]; then
        print_info "正在刪除 setup.sh..."
        delete_cmd="rm \"$SCRIPT_DIR/setup.sh\""
        
        if [ -d "$TEMPLATE_DIR" ]; then
            print_info "正在刪除 template 目錄..."
            delete_cmd="$delete_cmd && rm -rf \"$TEMPLATE_DIR\""
        fi
        
        # 確保臨時目錄也被刪除
        if [ -d "$TMP_DIR" ]; then
            delete_cmd="$delete_cmd && rm -rf \"$TMP_DIR\""
        fi
        
        print_info "清理完成"
    else
        print_info "保留 setup.sh 和 template 目錄"
    fi
    
    # 返回刪除命令以便稍後執行
    echo "$delete_cmd"
}

# 主程序
main() {
    echo -e "${GREEN}=================================${NC}"
    echo -e "${GREEN}   Go 微服務專案設定嚮導   ${NC}"
    echo -e "${GREEN}=================================${NC}"
    echo ""
    
    detect_os
    check_requirements
    check_template_dir
    read_project_name
    prepare_template_files
    copy_prepared_files
    init_go_module
    check_import_paths
    install_dependencies
    init_swagger
    show_completion
    
    DELETE_COMMAND=$(cleanup_script_and_template)
    
    echo ""
    print_info "感謝使用 Go 微服務專案設定嚮導！"
    
    # 如果需要刪除自身和模板目錄，使用子 shell 執行
    if [ -n "$DELETE_COMMAND" ]; then
        print_info "腳本將在顯示此消息後自行清理..."
        # 使用新的子進程執行，避免刪除當前執行的腳本出現問題
        (sleep 1 && eval "$DELETE_COMMAND") &
    fi
}

# 捕捉中斷信號
trap 'echo -e "\n${RED}操作被用戶中斷${NC}"; exit 1' INT

# 執行主程序
main