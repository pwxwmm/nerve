# Nerve Architecture

## Overview

Nerve is a distributed infrastructure intelligence platform designed to monitor and manage thousands of machines across multiple clusters.

## Components

### 1. Agent (nerve-agent)

The agent is a lightweight Go binary (~10MB) that runs on each managed machine.

**Responsibilities:**
- System information collection (CPU, Memory, GPU, Disk, Network)
- Heartbeat to maintain connection with center
- Task execution (commands, scripts, hooks)
- Secure communication with center

**Key Modules:**
- `core/agent.go` - Main agent logic
- `pkg/sysinfo/` - System information collection
- `core/heartbeat.go` - Heartbeat mechanism
- `core/task.go` - Task execution

### 2. Center (nerve-center)

The center is the management server that coordinates all agents.

**Responsibilities:**
- Agent registration and lifecycle management
- Task scheduling and distribution
- Data aggregation and storage
- API for management and monitoring

**Key Modules:**
- `api/` - HTTP REST API handlers
- `core/registry.go` - Agent registry
- `core/scheduler.go` - Task scheduling
- `pkg/storage/` - Data storage layer

## Communication Flow

```
Agent                          Center
  |                              |
  |---- Register (POST) -------->|
  |<----- 200 OK ----------------|
  |                              |
  |---- Heartbeat (30s) -------->|
  |<----- 200 OK ----------------|
  |                              |
  |<---- Get Tasks (GET) --------|
  |<----- Tasks [] --------------|
  |                              |
  |---- Execute Task ----------->|
  |                              |
  |---- Task Result (POST) ----->|
  |<----- 200 OK ----------------|
```

## Data Collection

The agent collects:

1. **CPU Information**
   - Model name
   - Logical cores
   - Architecture

2. **Memory Information**
   - Total memory
   - Individual DIMM details

3. **GPU Information**
   - GPU count
   - GPU type (NVIDIA/AMD)
   - Vendor information

4. **Disk Information**
   - RAID configuration
   - Partition details
   - Usage statistics

5. **Network Information**
   - Network interfaces
   - IP addresses

6. **IPMI Information**
   - IPMI IP address
   - Management interface

## Task System

Tasks are categorized into:

1. **Command Tasks** - Execute shell commands
2. **Script Tasks** - Execute script files
3. **Hook Tasks** - Execute plugin hooks

Tasks are:
- Polled by agents periodically
- Executed with timeout protection
- Results reported back to center

## Security

- **Authentication**: Token-based (Bearer token)
- **Transport**: HTTPS (TLS encryption)
- **Validation**: Server validates agent tokens
- **Isolation**: Tasks run in isolated context

## Scalability

Designed for:
- 6000+ agents
- Multiple clusters
- Horizontal server scaling
- Async task processing

## Storage

Current implementation uses in-memory storage. Production deployments should use:
- PostgreSQL for persistent storage
- Redis for caching
- etcd for coordination

## Future Enhancements

- WebSocket for real-time communication
- gRPC for efficient binary protocols
- Plugin system for extensibility
- Web UI for management
- Metrics and alerting

