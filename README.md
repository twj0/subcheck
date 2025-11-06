# subcheck

`subcheck` 是一个基于 Go 语言开发的代理订阅链接检查与管理工具。它通过自动化的方式，帮助用户测试、筛选和组织来自不同订阅源的代理节点。

## 1. 项目目标与主要功能

`subcheck` 旨在简化代理节点的管理流程，通过丰富的测试功能筛选出高质量、可用的节点，并将其转换为多种主流客户端支持的格式。

## 2. 使用指南

### 2.1 推荐：直接使用发布版二进制

- **确认架构**：在目标 VPS 或服务器上执行 `uname -m`（可能返回 `x86_64`、`aarch64` 等），并在 [GitHub Releases](https://github.com/twj0/subcheck/releases) 页面选择匹配架构的最新版本。


```bash
wget https://github.com/twj0/subcheck/releases/download/0.0.002/subcheck_linux_amd64
chmod +x subcheck_linux_amd64
```

- **准备配置**：复制模板并按需修改订阅链接、监听端口等参数。

```bash
mkdir -p ~/subcheck/config
cp config/config.example.yaml ~/subcheck/config/config.yaml
vim ~/subcheck/config/config.yaml
```

- **运行**：通过配置文件启动。监听端口由 `config.yaml` 中的 `listen-port` 控制，Web 面板位于 `http://<VPS_IP>:<端口>/admin`。

```bash
./subcheck_linux_amd64 -f ~/subcheck/config/config.yaml
```

> 如需长期运行，可将可执行文件移动到 `/usr/local/bin/subcheck`，并结合 `systemd`、`nohup` 或进程管理工具维护服务。

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

- **环境要求**：Go `1.24` 及以上、Git、GNU Make（可选）。
- **克隆与初始化**：

```bash
git clone https://github.com/twj0/subcheck.git
cd subcheck
cp config/config.example.yaml config/config.yaml
```

- **本地运行**：

```bash
go run . -f ./config/config.yaml
```

- **构建golang二进制**：

```powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -trimpath -ldflags "-s -w -X main.Version=dev -X main.CurrentCommit=unknown" -o subcheck_linux_amd64
```

```powershell
$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"
go build -trimpath -ldflags "-s -w -X main.Version=dev -X main.CurrentCommit=unknown" -o subcheck_linux_arm64
```



## 感谢
- [IPQuality](https://github.com/xykt/IPQuality)
- [subs-check](https://github.com/beck-8/subs-check)
