#!/usr/bin/env bash

# subcheck 一键部署脚本
# https://github.com/twj0/subcheck

# 定义颜色
BLUE="\033[1;34m"
GREEN="\033[1;32m"
RED="\033[1;31m"
YELLOW="\033[1;33m"
NC="\033[0m"

# 定义项目信息
GITHUB_REPO="twj0/subcheck"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/subcheck"
BINARY_NAME="subcheck"
CONFIG_NAME="config.yaml"
SERVICE_NAME="subcheck.service"

# 检查root权限
[[ $EUID -ne 0 ]] && echo -e "${RED}错误：请使用root用户运行此脚本！${NC}" && exit 1

# 检查并安装依赖
install_deps() {
    echo -e "${BLUE}正在检查并安装依赖 (curl, tar)...${NC}"
    if ! command -v curl &> /dev/null || ! command -v tar &> /dev/null; then
        apt-get update && apt-get install -y curl tar
    else
        echo -e "${GREEN}依赖已满足。${NC}"
    fi
}

# 获取系统架构
get_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64";;
        aarch64) ARCH="arm64";;
        *) echo -e "${RED}错误：不支持的架构: $ARCH${NC}"; exit 1;;
    esac
    echo -e "${GREEN}检测到系统架构: $ARCH${NC}"
}

# 下载并安装subcheck
install_subcheck() {
    echo -e "${BLUE}正在从 GitHub 下载最新版本的 subcheck...${NC}"
    LATEST_URL=$(curl -s https://api.github.com/repos/$GITHUB_REPO/releases/latest | grep "browser_download_url.*linux_${ARCH}.tar.gz" | cut -d '"' -f 4)

    if [ -z "$LATEST_URL" ]; then
        echo -e "${RED}错误：无法找到适用于 linux_${ARCH} 的最新 Release 版本。${NC}"
        echo -e "${YELLOW}请检查 https://github.com/$GITHUB_REPO/releases 是否有对应的压缩包。${NC}"
        exit 1
    fi

    TEMP_FILE=$(mktemp)
    curl -L -o "$TEMP_FILE" "$LATEST_URL"

    echo -e "${BLUE}正在解压并安装二进制文件到 ${INSTALL_DIR}...${NC}"
    tar -xzf "$TEMP_FILE" -C /tmp/
    install "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/"
    rm -f "$TEMP_FILE"
    rm -f "/tmp/${BINARY_NAME}"

    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${GREEN}subcheck 已成功安装到 ${INSTALL_DIR}/${BINARY_NAME}${NC}"
    else
        echo -e "${RED}错误：文件安装失败！${NC}"
        exit 1
    fi
}

# 创建配置文件
create_config() {
    mkdir -p $CONFIG_DIR
    if [ -f "${CONFIG_DIR}/${CONFIG_NAME}" ]; then
        echo -e "${YELLOW}检测到已存在的配置文件，跳过创建。${NC}"
        echo -e "${YELLOW}如果需要重新生成，请先删除 ${CONFIG_DIR}/${CONFIG_NAME}${NC}"
    else
        echo -e "${BLUE}正在创建默认配置文件...${NC}"
        # 从 GitHub 下载最新的 config.example.yaml
        EXAMPLE_CONFIG_URL="https://raw.githubusercontent.com/twj0/subcheck/main/speed-check/config/config.example.yaml"
        curl -s -o "${CONFIG_DIR}/${CONFIG_NAME}" "$EXAMPLE_CONFIG_URL"
        echo -e "${GREEN}配置文件已创建在 ${CONFIG_DIR}/${CONFIG_NAME}${NC}"
        echo -e "${YELLOW}请务必修改此文件，填入您的订阅链接和相关配置！${NC}"
    fi
}

# 创建 systemd 服务
create_systemd_service() {
    echo -e "${BLUE}正在创建 systemd 服务...${NC}"
    cat > /etc/systemd/system/$SERVICE_NAME <<-EOF
[Unit]
Description=subcheck Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=${INSTALL_DIR}/${BINARY_NAME} -f ${CONFIG_DIR}/${CONFIG_NAME}
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable $SERVICE_NAME
    echo -e "${GREEN}systemd 服务已创建并设置为开机自启。${NC}"
}

# 主函数
main() {
    install_deps
    get_arch
    install_subcheck
    create_config
    create_systemd_service

    echo -e "\n${GREEN}🎉 subcheck 安装完成！ 🎉${NC}"
    echo -e "\n${YELLOW}重要提示:${NC}"
    echo -e "1. 配置文件位于: ${GREEN}${CONFIG_DIR}/${CONFIG_NAME}${NC}"
    echo -e "   ${YELLOW}请立即编辑此文件，填入您的订阅链接等信息。${NC}"
    echo -e "2. 使用以下命令管理服务:"
    echo -e "   - 启动服务: ${GREEN}systemctl start ${SERVICE_NAME}${NC}"
    echo -e "   - 查看状态: ${GREEN}systemctl status ${SERVICE_NAME}${NC}"
    echo -e "   - 查看日志: ${GREEN}journalctl -u ${SERVICE_NAME} -f${NC}"
    echo -e "   - 停止服务: ${GREEN}systemctl stop ${SERVICE_NAME}${NC}"
    echo -e "\n请按照 ${BLUE}README.md${NC} 的指引继续操作。"
}

# 执行主函数
main
