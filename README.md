# subcheck 
基于[subs-check](https://github.com/beck-8/subs-check)开发，

`subcheck` 是一个基于 Go 语言开发的代理订阅链接检查与管理工具。它通过自动化的方式，帮助用户测试、筛选和组织来自不同订阅源的代理节点。

## 1. 项目目标与主要功能

`subcheck` 旨在简化代理节点的管理流程，通过丰富的测试功能筛选出高质量、可用的节点，并将其转换为多种主流客户端支持的格式。

## 2. 使用指南

### 2.0 快速开始：一键部署脚本

- **执行部署脚本**：在具备 `bash` 与 `systemd` 的 Linux 主机上，可一键完成下载、配置与服务安装。

```bash
curl -fsSL https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
```
如果是大陆用户`curl` github 超时(timeout)
可以考虑使用镜像网站(比如使用[ghfast.top](https://ghfast.top))加速GitHub
```bash
curl -fsSL https://ghfast.top/https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
```

或使用 `wget`：

```bash
wget -qO- https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
```
同理
```bash
wget -qO- https://ghfast.top/https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
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

### 2.3 Web 管理面板

部署完成后，可通过浏览器访问 Web 管理面板进行可视化管理。

- **访问地址**：`http://<服务器IP>:<端口>/admin`（默认端口 `8199`）
- **功能特性**：
  - 在线编辑配置文件
  - 手动触发节点检测
  - 查看实时检测进度和状态
  - 查看日志输出
  - 查询速度测试结果和 IP 质量检测结果
  - 管理订阅链接（增删改查）
  - 数据统计仪表板

#### API 密钥说明

- **页面访问**：直接访问 `/admin` 等页面**无需** API 密钥
- **API 调用**：所有 `/api/*` 接口需要在 HTTP 请求头中携带 API 密钥：
  ```
  X-API-Key: your-api-key
  ```

- **密钥配置**：
  - 如果未在配置文件中设置 `api-key`，系统会自动生成一个 6 位数字密钥
  - 生成的密钥会在启动日志中显示：`未设置api-key，已生成一个随机api-key api-key=123456`
  - 建议在 `config.yaml` 中设置固定密钥：
    ```yaml
    enable-web-ui: true
    api-key: "your-secret-key-here"
    ```

- **订阅输出链接**（无需密钥）：
  - Clash 格式：`http://<IP>:<端口>/sub/all.yaml`
  - Base64 格式：`http://<IP>:<端口>/sub/base64.txt`
  - Mihomo 配置：`http://<IP>:<端口>/sub/mihomo.yaml`


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
