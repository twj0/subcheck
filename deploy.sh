#!/usr/bin/env bash

# subcheck 一键部署脚本
# https://github.com/twj0/subcheck
#
# 使用方法:
#   默认使用加速镜像: curl -fsSL https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
#   不使用加速镜像: curl -fsSL https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo GITHUB_PROXY= bash
#   自定义镜像: curl -fsSL https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo GITHUB_PROXY=https://gh-proxy.com/ bash

set -euo pipefail

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
BIN_NAME="subcheck"
SERVICE_NAME="subcheck.service"
IP_SCRIPT_PATH="${INSTALL_DIR}/ipcheck/ip.sh"

# GitHub加速镜像（中国大陆用户）
GITHUB_PROXY="${GITHUB_PROXY:-https://ghfast.top/}"

MODE=""

ensure_dep() {
    local dep=$1
    if ! command -v "$dep" &>/dev/null; then
        echo -e "${YELLOW}缺少依赖: $dep${NC}"
        missing_deps+=("$dep")
    fi
}

install_deps() {
    missing_deps=()
    ensure_dep curl
    ensure_dep jq
    ensure_dep tar

    if ((${#missing_deps[@]} > 0)); then
        if command -v apt-get &>/dev/null; then
            echo -e "${BLUE}安装依赖: ${missing_deps[*]}${NC}"
            apt-get update
            apt-get install -y "${missing_deps[@]}"
        elif command -v yum &>/dev/null; then
            echo -e "${BLUE}安装依赖: ${missing_deps[*]}${NC}"
            yum install -y "${missing_deps[@]}"
        elif command -v dnf &>/dev/null; then
            echo -e "${BLUE}安装依赖: ${missing_deps[*]}${NC}"
            dnf install -y "${missing_deps[@]}"
        elif command -v apk &>/dev/null; then
            echo -e "${BLUE}安装依赖: ${missing_deps[*]}${NC}"
            apk add --no-cache "${missing_deps[@]}"
        else
            echo -e "${RED}无法自动安装依赖，请手动安装: ${missing_deps[*]}${NC}"
            exit 1
        fi
    else
        echo -e "${GREEN}依赖已满足。${NC}"
    fi
}

detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) echo "linux_amd64" ;;
        aarch64|arm64) echo "linux_arm64" ;;
        armv7l|armhf) echo "linux_armv7" ;;
        armv6l) echo "linux_armv6" ;;
        *)
            echo -e "${RED}暂不支持的架构: $arch${NC}"
            exit 1
            ;;
    esac
}

fetch_latest_release() {
    echo -e "${BLUE}获取最新版本信息...${NC}"
    LATEST_JSON=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest")
    if [[ -z "$LATEST_JSON" || "$LATEST_JSON" == *"Not Found"* ]]; then
        echo -e "${RED}无法获取最新版本信息${NC}"
        exit 1
    fi
    LATEST_TAG=$(echo "$LATEST_JSON" | jq -r '.tag_name')
    if [[ -z "$LATEST_TAG" || "$LATEST_TAG" == "null" ]]; then
        echo -e "${RED}最新版本号解析失败${NC}"
        exit 1
    fi
    echo -e "${GREEN}最新版本: ${LATEST_TAG}${NC}"

    TARGET_ARCH=$(detect_arch)
    ASSET_NAME="${BIN_NAME}_${TARGET_ARCH}"
    DOWNLOAD_URL=$(echo "$LATEST_JSON" | jq -r ".assets[] | select(.name == \"${ASSET_NAME}\") | .browser_download_url")

    if [[ -z "$DOWNLOAD_URL" ]]; then
        echo -e "${RED}未找到适用于架构 ${TARGET_ARCH} 的二进制文件${NC}"
        exit 1
    fi
}

download_binary() {
    mkdir -p "$INSTALL_DIR"
    echo -e "${BLUE}下载二进制文件...${NC}"
    local proxied_url="${GITHUB_PROXY}${DOWNLOAD_URL}"
    curl -L "$proxied_url" -o "${INSTALL_DIR}/${BIN_NAME}"
    chmod +x "${INSTALL_DIR}/${BIN_NAME}"
    echo -e "${GREEN}二进制文件已安装到 ${INSTALL_DIR}/${BIN_NAME}${NC}"
}

