# Nerve - Distributed Infrastructure Intelligence Platform

**Nerve** — the distributed intelligence beneath your infrastructure.

A lightweight, production-ready agent system for large-scale infrastructure monitoring, asset management, and control across thousands of machines.

## 👥 Authors

- **mmwei3** (2025-10-28)
- **组织**: 智算运营部

## 🎯 Features

- **Zero-dependency installation**: Single binary, no Python needed
- **One-command setup**: `curl | sh` install with authentication
- **Agent ↔ Server** bidirectional communication with heartbeat (WebSocket)
- **Complete asset collection**: Detailed hardware info (CPU, Memory, GPU, Disk, Network, IPMI)
- **Task execution engine**: Commands, scripts, and Hook plugins with timeout protection
- **Plugin system**: Dynamic Hook plugin loading and execution
- **Web UI**: Vue.js modern management interface
- **Cluster management**: Multi-cluster support with Agent grouping
- **Alert system**: Rule engine and notification mechanism
- **Prometheus integration**: Complete metrics collection and monitoring
- **Agent binary distribution**: Automatic binary download and installation
- **Task execution**: Remote command execution and hook system
- **Horizontal scaling**: Designed for 6000+ machines across clusters
- **MongoDB + Redis**: High-performance data storage

## 🏗️ Architecture

```
┌─────────────┐         HTTP/gRPC          ┌──────────────┐
│   Agent     │ ◄──────────────────────────► │   Center     │
│ (nerve-agent)│   Heartbeat + Tasks       │(nerve-center)│
│             │                             │              │
│ • Collector │                             │ • API Server │
│ • Heartbeat │                             │ • Scheduler  │
│ • Task Exec │                             │ • Registry   │
│ • Hook Sys  │                             │ • Storage    │
└─────────────┘                             └──────────────┘
```

## 🚀 Quick Start

### 立即开始（3步）

```bash
# 1. 初始化数据库（MongoDB）
./scripts/init-db.sh

# 2. 构建并启动 Server
export GOPROXY="https://mirrors.aliyun.com/goproxy/,direct" && export GO111MODULE=on && go env GOPROXY
go mod download
cd server && go build -o nerve-center && cd ..
./server/nerve-center --addr :8090 --debug

# 3. 验证运行
curl http://localhost:8090/health
```

### 📦 数据库配置

系统已配置 MongoDB 和 Redis：

- **MongoDB**: `172.29.228.139` (nerve database)
- **Redis**: `172.29.228.139:6379`
- **配置文件**: `server/config/server.yaml`

详细步骤请查看 **[快速开始指南](QUICKSTART.md)**

### Agent Installation

```bash
curl -fsSL https://your-server/install.sh | sh -s -- \
  --token=<auth_token> \
  --server=https://nerve-center.example.com/api
```

The agent will:
1. Download and install the binary
2. Register with the server
3. Start collecting system info
4. Begin heartbeat (every 30s)
5. Listen for tasks and hooks

## 📊 Data Collection

Agent collects:
- CPU (type, cores, threads, info)
- Memory (total, DIMM info)
- GPU (count, type, vendors, details)
- Disk (RAID, partitions, usage)
- Network (interfaces, IPs)
- IPMI info
- OS version
- Hostname, SN, product, brand

## 🪝 Hook System

```yaml
# Example hook plugin
name: custom-script
description: Run custom script
trigger: scheduled
script: /opt/scripts/monitor.sh
```

## 🔐 Security

- Token-based authentication
- TLS/HTTPS support
- Agent-server mutual authentication
- Secure task execution

## 📝 Project Structure

```
nerve/
├── agent/           # Agent implementation
├── server/          # Center server implementation  
├── deploy/          # Deployment scripts
├── scripts/         # Database & utility scripts
└── docs/            # Documentation
```

详细结构请查看 [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)

## 📚 Documentation

- [快速开始](QUICKSTART.md) - Getting started guide
- [项目结构](PROJECT_STRUCTURE.md) - Project structure
- [架构设计](docs/ARCHITECTURE.md) - Architecture design
- [API 文档](docs/API.md) - API documentation
- [部署指南](docs/DEPLOYMENT.md) - Deployment guide
- [Hook 插件](docs/HOOK_PLUGIN.md) - Plugin system

## 🌟 Roadmap

- [ ] Web UI (Vue + REST API)
- [ ] Plugin marketplace
- [ ] Advanced scheduling & alerting
- [ ] Multi-cluster management
- [ ] gRPC streaming for real-time operations

## 🛠️ Development

```bash
make build-all         # Build all components
make run-server        # Run server
make run-agent         # Run agent
make test              # Run tests
```

## 📄 License

MIT License
