# Nerve 项目结构

## 📁 完整目录结构

```
nerve/
├── README.md                 # 项目主文档
├── QUICKSTART.md             # 快速开始指南
├── Makefile                  # 构建和部署脚本
├── go.mod                    # Go 模块定义
├── .gitignore               # Git 忽略文件
│
├── agent/                    # Agent 组件
│   ├── main.go              # Agent 主入口
│   ├── core/
│   │   └── agent.go         # Agent 核心逻辑（注册、心跳、任务）
│   ├── pkg/
│   │   ├── log/             # 日志模块
│   │   └── sysinfo/         # 系统信息采集
│   └── config/
│       └── config.yaml      # Agent 配置文件
│
├── server/                   # Center Server 组件
│   ├── main.go              # Server 主入口
│   ├── api/
│   │   └── handler.go       # HTTP API 处理器
│   ├── core/
│   │   ├── registry.go      # Agent 注册表
│   │   └── scheduler.go     # 任务调度器
│   ├── pkg/
│   │   ├── log/             # 日志模块
│   │   └── storage/         # 存储层
│   └── config/
│       └── server.yaml      # Server 配置文件
│
├── deploy/                   # 部署文件
│   ├── install.sh           # 一键安装脚本
│   ├── nerve-agent.service  # systemd 服务配置
│   └── docker-compose.yml   # Docker Compose 配置
│
└── docs/                     # 文档目录
    ├── ARCHITECTURE.md      # 架构文档
    ├── DEPLOYMENT.md        # 部署指南
    ├── API.md               # API 文档
    ├── HOOK_PLUGIN.md       # Hook 插件文档
    └── TODO.md              # 开发计划
```

## 🎯 核心组件说明

### 1. Agent (`agent/`)

Agent 是运行在每台机器上的轻量级组件。

**主要文件:**
- `main.go` - Agent 主入口，解析参数，启动心跳和任务监听
- `core/agent.go` - 核心逻辑：注册、心跳、任务执行
- `core/task_executor.go` - 任务执行器，支持超时保护
- `core/plugin_manager.go` - 插件管理器，动态加载 Hook 插件
- `pkg/sysinfo/sysinfo.go` - 基础系统信息采集
- `pkg/sysinfo/detailed.go` - 详细系统信息采集（CPU、内存、GPU、磁盘、网络）

**工作流程:**
```
启动 → 注册 → 心跳(30s) → 拉取任务 → 执行任务 → 上报结果
```

**新增功能:**
- 任务超时保护机制
- Hook 插件动态加载
- 详细硬件信息采集
- 系统性能监控

### 2. Server (`server/`)

Center Server 是管理所有 Agent 的中央服务器。

**主要文件:**
- `main.go` - Server 主入口（标准模式）
- `main_secure.go` - Server 主入口（安全模式，TLS/HTTPS）
- `api/router.go` - API 路由和处理器（注册、心跳、任务）
- `core/registry.go` - Agent 注册表，管理在线 Agent
- `core/scheduler.go` - 任务调度器，分配任务给 Agent

**核心功能:**
- Agent 注册和管理
- 任务调度和分发
- Agent 状态监控
- 数据聚合和存储
- WebSocket 实时通信
- 集群管理
- 告警系统
- Prometheus 指标收集
- Agent 二进制分发
- 安全功能（TLS、Token、审计、权限）

**新增模块:**
- `pkg/websocket/` - WebSocket 连接管理
- `pkg/cluster/` - 多集群管理
- `pkg/alert/` - 告警系统
- `pkg/metrics/` - Prometheus 指标收集
- `pkg/binary/` - Agent 二进制分发
- `pkg/security/` - 安全功能（TLS、Token、审计、权限）

### 3. Web UI (`web/`)

Vue.js 现代化管理界面。

**主要文件:**
- `index.html` - 单页面应用，包含所有管理功能

**功能特性:**
- Agent 实时监控
- 任务管理（创建、分发、监控）
- 集群管理界面
- 告警管理界面
- 插件管理界面
- 响应式设计

### 4. 部署文件 (`deploy/`)

提供便捷的部署方案。