prepare_assets() {
    mkdir -p "${INSTALL_DIR}/ipcheck"
    if [[ ! -f "$IP_SCRIPT_PATH" ]]; then
        echo -e "${BLUE}下载 ip.sh...${NC}"
        curl -sL "${GITHUB_PROXY}https://raw.githubusercontent.com/twj0/IPQuality/main/ip.sh" -o "$IP_SCRIPT_PATH"
        chmod +x "$IP_SCRIPT_PATH"
    else
        echo -e "${GREEN}检测到 existing ip.sh，跳过下载。${NC}"
    fi

    mkdir -p "$CONFIG_DIR"
    if [[ ! -f "${CONFIG_DIR}/${CONFIG_NAME}" ]]; then
        echo -e "${BLUE}下载配置模板...${NC}"
        curl -sL "${GITHUB_PROXY}https://raw.githubusercontent.com/${GITHUB_REPO}/master/config/config.example.yaml" -o "${CONFIG_DIR}/${CONFIG_NAME}"
        echo -e "${GREEN}配置文件已写入: ${CONFIG_DIR}/${CONFIG_NAME}${NC}"
    else
        echo -e "${YELLOW}检测到已有配置文件，保留现有配置。${NC}"
    fi
}

create_systemd_service() {
    echo -e "${BLUE}生成 systemd 服务...${NC}"
    cat > /etc/systemd/system/$SERVICE_NAME <<-EOF
[Unit]
Description=subcheck Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${BIN_NAME} -f ${CONFIG_DIR}/${CONFIG_NAME}
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    systemctl enable $SERVICE_NAME
    echo -e "${GREEN}systemd 服务已创建并设置为开机自启。${NC}"
}

install_global_command() {
    echo -e "${BLUE}安装全局命令...${NC}"
    if [[ "${MODE}" == "systemd" ]]; then
        local dest="/usr/local/bin/subcheck"
        curl -sL "${GITHUB_PROXY}https://raw.githubusercontent.com/${GITHUB_REPO}/master/subcheck-cli.sh" -o "$dest"
        chmod +x "$dest"
    else
        local dest="$HOME/.local/bin/subcheck"
        mkdir -p "$HOME/.local/bin"
        curl -sL "${GITHUB_PROXY}https://raw.githubusercontent.com/${GITHUB_REPO}/master/subcheck-cli.sh" -o "$dest"
        chmod +x "$dest"
        echo -e "${YELLOW}如命令未生效，请将以下路径加入PATH: ${GREEN}export PATH=\"$HOME/.local/bin:$PATH\"${NC}"
    fi
    echo -e "${GREEN}全局命令已安装，可使用 'subcheck' 命令打开管理面板${NC}"
}

configure_sub_urls() {
    echo -e "${GREEN}请输入您的订阅链接 (多个链接用空格分隔，直接回车跳过):${NC}"
    read -r SUB_URLS || true
    [[ -z "$SUB_URLS" ]] && return

    TMP_FILE=$(mktemp)
    for url in $SUB_URLS; do
        echo "$url" >>"$TMP_FILE"
    done

    awk -v urls_file="$TMP_FILE" '
    function load_urls() {
        if (loaded) return
        loaded = 1
        while ((getline line < urls_file) > 0) {
            if (length(line) > 0) {
                urls[++idx] = line
            }
        }
        close(urls_file)
    }
    function print_urls() {
        load_urls()
        for (i = 1; i <= idx; i++) {
            printf("  - %s\n", urls[i])
        }
    }
    {
        if (!done && /^sub-urls:/) {
            print "sub-urls:"
            print_urls()
            done = 1
            skip = 1
            next
        }
        if (skip) {
            if ($0 ~ /^[A-Za-z0-9_-]+:/) {
                skip = 0
                print $0
            }
            next
        }
        print $0
    }
    END {
        if (!done) {
            print ""
            print "sub-urls:"
            print_urls()
        }
    }
    ' "${CONFIG_DIR}/${CONFIG_NAME}" >"${CONFIG_DIR}/${CONFIG_NAME}.tmp"

    mv "${CONFIG_DIR}/${CONFIG_NAME}.tmp" "${CONFIG_DIR}/${CONFIG_NAME}"
    rm -f "$TMP_FILE"
    echo -e "${GREEN}订阅链接已写入配置文件。${NC}"
}

