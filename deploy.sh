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
INSTALL_DIR="/opt/subcheck"
CONFIG_DIR="/etc/subcheck"
CONFIG_NAME="config.yaml"
SERVICE_NAME="subcheck.service"

# 检查root权限
[[ $EUID -ne 0 ]] && echo -e "${RED}错误：请使用root用户运行此脚本！${NC}" && exit 1

# 检查并安装依赖
install_deps() {
    echo -e "${BLUE}正在检查并安装依赖...${NC}"
    if ! command -v git &> /dev/null || ! command -v go &> /dev/null; then
        apt-get update && apt-get install -y git golang-go
    else
        echo -e "${GREEN}依赖已满足。${NC}"
    fi
}

prepare_project() {
    echo -e "${BLUE}正在准备 Go 依赖...${NC}"
    cd "$INSTALL_DIR"
    if ! go mod tidy; then
        echo -e "${RED}go mod tidy 失败，请检查网络或Go环境${NC}"
        exit 1
    fi

    if ! go mod download; then
        echo -e "${RED}下载 Go 依赖失败${NC}"
        exit 1
    fi

    echo -e "${GREEN}依赖准备完成${NC}"
}

# 克隆并安装subcheck
install_subcheck() {
    echo -e "${BLUE}正在从 GitHub 克隆 subcheck 源码...${NC}"

    if [ -d "$INSTALL_DIR" ]; then
        echo -e "${YELLOW}检测到已存在的安装目录，正在更新...${NC}"
        cd "$INSTALL_DIR"
        git pull
    else
        git clone "https://github.com/${GITHUB_REPO}.git" "$INSTALL_DIR"
        cd "$INSTALL_DIR"
    fi

    echo -e "${GREEN}源码已准备完成${NC}"
}

# 创建配置文件
create_config() {
    mkdir -p $CONFIG_DIR
    if [ -f "${CONFIG_DIR}/${CONFIG_NAME}" ]; then
        echo -e "${YELLOW}检测到已存在的配置文件，跳过创建。${NC}"
        return
    fi

    echo -e "${BLUE}正在创建配置文件...${NC}"
    EXAMPLE_CONFIG_URL="https://raw.githubusercontent.com/twj0/subcheck/master/config/config.example.yaml"
    curl -s -o "${CONFIG_DIR}/${CONFIG_NAME}" "$EXAMPLE_CONFIG_URL"

    echo -e "${GREEN}请输入您的订阅链接 (多个链接用空格分隔，直接回车跳过):${NC}"
    read -r SUB_URLS

    if [ -n "$SUB_URLS" ]; then
        # 将空格分隔的链接转换为 YAML 数组格式
        echo "sub-urls:" > /tmp/sub_urls.tmp
        for url in $SUB_URLS; do
            echo "  - $url" >> /tmp/sub_urls.tmp
        done
        # 替换配置文件中的 sub-urls 部分
        sed -i '/^sub-urls:/,/^[a-z-]*:/{ /^sub-urls:/r /tmp/sub_urls.tmp
d; /^  -/d; }' "${CONFIG_DIR}/${CONFIG_NAME}"
        rm -f /tmp/sub_urls.tmp
        echo -e "${GREEN}订阅链接已配置${NC}"
    fi

    echo -e "${GREEN}配置文件已创建: ${CONFIG_DIR}/${CONFIG_NAME}${NC}"
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
WorkingDirectory=${INSTALL_DIR}
ExecStart=/usr/bin/env bash -c 'cd ${INSTALL_DIR} && go run . -f ${CONFIG_DIR}/${CONFIG_NAME}'
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
    install_subcheck
    prepare_project
    create_config
    create_systemd_service

    echo -e "\n${GREEN}🎉 subcheck 安装完成！ 🎉${NC}"
    echo -e "\n${YELLOW}服务管理命令:${NC}"
    echo -e "  启动: ${GREEN}systemctl start ${SERVICE_NAME}${NC}"
    echo -e "  状态: ${GREEN}systemctl status ${SERVICE_NAME}${NC}"
    echo -e "  日志: ${GREEN}journalctl -u ${SERVICE_NAME} -f${NC}"
    echo -e "  停止: ${GREEN}systemctl stop ${SERVICE_NAME}${NC}"
    echo -e "\n${YELLOW}配置文件: ${GREEN}${CONFIG_DIR}/${CONFIG_NAME}${NC}"
    echo -e "\n${GREEN}现在启动服务? (Y/n):${NC}"
    read -r START_NOW
    if [[ "$START_NOW" != "n" && "$START_NOW" != "N" ]]; then
        systemctl start ${SERVICE_NAME}
        echo -e "${GREEN}服务已启动！${NC}"
        sleep 2
        systemctl status ${SERVICE_NAME} --no-pager
    fi
}

# 执行主函数
main
