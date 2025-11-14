# subcheck 
基于[subs-check](https://github.com/beck-8/subs-check)开发，

`subcheck` 是一个基于 Go 语言开发的代理订阅链接检查与管理工具。它通过自动化的方式，帮助用户测试、筛选和组织来自不同订阅源的代理节点。

## 1. 项目目标与主要功能

`subcheck` 旨在简化代理节点的管理流程，通过丰富的测试功能筛选出高质量、可用的节点，并将其转换为多种主流客户端支持的格式。

---

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

**脚本行为**：自动检测架构选择最新发布版二进制，并同步 `ipcheck/ip.sh` 与 `/etc/subcheck/config.yaml`，最终创建 `subcheck.service` systemd 服务。

如果想**删除**也可以使用curl命令或者wget命令运行题目里下的del.sh脚本

```bash
wget -qO- https://raw.githubusercontent.com/twj0/subcheck/master/del.sh | sudo bash
```
同理
```bash
wget -qO- https://ghfast.top/https://raw.githubusercontent.com/twj0/subcheck/master/del.sh | sudo bash
```


### 2.1 启动服务

部署完成后，可以通过以下方式启动和管理服务：

#### 使用 systemd（系统级安装）

如果使用 root 权限执行部署脚本，服务将作为 systemd 服务安装：

```bash
# 启动服务
sudo systemctl start subcheck.service

# 查看服务状态
sudo systemctl status subcheck.service

# 重启服务
sudo systemctl restart subcheck.service

# 查看日志
sudo journalctl -u subcheck.service -f
```

#### 使用用户级服务管理脚本

如果以普通用户执行部署脚本，将安装用户级服务管理脚本：

```bash
# 启动服务
~/.local/share/subcheck/subcheck-service start

# 或者如果 PATH 已配置
subcheck-service start

# 查看服务状态
subcheck-service status

# 重启服务
subcheck-service restart

# 查看日志
subcheck-service logs
```

#### 使用全局命令面板

无论哪种安装方式，都可以使用全局命令 `subcheck` 打开交互式管理面板：

```bash
subcheck
```

---

#### 其它方式

### 2.2 直接使用发布版

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


### 2.3 Docker 部署（可选）

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

---

## 3. 配置文件说明

### 3.1 订阅链接配置

配置文件位于 `/etc/subcheck/config.yaml`（systemd）或 `~/.config/subcheck/config.yaml`（用户模式）。

#### 订阅链接填写规则

```yaml
sub-urls:
  - https://example.com/sub1.txt
  - "https://example.com/sub2.txt"
  - https://raw.githubusercontent.com/user/repo/main/config.yaml
```

**引号说明**：
- **不加引号**：适用于简单 URL，YAML 会自动识别
- **加引号**：当 URL 包含特殊字符（如 `#`、`:`、`@`）时必须加引号

#### GitHub 订阅加速（中国大陆用户）

如果订阅链接来自 GitHub，可能被墙，建议使用加速代理：

**方法 1：使用 `github-proxy` 配置**
```yaml
github-proxy: "https://ghfast.top/"
# 或
github-proxy: "https://gh-proxy.com/"

sub-urls:
  - https://raw.githubusercontent.com/user/repo/main/nodes.yaml
```

**方法 2：直接在订阅链接中添加代理前缀**
```yaml
sub-urls:
  - https://ghfast.top/https://raw.githubusercontent.com/user/repo/main/nodes.yaml
  - https://gh-proxy.com/https://raw.githubusercontent.com/user/repo/main/config.yaml
```

**常用 GitHub 加速镜像**：
- `https://ghfast.top/`
- `https://gh-proxy.com/`
- `https://mirror.ghproxy.com/`

#### 订阅链接高级用法

**添加备注标签**：
```yaml
sub-urls:
  - https://example.com/sub.txt#我的订阅
  - https://example.com/sub2.txt#备用订阅
```
备注会自动添加到节点名称末尾，方便区分来源。

**指定订阅类型**：
```yaml
sub-urls:
  - https://example.com/sub.txt?flag=clash.meta
```

**使用时间占位符**（动态订阅）：
```yaml
sub-urls:
  - https://example.com/daily-{Y}-{m}-{d}.yaml
  - https://example.com/config/{Ymd}.yaml
```

**远程订阅清单**：
```yaml
sub-urls-remote:
  - https://example.com/sub-list.txt
  - https://raw.githubusercontent.com/user/repo/main/subscriptions.yaml
```

### 3.2 保存方式配置

支持多种保存方式，可同时保存到多个位置：

```yaml
# 单个保存方式
save-method: local

# 多个保存方式（推荐）
save-method: [local, telegraph, github-raw]
```

**支持的保存方式**：
- `local` - 本地文件系统（始终包含）
- `telegraph` - Telegraph 匿名发布（无需配置）
- `github-raw` - GitHub 仓库（需要配置 token）
- `gist` - GitHub Gist
- `webdav` - WebDAV 服务器
- `s3` - S3 兼容存储
- `r2` - Cloudflare R2

**Telegraph 配置**（可选）：
```yaml
telegraph-token: ""  # 留空则每次创建新页面
telegraph-path: ""   # 留空则每次创建新页面
```

**GitHub Raw 配置**：
```yaml
github-raw-token: "ghp_xxxxxxxxxxxx"
github-raw-owner: "your-username"
github-raw-repo: "proxy-nodes"
github-raw-branch: "main"
github-raw-path: "sub/"
```

### 3.3 其他重要配置

```yaml
# 检测间隔（分钟）
check-interval: 120

# 或使用 cron 表达式
cron-expression: "0 */2 * * *"  # 每2小时

# 并发数
concurrent: 20

# 最低速度（KB/s）
min-speed: 512

# 超时时间（毫秒）
timeout: 5000

# 监听端口
listen-port: ":8199"

# Web 管理面板
enable-web-ui: true
api-key: "123456"  # 建议修改
```

---

## 4. Web 管理面板

部署完成后，可通过浏览器访问 Web 管理面板进行可视化管理。

- **访问地址**：`http://<服务器IP>:<端口>/admin`（默认端口 `8199`）
- **功能特性**：
  - 📊 仪表盘 - 数据统计概览
  - ⚡ 速度测试 - 查看节点速度测试结果
  - 🛡️ IP纯净度 - 查看 IP 质量检测结果
  - 🔗 订阅管理 - 管理订阅链接（增删改查）
  - 📝 在线编辑配置文件
  - 🚀 手动触发节点检测
  - 📈 查看实时检测进度和状态
  - 📋 查看日志输出

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

---


## 5. 本地开发与构建

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
