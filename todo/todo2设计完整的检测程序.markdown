基于您的设计理念和需求，我建议采用以下设计方案来构建一个完整的系统：

## 系统架构设计

### 1. 核心功能模块

系统将包含以下核心模块：
1. **订阅管理模块** - 管理用户订阅链接
2. **调度模块** - 定时执行速度测试和IP纯净度测试
3. **速度测试模块** - 基于现有的speed-check功能
4. **IP质量检测模块** - 集成ip-quality-check功能
5. **Web管理界面** - 提供图形化操作界面
6. **数据存储模块** - 存储测试结果和配置信息

### 2. 技术选型

- **后端语言**: Go (与现有speed-check保持一致)
- **前端框架**: Vue.js/React (现代化Web界面)
- **数据库**: SQLite (轻量级，易于部署)
- **任务调度**: cron或Go内置定时器
- **部署方式**: systemd服务或Docker容器

### 3. 数据库设计

```sql
-- 订阅信息表
CREATE TABLE subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 速度测试结果表
CREATE TABLE speed_test_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subscription_id INTEGER,
    node_name VARCHAR(255),
    delay INTEGER,
    download_speed REAL,
    upload_speed REAL,
    test_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
);

-- IP质量测试结果表
CREATE TABLE ip_quality_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subscription_id INTEGER,
    ip_address VARCHAR(45),
    fraud_score INTEGER,
    risk_level VARCHAR(50),
    is_proxy BOOLEAN,
    is_vpn BOOLEAN,
    is_tor BOOLEAN,
    country_code VARCHAR(10),
    test_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
);

-- 系统配置表
CREATE TABLE system_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    speed_test_interval INTEGER DEFAULT 86400, -- 默认24小时(秒)
    ip_quality_test_interval INTEGER DEFAULT 2592000, -- 默认30天(秒)
    max_concurrent_tests INTEGER DEFAULT 5
);
```

### 4. 定时任务设计

根据您的需求，系统将实现以下定时任务：

1. **每日速度测试** - 每24小时执行一次所有启用订阅的速度测试
2. **每月IP纯净度测试** - 每30天执行一次所有启用订阅的IP质量测试

### 5. 系统流程设计

```
用户操作流程:
1. 用户克隆仓库
2. 配置订阅链接
3. 启动服务(systemd/docker)
4. 通过Web界面管理配置和查看结果

系统运行流程:
1. 系统启动时读取配置
2. 初始化定时任务调度器
3. 按照配置执行定时任务
4. 将结果存储到数据库
5. 提供Web界面供用户查看
```

### 6. Web界面设计

Web界面将包含以下功能页面：

1. **仪表板** - 显示总体统计信息和最近测试结果
2. **订阅管理** - 添加/编辑/删除订阅链接
3. **测试结果** - 查看历史测试结果，支持筛选和排序
4. **系统配置** - 配置测试间隔、并发数等参数
5. **实时状态** - 显示当前正在运行的测试任务

### 7. 部署方案

#### systemd服务部署:
```ini
[Unit]
Description=SubCheck Service
After=network.target

[Service]
Type=simple
User=subcheck
WorkingDirectory=/opt/subcheck
ExecStart=/opt/subcheck/subcheck
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

#### Docker部署:
```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o subcheck .

FROM ubuntu:24.04
RUN apt-get update && apt-get install -y curl jq
COPY --from=builder /app/subcheck /usr/local/bin/subcheck
EXPOSE 8080
CMD ["subcheck"]
```

### 8. API设计

```go
// 订阅管理API
POST   /api/subscriptions        // 创建订阅
GET    /api/subscriptions        // 获取所有订阅
GET    /api/subscriptions/{id}   // 获取特定订阅
PUT    /api/subscriptions/{id}   // 更新订阅
DELETE /api/subscriptions/{id}   // 删除订阅

// 测试结果API
GET /api/results/speed          // 获取速度测试结果
GET /api/results/ip-quality     // 获取IP质量测试结果
GET /api/results/dashboard      // 获取仪表板数据

// 系统配置API
GET  /api/config               // 获取系统配置
POST /api/config               // 更新系统配置

// 手动触发测试API
POST /api/test/speed           // 手动触发速度测试
POST /api/test/ip-quality      // 手动触发IP质量测试
```

### 9. 安全和权限考虑

1. **API认证**: 实现简单的API密钥认证
2. **速率限制**: 对API请求进行速率限制
3. **输入验证**: 对所有用户输入进行验证
4. **日志记录**: 记录所有重要操作

### 10. 错误处理和监控

1. **错误重试机制**: 对网络请求实现重试机制
2. **超时控制**: 设置合理的超时时间
3. **监控告警**: 实现基本的系统健康监控
4. **日志系统**: 结构化日志记录，便于问题排查

这个设计方案能够满足您的所有需求：
- 支持Ubuntu操作系统部署
- 实现定时运行测试
- 后台稳定运行
- 支持数据库存储
- 提供图形化Web界面
- 控制IP查询频率，避免超出免费额度限制

通过这种架构，用户可以方便地克隆仓库、配置订阅、启动服务，并通过浏览器界面管理整个系统。