**安装脚本:**
```bash
curl -fsSL https://server/install.sh | sh -s -- --token=xxx
```

**Docker 部署:**
```bash
docker-compose up -d
```

**Prometheus 集成:**
```bash
cd deploy/prometheus
docker-compose up -d
```

### 5. 脚本工具 (`scripts/`)

数据库和系统管理脚本。

**主要脚本:**
- `init-db.sh` - 自动初始化 MongoDB 数据库
- `test-db.sh` - 测试数据库连接
- `mongodb-init.js` - MongoDB 初始化脚本

## 🔄 数据流

```
Agent 启动
   ↓
注册到 Center
   ↓
定期心跳 (上报系统信息)
   ↓
拉取任务 (定时轮询)
   ↓
执行任务 (命令/脚本/钩子)
   ↓
上报结果
   ↓
继续心跳循环
```

## 📊 系统信息采集

Agent 采集以下详细信息：

1. **CPU**: 型号、厂商、频率、缓存、指令集、核心数、架构
2. **Memory**: 总量、DIMM 详情、ECC 状态、内存频率
3. **GPU**: NVIDIA/AMD 检测、显存、温度、功耗、驱动版本
4. **Disk**: 容量、型号、SMART 状态、文件系统、RAID 信息
5. **Network**: 接口详情、IP 地址、流量统计、MAC 地址
6. **IPMI**: 管理接口信息、BMC 状态
7. **System**: 操作系统、内核版本、启动时间、负载

## 🪝 任务执行系统

支持三种任务类型，均具备超时保护：

1. **Command** - 执行 Shell 命令（超时保护）
2. **Script** - 执行脚本文件（沙箱隔离）
3. **Hook** - 执行插件钩子（动态加载）

**新增特性:**
- 任务超时保护机制
- Hook 插件动态加载
- 任务执行结果收集
- 任务历史记录

## 🚀 快速开始

### 标准模式
```bash
# 1. 初始化数据库
./scripts/init-db.sh

# 2. 构建
make build-all

# 3. 启动 Server
./server/nerve-center --addr :8080 --debug

# 4. 启动 Agent
./agent/nerve-agent --server=http://localhost:8080 --token=test --debug

# 5. 访问 Web UI
open http://localhost:8080/web/
```

### 安全模式
```bash
# 启动安全模式 Server
./server/nerve-center --tls --cert server.crt --key server.key --addr :8443

# 访问 HTTPS
open https://localhost:8443/web/
```

### Prometheus 集成
```bash
# 启动 Prometheus
cd deploy/prometheus
docker-compose up -d

# 访问 Prometheus
open http://localhost:9090
```

## 📝 API 端点

### Agent 管理
- `POST /api/agents/register` - 注册 Agent
- `POST /api/agents/heartbeat` - Agent 心跳
- `GET /api/agents/tasks` - 拉取任务
- `POST /api/agents/tasks/result` - 上报任务结果
- `GET /api/agents/list` - 列出所有 Agent
- `GET /api/agents/:id` - 获取 Agent 详情
- `POST /api/agents/:id/restart` - 重启 Agent

### 任务管理
- `GET /api/tasks/list` - 列出所有任务
- `POST /api/tasks` - 创建任务
- `GET /api/tasks/:id` - 获取任务详情
- `POST /api/tasks/:id/cancel` - 取消任务

### 集群管理
- `GET /api/clusters/list` - 列出所有集群
- `POST /api/clusters` - 创建集群
- `GET /api/clusters/:id` - 获取集群详情
- `PUT /api/clusters/:id` - 更新集群
- `DELETE /api/clusters/:id` - 删除集群

### 告警管理
- `GET /api/alerts/list` - 列出所有告警
- `POST /api/alerts/rules` - 创建告警规则
- `GET /api/alerts/rules` - 列出告警规则
- `POST /api/alerts/:id/resolve` - 解决告警

### 安全功能
- `POST /api/auth/login` - 用户登录
- `GET /api/tokens` - 列出 Token
- `POST /api/tokens/generate` - 生成 Token
- `POST /api/tokens/rotate` - 轮换 Token
- `GET /api/roles` - 列出角色
- `GET /api/users` - 列出用户
- `GET /api/audit/logs` - 查看审计日志