start_service_prompt() {
    echo -e "\n${GREEN}🎉 subcheck 安装完成！ 🎉${NC}"
    echo -e "\n${YELLOW}快速管理:${NC}"
    echo -e "  管理面板: ${GREEN}subcheck${NC}"
    echo -e "\n${YELLOW}服务管理命令:${NC}"
    if [[ "${MODE}" == "systemd" ]]; then
        echo -e "  启动: ${GREEN}systemctl start ${SERVICE_NAME}${NC}"
        echo -e "  状态: ${GREEN}systemctl status ${SERVICE_NAME}${NC}"
    else
        echo -e "  启动: ${GREEN}subcheck-service start${NC}"
        echo -e "  状态: ${GREEN}subcheck-service status${NC}"
        echo -e "  日志: ${GREEN}subcheck-service logs${NC}"
    fi
    echo -e "\n${YELLOW}Web控制面板:${NC}"
    echo -e "  地址: ${GREEN}http://YOUR_IP:8199/admin${NC}"
    echo -e "  密钥: ${GREEN}123456${NC} (请在配置文件中修改)"
}

main() {
    echo -e "${BLUE}=== subcheck 一键部署脚本 ===${NC}"
    if [[ -n "$GITHUB_PROXY" ]]; then
        echo -e "${GREEN}使用GitHub加速镜像: ${GITHUB_PROXY}${NC}"
        echo -e "${YELLOW}如需禁用加速，请设置: GITHUB_PROXY= bash deploy.sh${NC}"
    else
        echo -e "${YELLOW}未使用GitHub加速镜像，下载可能较慢${NC}"
    fi
    echo ""

    MODE="${SUBCHECK_MODE:-}"
    if [[ -z "$MODE" ]]; then
        pid1="$(cat /proc/1/comm 2>/dev/null || ps -p 1 -o comm= 2>/dev/null || echo "")"
        if command -v systemctl &>/dev/null && [[ "${pid1,,}" == systemd* ]]; then
            MODE="systemd"
        else
            MODE="user"
        fi
    fi

    if [[ "$MODE" == "systemd" ]]; then
        if [[ $EUID -ne 0 ]]; then
            echo -e "${RED}错误：请使用root用户运行此脚本！${NC}"
            exit 1
        fi
    else
        XDG_DATA_HOME="${XDG_DATA_HOME:-$HOME/.local/share}"
        XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"
        XDG_STATE_HOME="${XDG_STATE_HOME:-$HOME/.local/state}"
        INSTALL_DIR="${XDG_DATA_HOME}/subcheck"
        CONFIG_DIR="${XDG_CONFIG_HOME}/subcheck"
        IP_SCRIPT_PATH="${INSTALL_DIR}/ipcheck/ip.sh"
        mkdir -p "$INSTALL_DIR" "$CONFIG_DIR" "${XDG_STATE_HOME}/subcheck/logs"
    fi

    install_deps
    fetch_latest_release
    download_binary
    prepare_assets
    configure_sub_urls

    if [[ "$MODE" == "systemd" ]]; then
        create_systemd_service
    else
        svc_path="${INSTALL_DIR}/subcheck-service"
        curl -sL "${GITHUB_PROXY}https://raw.githubusercontent.com/${GITHUB_REPO}/master/subcheck-service" -o "$svc_path"
        chmod +x "$svc_path"
        mkdir -p "$HOME/.local/bin"
        cp -f "$svc_path" "$HOME/.local/bin/subcheck-service" 2>/dev/null || true
    fi

    install_global_command
    start_service_prompt
}

main
