# subcheck

`subcheck` 是一个基于 Go 语言开发的代理订阅链接检查与管理工具。它通过自动化的方式，帮助用户测试、筛选和组织来自不同订阅源的代理节点。

## 1. 项目目标与主要功能

`subcheck` 旨在简化代理节点的管理流程，通过丰富的测试功能筛选出高质量、可用的节点，并将其转换为多种主流客户端支持的格式。

## 2. 使用指南

### 2.0 快速开始：一键部署脚本

- **执行部署脚本**：在具备 `bash` 与 `systemd` 的 Linux 主机上，可一键完成下载、配置与服务安装。

```bash
curl -fsSL https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
```

或使用 `wget`：

```bash
wget -qO- https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
```

- **脚本行为**：自动检测架构选择最新发布版二进制，并同步 `ipcheck/ip.sh` 与 `/etc/subcheck/config.yaml`，最终创建 `subcheck.service` systemd 服务。

### 2.1 推荐：直接使用发布版二进制

- **确认架构**：在目标 VPS 上执行 `uname -m`（可能返回 `x86_64`、`aarch64` 等），并在 [GitHub Releases](https://github.com/twj0/subcheck/releases) 页面选择匹配架构的最新版本（文件名形如 `subcheck_linux_<arch>`）。

- **下载与赋权**：示例以 Linux AMD64 为例，请替换为最新版本号或使用脚本自动获取。

```bash
VERSION=$(curl -s https://api.github.com/repos/twj0/subcheck/releases/latest | jq -r .tag_name)
wget https://github.com/twj0/subcheck/releases/download/${VERSION}/subcheck_linux_amd64
```

```bash
chmod +x subcheck_linux_amd64
```

- **准备配置**：复制模板并按需修改订阅链接、监听端口等参数。

```bash
mkdir -p ~/subcheck/config
curl -fsSL -o ~/subcheck/config/config.yaml \
  https://raw.githubusercontent.com/twj0/subcheck/master/config/config.example.yaml
vi ~/subcheck/config/config.yaml
```

- **运行**：监听端口由 `config.yaml` 的 `listen-port` 控制，Web 面板位于 `http://<VPS_IP>:<端口>/admin`。

```bash
./subcheck_linux_amd64 -f ~/subcheck/config/config.yaml
```

> 建议将二进制移动到 `/usr/local/bin/subcheck` 并结合 `systemd`、`nohup` 等方式守护运行。

### 2.2 Docker 部署（可选）

- **构建镜像**：

```bash
docker build -t subcheck:latest .
```

- **启动容器**：挂载本地配置与输出目录，便于管理。

```bash
docker run -d --name subcheck \
  -p 14567:14567 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/output:/app/output \
  --restart=always \
  subcheck:latest
```

- **Docker Compose 示例**：

```yaml
services:
  subcheck:
    build:
      context: .
    image: subcheck:latest
    container_name: subcheck
    ports:
      - "14567:14567"
    volumes:
      - ./config:/app/config
      - ./output:/app/output
    environment:
      - LOG_LEVEL=info
    restart: always
```

执行 `docker compose up -d --build` 即可完成部署。

## 3. 本地开发与构建

- **环境要求**：Go `1.24` 及以上、Git、GNU Make（可选，仅在使用 `Makefile` 时需要）。
- **克隆与初始化**：

```bash
git clone https://github.com/twj0/subcheck.git
cd subcheck
cp config/config.example.yaml config/config.yaml
```

- **构建 Go 二进制**（Windows PowerShell 示例，可按需调整架构）：

```powershell
$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"
go build -trimpath -ldflags "-s -w -X main.Version=dev -X main.CurrentCommit=unknown" -o subcheck_linux_arm64
```

```powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -trimpath -ldflags "-s -w -X main.Version=dev -X main.CurrentCommit=unknown" -o subcheck_linux_amd64
```



## 感谢
- [IPQuality](https://github.com/xykt/IPQuality)
- [subs-check](https://github.com/beck-8/subs-check)
