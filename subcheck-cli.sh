#!/usr/bin/env bash

# subcheck 全局命令面板
# 安装位置: /usr/local/bin/subcheck

set -euo pipefail

BLUE="\033[1;34m"
GREEN="\033[1;32m"
RED="\033[1;31m"
YELLOW="\033[1;33m"
NC="\033[0m"

INSTALL_DIR="/opt/subcheck"
CONFIG_DIR="/etc/subcheck"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"
SERVICE_NAME="subcheck.service"
GITHUB_REPO="twj0/subcheck"
GITHUB_PROXY="${GITHUB_PROXY:-https://ghfast.top/}"

show_menu() {
    clear
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}   subcheck 管理面板${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
    echo -e "${GREEN}0.${NC} 退出"
    echo -e "${GREEN}1.${NC} 更新程序"
    echo -e "${GREEN}2.${NC} 卸载程序"
    echo -e "${GREEN}3.${NC} 编辑配置文件"
    echo -e "${GREEN}4.${NC} 查看服务状态"
    echo -e "${GREEN}5.${NC} 重启服务"
    echo -e "${GREEN}6.${NC} 查看日志"
    echo ""
    echo -ne "${YELLOW}请选择操作 [0-6]: ${NC}"
}

update_subcheck() {
    echo -e "${BLUE}开始更新 subcheck...${NC}"

    if ! command -v jq &>/dev/null; then
        echo -e "${RED}缺少依赖 jq，请先安装${NC}"
        return 1
    fi

    LATEST_JSON=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest")
    LATEST_TAG=$(echo "$LATEST_JSON" | jq -r '.tag_name')

    if [[ -z "$LATEST_TAG" || "$LATEST_TAG" == "null" ]]; then
        echo -e "${RED}获取最新版本失败${NC}"
        return 1
    fi

    echo -e "${GREEN}最新版本: ${LATEST_TAG}${NC}"

    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) TARGET_ARCH="linux_amd64" ;;
        aarch64|arm64) TARGET_ARCH="linux_arm64" ;;
        armv7l|armhf) TARGET_ARCH="linux_armv7" ;;
        armv6l) TARGET_ARCH="linux_armv6" ;;
        *)
            echo -e "${RED}不支持的架构: $arch${NC}"
            return 1
            ;;
    esac

    DOWNLOAD_URL=$(echo "$LATEST_JSON" | jq -r ".assets[] | select(.name == \"subcheck_${TARGET_ARCH}\") | .browser_download_url")

    if [[ -z "$DOWNLOAD_URL" ]]; then
        echo -e "${RED}未找到适用的二进制文件${NC}"
        return 1
    fi

    systemctl stop ${SERVICE_NAME} 2>/dev/null || true

    echo -e "${BLUE}下载新版本...${NC}"
    curl -L "${GITHUB_PROXY}${DOWNLOAD_URL}" -o "${INSTALL_DIR}/subcheck.new"
    chmod +x "${INSTALL_DIR}/subcheck.new"
    mv "${INSTALL_DIR}/subcheck.new" "${INSTALL_DIR}/subcheck"

    systemctl start ${SERVICE_NAME}

    echo -e "${GREEN}更新完成！${NC}"
}

uninstall_subcheck() {
    echo -e "${RED}确认要卸载 subcheck 吗？此操作不可恢复！${NC}"
    echo -ne "${YELLOW}输入 yes 确认卸载: ${NC}"
    read -r confirm

    if [[ "$confirm" != "yes" ]]; then
        echo -e "${YELLOW}已取消卸载${NC}"
        return
    fi

    curl -fsSL "${GITHUB_PROXY}https://raw.githubusercontent.com/${GITHUB_REPO}/master/del.sh" | bash
}

edit_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo -e "${RED}配置文件不存在: ${CONFIG_FILE}${NC}"
        return 1
    fi

    if command -v nano &>/dev/null; then
        nano "$CONFIG_FILE"
    elif command -v vim &>/dev/null; then
        vim "$CONFIG_FILE"
    elif command -v vi &>/dev/null; then
        vi "$CONFIG_FILE"
    else
        echo -e "${RED}未找到可用的编辑器 (nano/vim/vi)${NC}"
        return 1
    fi

    echo -e "${YELLOW}配置已修改，是否重启服务使其生效？[y/N]${NC}"
    read -r restart
    if [[ "$restart" =~ ^[Yy]$ ]]; then
        systemctl restart ${SERVICE_NAME}
        echo -e "${GREEN}服务已重启${NC}"
    fi
}

show_status() {
    systemctl status ${SERVICE_NAME} --no-pager
}

restart_service() {
    echo -e "${BLUE}重启服务...${NC}"
    systemctl restart ${SERVICE_NAME}
    echo -e "${GREEN}服务已重启${NC}"
}

show_logs() {
    echo -e "${BLUE}显示最近日志 (按 Ctrl+C 退出)${NC}"
    journalctl -u ${SERVICE_NAME} -f --no-pager
}

main() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}请使用 root 权限运行${NC}"
        exit 1
    fi

    while true; do
        show_menu
        read -r choice

        case $choice in
            0)
                echo -e "${GREEN}再见！${NC}"
                exit 0
                ;;
            1)
                update_subcheck
                ;;
            2)
                uninstall_subcheck
                exit 0
                ;;
            3)
                edit_config
                ;;
            4)
                show_status
                ;;
            5)
                restart_service
                ;;
            6)
                show_logs
                ;;
            *)
                echo -e "${RED}无效选项${NC}"
                ;;
        esac

        echo ""
        echo -ne "${YELLOW}按回车键继续...${NC}"
        read -r
    done
}

main
