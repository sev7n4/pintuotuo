#!/bin/bash

# Pintuotuo 工作流验证脚本
# 用于验证三环境一致性、分支状态等
# 用法: ./scripts/verify-workflow.sh [check|sync|status]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
SERVER_IP="119.29.173.89"
SSH_KEY="$HOME/.ssh/tencent_cloud_deploy"
DEPLOY_PATH="/opt/pintuotuo"

# 打印带颜色的消息
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查版本一致性
check_version_sync() {
    print_info "检查三环境版本一致性..."
    echo ""
    
    echo -e "${YELLOW}=== 本地环境 ===${NC}"
    LOCAL_COMMIT=$(git log --oneline -1 | cut -d' ' -f1)
    echo "Commit: $LOCAL_COMMIT"
    echo "$(git log --oneline -3)"
    echo ""
    
    echo -e "${YELLOW}=== 远程仓库 (origin/main) ===${NC}"
    REMOTE_COMMIT=$(git log origin/main --oneline -1 | cut -d' ' -f1)
    echo "Commit: $REMOTE_COMMIT"
    echo "$(git log origin/main --oneline -3)"
    echo ""
    
    echo -e "${YELLOW}=== 腾讯云服务器 ===${NC}"
    if [ -f "$SSH_KEY" ]; then
        SERVER_COMMIT=$(ssh -i "$SSH_KEY" -o ConnectTimeout=10 root@$SERVER_IP "cd $DEPLOY_PATH && git log --oneline -1" 2>/dev/null | cut -d' ' -f1)
        if [ $? -eq 0 ]; then
            echo "Commit: $SERVER_COMMIT"
            ssh -i "$SSH_KEY" root@$SERVER_IP "cd $DEPLOY_PATH && git log --oneline -3"
        else
            print_warning "无法连接到服务器"
            SERVER_COMMIT="UNKNOWN"
        fi
    else
        print_warning "SSH 密钥不存在: $SSH_KEY"
        SERVER_COMMIT="UNKNOWN"
    fi
    echo ""
    
    # 验证一致性
    echo -e "${BLUE}=== 版本一致性检查 ===${NC}"
    if [ "$LOCAL_COMMIT" = "$REMOTE_COMMIT" ]; then
        print_success "本地与远程一致: $LOCAL_COMMIT"
    else
        print_error "本地与远程不一致!"
        echo "  本地: $LOCAL_COMMIT"
        echo "  远程: $REMOTE_COMMIT"
        return 1
    fi
    
    if [ "$SERVER_COMMIT" != "UNKNOWN" ]; then
        if [ "$LOCAL_COMMIT" = "$SERVER_COMMIT" ]; then
            print_success "本地与服务器一致: $LOCAL_COMMIT"
        else
            print_warning "本地与服务器不一致!"
            echo "  本地: $LOCAL_COMMIT"
            echo "  服务器: $SERVER_COMMIT"
        fi
    fi
    
    return 0
}

# 检查分支状态
check_branch_status() {
    print_info "检查分支状态..."
    echo ""
    
    CURRENT_BRANCH=$(git branch --show-current)
    echo -e "当前分支: ${YELLOW}$CURRENT_BRANCH${NC}"
    
    STATUS=$(git status --porcelain)
    if [ -z "$STATUS" ]; then
        print_success "工作区干净"
        return 0
    else
        print_warning "工作区有未提交的变更:"
        echo "$STATUS"
        return 1
    fi
}

# 检查是否在 main 分支
check_main_branch() {
    CURRENT_BRANCH=$(git branch --show-current)
    if [ "$CURRENT_BRANCH" = "main" ]; then
        print_warning "当前在 main 分支，请创建功能分支进行开发"
        return 1
    fi
    print_success "当前分支: $CURRENT_BRANCH"
    return 0
}

# 同步到最新
sync_to_latest() {
    print_info "同步到最新版本..."
    
    CURRENT_BRANCH=$(git branch --show-current)
    
    if [ "$CURRENT_BRANCH" != "main" ]; then
        print_warning "当前不在 main 分支，请先切换到 main"
        return 1
    fi
    
    git pull origin main
    print_success "已同步到最新"
}

# 显示完整状态
show_status() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║           Pintuotuo 工作流状态检查                          ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    check_version_sync
    echo ""
    check_branch_status
    echo ""
    
    echo -e "${BLUE}=== 检查清单 ===${NC}"
    echo "□ 版本一致性已验证"
    echo "□ 工作区干净"
    echo "□ 已创建功能分支"
    echo "□ 测试通过"
    echo "□ 代码格式化完成"
    echo "□ Lint 检查通过"
    echo ""
}

# 主函数
main() {
    case "${1:-status}" in
        check)
            check_version_sync
            check_branch_status
            ;;
        sync)
            sync_to_latest
            ;;
        status)
            show_status
            ;;
        branch)
            check_main_branch
            ;;
        *)
            echo "用法: $0 {check|sync|status|branch}"
            echo ""
            echo "  check  - 检查版本一致性和分支状态"
            echo "  sync   - 同步到最新版本"
            echo "  status - 显示完整状态"
            echo "  branch - 检查当前分支"
            exit 1
            ;;
    esac
}

main "$@"
