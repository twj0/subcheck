# subcheck

`subcheck` 是一个基于 Go 语言开发的代理订阅链接检查与管理工具。它通过自动化的方式，帮助用户测试、筛选和组织来自不同订阅源的代理节点。

## 1. 项目目标与主要功能

`subcheck` 旨在简化代理节点的管理流程，通过丰富的测试功能筛选出高质量、可用的节点，并将其转换为多种主流客户端支持的格式。

## 2. 快速安装

### 一键安装（推荐）
在 Ubuntu/Debian 系统上使用以下命令一键安装：

```bash
curl -fsSL https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
```

或使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/twj0/subcheck/master/deploy.sh | sudo bash
```

安装过程中可以输入订阅链接，或直接回车使用默认配置。安装完成后会提示是否立即启动服务。

## 3. 手动构建与运行

### 配置
1.  将 `config/config.example.yaml` 复制为 `config/config.yaml`
2.  编辑 `config.yaml`，将你的订阅链接添加到 `sub-urls` 列表中
3.  根据需要自定义其他设置，如 `check-interval` (检查间隔), `min-speed` (最低速度), `save-method` (保存方式) 和通知设置

### 从源码构建
项目使用 `Makefile` 来简化构建过程。

- **为当前环境构建:**
  ```shell
  make build
  ```
- **为所有目标平台构建 (Linux AMD64, ARM64):**
  ```shell
  make build-all
  ```

### 运行程序
- **从源码运行:**
  ```shell
  go run . -f ./config/config.yaml
  ```
- **从二进制文件运行:**
  ```shell
  ./subcheck -f ./config/config.yaml
  ```

程序将在启动时执行一次初始检查（除非设置了 Cron 计划），然后根据配置的计划周期性运行。

## 4. Docker 部署

### 构建镜像
在已有 Docker 的环境中进入项目根目录，执行：

```bash
docker build -t subcheck:latest .
```

构建时可通过以下构建参数覆盖版本信息：

```bash
docker build \
  --build-arg VERSION=v1.0.0 \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  -t subcheck:latest .
```

### 启动容器
默认镜像会在容器内复制 `config/` 与 `assets/` 目录。首次运行前请修改 `config/config.yaml` 或挂载外部配置目录。

```bash
docker run -d --name subcheck \
  -p 14567:14567 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/output:/app/output \
  --restart=always \
  subcheck:latest
```

- `-p 14567:14567`：将 Web UI 映射到宿主机端口，可按需调整。
- `-v $(pwd)/config:/app/config`：挂载本地配置目录，便于在宿主机编辑 `config.yaml`。
- `-v $(pwd)/output:/app/output`：持久化生成的订阅文件（`SaveMethod` 为 `local` 时使用）。
- 根据需要附加 `-e LOG_LEVEL=debug`、`-e MIHOMO_DEBUG=1` 等环境变量。

容器后台运行后，可通过 `http://<VPS_IP>:14567` 访问管理界面。使用 `docker logs -f subcheck` 查看日志，使用 `docker stop subcheck && docker rm subcheck` 停止并移除容器。

### Docker Compose
也可使用 `docker-compose.yml` 快速部署：

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

执行 `docker compose up -d --build` 即可构建并启动服务，后续使用 `docker compose logs -f` 查看日志。

## 感谢
- [IPQuality](https://github.com/xykt/IPQuality)
- [subs-check](https://github.com/beck-8/subs-check)
