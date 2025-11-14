#!/usr/bin/env bash

# subcheck 完全卸载脚本
# https://github.com/twj0/subcheck
#
# 使用方法:
#   curl -fsSL https://raw.githubusercontent.com/twj0/subcheck/master/del.sh | sudo bash

set -euo pipefail

BLUE="\033[1;34m"
GREEN="\033[1;32m"
RED="\033[1;31m"
YELLOW="\033[1;33m"
NC="\033[0m"

INSTALL_DIR="/opt/subcheck"
CONFIG_DIR="/etc/subcheck"
SERVICE_NAME="subcheck.service"
CLI_PATH="/usr/local/bin/subcheck"

[[ $EUID -ne 0 ]] && {
    echo -e "${RED}错误：请使用root用户运行此脚本！${NC}"
    exit 1
}

echo -e "${BLUE}=== subcheck 卸载脚本 ===${NC}"
echo -e "${RED}警告：此操作将完全删除 subcheck 及其所有配置文件！${NC}"


echo -e "${BLUE}停止并禁用服务...${NC}"
if systemctl is-active --quiet ${SERVICE_NAME}; then
    systemctl stop ${SERVICE_NAME}
    echo -e "${GREEN}服务已停止${NC}"
fi

if systemctl is-enabled --quiet ${SERVICE_NAME} 2>/dev/null; then
    systemctl disable ${SERVICE_NAME}
    echo -e "${GREEN}服务已禁用${NC}"
fi

echo -e "${BLUE}删除 systemd 服务文件...${NC}"
if [[ -f "/etc/systemd/system/${SERVICE_NAME}" ]]; then
    rm -f "/etc/systemd/system/${SERVICE_NAME}"
    systemctl daemon-reload
    echo -e "${GREEN}服务文件已删除${NC}"
fi

echo -e "${BLUE}删除程序文件...${NC}"
if [[ -d "$INSTALL_DIR" ]]; then
    rm -rf "$INSTALL_DIR"
    echo -e "${GREEN}程序目录已删除: ${INSTALL_DIR}${NC}"
fi

echo -e "${BLUE}删除配置文件...${NC}"
if [[ -d "$CONFIG_DIR" ]]; then
    rm -rf "$CONFIG_DIR"
    echo -e "${GREEN}配置目录已删除: ${CONFIG_DIR}${NC}"
fi

echo -e "${BLUE}删除全局命令...${NC}"
if [[ -f "$CLI_PATH" ]]; then
    rm -f "$CLI_PATH"
    echo -e "${GREEN}全局命令已删除: ${CLI_PATH}${NC}"
fi

echo -e "\n${GREEN}🎉 subcheck 已完全卸载！ 🎉${NC}"
echo -e "${YELLOW}感谢使用 subcheck！${NC}"
