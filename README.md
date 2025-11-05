# subcheck

`subcheck` 是一个基于 Go 语言开发的代理订阅链接检查与管理工具。它通过自动化的方式，帮助用户测试、筛选和组织来自不同订阅源的代理节点。

## 1. 项目目标与主要功能

`subcheck` 旨在简化代理节点的管理流程，通过丰富的测试功能筛选出高质量、可用的节点，并将其转换为多种主流客户端支持的格式。


## 2. 构建与运行

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

## 感谢
- [IPQuality](https://github.com/xykt/IPQuality)
- [subs-check](https://github.com/beck-8/subs-check)