### 系统功能
- `GET /metrics` - Prometheus 指标
- `GET /api/metrics` - 系统指标
- `GET /health` - 健康检查
- `GET /ws` - WebSocket 连接

## 🔐 安全特性

### 认证和授权
- Token 认证和轮换机制
- 基于角色的访问控制（RBAC）
- 细粒度权限管理
- 用户和角色管理

### 通信安全
- TLS/HTTPS 全链路加密
- 自动证书生成（开发环境）
- 支持 CA 证书（生产环境）
- HTTP 到 HTTPS 自动重定向

### 审计和监控
- 完整审计日志系统
- 操作追踪和记录
- 安全事件监控
- Prometheus 安全指标

### 任务安全
- 任务超时保护
- 沙箱执行环境
- 权限验证
- 结果验证

## 📈 扩展性

### 规模支持
- 支持 6000+ Agent
- 多集群管理
- 水平扩展 Server
- 异步任务处理

### 监控和告警
- Prometheus 指标集成
- Grafana 仪表板
- 自定义告警规则
- 多种通知方式

### 插件系统
- Hook 插件动态加载
- 插件版本管理
- 插件配置管理
- 插件市场支持

### 数据存储
- MongoDB 主存储
- Redis 缓存层
- PostgreSQL 备选方案
- TTL 自动清理

## 🛠️ 开发命令

```bash
make help              # 查看所有命令
make build-all         # 构建所有组件
make run-server        # 运行 Server
make run-agent         # 运行 Agent
make test              # 运行测试
make clean             # 清理构建文件
make release           # 创建发布包
```

## 📚 文档

- [快速开始](QUICKSTART.md)
- [架构设计](docs/ARCHITECTURE.md)
- [API 文档](docs/API.md)
- [部署指南](docs/DEPLOYMENT.md)
- [开发计划](docs/TODO.md)

## 🎨 设计理念

**口号**: "Nerve — the distributed intelligence beneath your infrastructure."

- **轻量级**: Agent 仅 ~10MB 单文件
- **无依赖**: 不需要 Python 环境
- **易于部署**: 一条 curl 命令安装
- **高可靠**: 自动重启和故障恢复
- **可扩展**: 插件系统和 Hook 机制

## 🔮 技术栈

### 后端技术
- **语言**: Go 1.21+
- **框架**: Gin (HTTP), Gorilla WebSocket
- **存储**: MongoDB + Redis + PostgreSQL
- **监控**: Prometheus + Grafana
- **安全**: TLS/HTTPS, RBAC, 审计日志

### 前端技术
- **框架**: Vue.js 3
- **UI 库**: Element Plus
- **通信**: Axios (HTTP), WebSocket
- **构建**: 原生 HTML/JS/CSS

### 部署技术
- **容器**: Docker + Docker Compose
- **服务**: systemd
- **代理**: Nginx (可选)
- **证书**: Let's Encrypt, 自签名证书

### 通信协议
- **API**: HTTP REST API
- **实时**: WebSocket 双向通信
- **指标**: Prometheus 协议
- **安全**: TLS 1.2+

## 💡 特色功能

### 部署和运维
1. **一键安装**: curl | sh 即可部署 Agent
2. **自动心跳**: 30 秒心跳，自动检测离线
3. **二进制分发**: 自动下载和版本管理
4. **多平台支持**: Linux、Darwin、Windows

### 任务和插件
5. **任务系统**: Server 下发任务，Agent 执行
6. **Hook 机制**: 可扩展的插件系统
7. **超时保护**: 防止任务卡死
8. **动态加载**: 插件热加载

### 监控和管理
9. **数据采集**: 完整的硬件和系统信息
10. **实时监控**: WebSocket 实时通信
11. **Web UI**: Vue.js 现代化管理界面
12. **集群管理**: 多集群支持和 Agent 分组

### 告警和安全
13. **告警系统**: 规则引擎和通知机制
14. **Prometheus 集成**: 完整指标收集
15. **安全功能**: TLS、Token、审计、权限
16. **横向扩展**: 支持大规模集群管理

