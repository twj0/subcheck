# subcheck

`subcheck` 是一个基于 Go 语言开发的代理订阅链接检查与管理工具。它通过自动化的方式，帮助用户测试、筛选和组织来自不同订阅源的代理节点。

## 1. 项目目标与主要功能

`subcheck` 旨在简化代理节点的管理流程，通过丰富的测试功能筛选出高质量、可用的节点，并将其转换为多种主流客户端支持的格式。

### 核心功能:
- **订阅聚合**: 将多个来源（如 Clash, V2Ray, Base64 等）的订阅链接合并为统一的节点池。
- **节点健康检查**:
    - **可用性测试**: 通过延迟测试过滤掉无法连接的节点。
    - **速度测试**: 测量节点的下载速度，以识别高性能节点。
    - **流媒体与服务解锁**: 检测节点是否能访问 Netflix, Disney+, YouTube, OpenAI 等流行服务。
- **节点处理**:
    - **去重**: 根据节点属性移除重复的代理节点。
    - **重命名**: 基于节点的 IP 地理位置信息自动重命名。
- **订阅转换**: 集成 `sub-store`，将筛选后的节点列表转换为多种格式（Clash, ClashMeta, V2Ray, Sing-Box 等）。
- **Web 界面**: 提供一个管理后台 (`/admin`)，用于状态监控、配置管理和手动触发检查。
- **灵活调度**: 支持基于固定间隔 (interval) 和 Cron 表达式的自动检查任务。
- **持久化存储**: 将测试结果和筛选后的节点列表保存到多种后端，包括：
    - 本地文件系统
    - Cloudflare R2
    - GitHub Gist
    - WebDAV
    - S3 兼容的对象存储
- **消息通知**: 通过 Apprise 支持超过100种通知渠道，发送状态更新和检查结果。
- **IP 质量分析**: 集成独立的 Shell 脚本，对 IP 地址进行深入分析，包括风险评分、代理/VPN/Tor 使用情况等。

## 2. 项目结构

项目主要由 `speed-check` Go 应用和 `ip-quality-check` Shell 脚本工具两部分组成。

- **`speed-check/`**: 核心 Go 应用程序。
    - **`main.go`**: 程序入口点。
    - **`app/app.go`**: 包含应用的核心逻辑，如调度器、配置监控和主检查循环。
    - **`check/`**: 实现延迟、速度和平台解锁测试的逻辑。
    - **`config/`**: 管理 `config.yaml` 配置文件的加载与解析。
    - **`save/`**: 包含将结果保存到不同存储后端的模块。
    - **`proxy/`**: 提供处理代理信息的工具函数。
    - **`assets/`**: 包含嵌入的资源文件，如 `sub-store` 二进制文件和 Web UI 模板。
    - **`Makefile`**: 定义构建和自动化任务。
    - **`go.mod`**: Go 模块依赖列表。
- **`ip-quality-check/`**: 一个独立的 Shell 脚本，用于高级 IP 地址分析，可被主程序调用。

## 3. 构建与运行

### 配置
1.  将 `speed-check/config/config.example.yaml` 复制为 `speed-check/config/config.yaml`。
2.  编辑 `config.yaml`，将你的订阅链接添加到 `sub-urls` 列表中。
3.  根据需要自定义其他设置，如 `check-interval` (检查间隔), `min-speed` (最低速度), `save-method` (保存方式) 和通知设置。

### 从源码构建
项目使用 `Makefile` 来简化构建过程。

- **为当前环境构建:**
  ```shell
  cd speed-check
  make build
  ```
- **为所有目标平台构建 (Linux AMD64, ARM64):**
  ```shell
  cd speed-check
  make build-all
  ```

### 运行程序
- **从源码运行:**
  ```shell
  cd speed-check
  go run . -f ./config/config.yaml
  ```
- **从二进制文件运行:**
  构建完成后，运行生成的可执行文件：
  ```shell
  cd speed-check
  ./subcheck -f ./config/config.yaml
  ```

程序将在启动时执行一次初始检查（除非设置了 Cron 计划），然后根据配置的计划周期性运行。

## 4. 开发规范

- **语言**: 主程序使用 Go 语言编写，并辅以一个工具性的 Shell 脚本。
- **配置**: 应用行为由一个中心的 `config.yaml` 文件控制。
- **依赖**: Go 模块通过 `go.mod` 进行管理。
- **Web 框架**: Web 界面基于 Gin 框架构建。
- **任务调度**: 定时任务由 `robfig/cron` 库处理。
- **文件监控**: 配置文件变更通过 `fsnotify` 进行监控。
- **代码风格**: 代码遵循标准的 Go 格式和约定。
