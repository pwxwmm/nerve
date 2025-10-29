# Nerve Center API 参考文档

## 概述

Nerve Center 提供 RESTful API 接口用于管理 Agent、任务、集群和系统。

**作者**: mmwei3 (2025-10-28)  
**组织**: Wethers: cloudWays

## 基础信息

- **基础URL**: `http://localhost:8090/api`
- **API版本**: v1
- **认证方式**: Bearer Token (在 Authorization 头中)

## Agent 管理 API

### 1. 注册 Agent

**POST** `/api/agents/register`

注册新的 Agent 到系统中。

**请求头**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体**:
```json
{
  "hostname": "server-01",
  "cpu_type": "Intel Xeon E5-2680",
  "cpu_logic": 16,
  "memsum": 34359738368,
  "memory": "32GB",
  "sn": "ABC123456",
  "product": "Dell PowerEdge R740",
  "brand": "Dell",
  "netcard": ["eth0", "eth1"],
  "basearch": "x86_64",
  "disk": {
    "total": "2TB",
    "used": "500GB"
  },
  "raid": "RAID1",
  "ipmi_ip": "192.168.1.100",
  "manageip": "192.168.1.10",
  "storageip": "192.168.2.10",
  "paramip": "192.168.3.10",
  "os": "Ubuntu 20.04",
  "gpu_num": 2,
  "gpu_type": "NVIDIA Tesla V100",
  "gpu_vendors": ["NVIDIA"],
  "agent_version": "1.0.0"
}
```

**响应**:
```json
{
  "id": "server-01-a1b2c3d4",
  "status": "registered",
  "message": "Agent registered successfully"
}
```

### 2. 获取 Agent 列表

**GET** `/api/agents`

获取所有已注册的 Agent 列表。

**响应**:
```json
{
  "agents": [
    {
      "id": "agent-001",
      "hostname": "server-01",
      "status": "online",
      "cpu_type": "Intel Xeon E5-2680",
      "cpu_logic": 16,
      "memory": "32GB",
      "os": "Ubuntu 20.04",
      "last_seen": "2025-10-28T15:30:00Z",
      "registered_at": "2025-10-27T15:30:00Z"
    }
  ],
  "total": 1
}
```

### 3. 获取单个 Agent

**GET** `/api/agents/{id}`

获取指定 ID 的 Agent 详细信息。

**响应**:
```json
{
  "agent": {
    "id": "agent-001",
    "hostname": "server-01",
    "status": "online",
    "cpu_type": "Intel Xeon E5-2680",
    "cpu_logic": 16,
    "memory": "32GB",
    "os": "Ubuntu 20.04",
    "sn": "ABC123456",
    "product": "Dell PowerEdge R740",
    "brand": "Dell",
    "netcard": ["eth0", "eth1"],
    "basearch": "x86_64",
    "gpu_num": 2,
    "gpu_type": "NVIDIA Tesla V100",
    "last_seen": "2025-10-28T15:30:00Z",
    "registered_at": "2025-10-27T15:30:00Z"
  }
}
```

### 4. 更新 Agent 状态

**PUT** `/api/agents/{id}/status`

更新指定 Agent 的状态。

**请求体**:
```json
{
  "status": "maintenance",
  "reason": "Scheduled maintenance"
}
```

**状态值**:
- `online`: 在线
- `offline`: 离线
- `maintenance`: 维护中
- `error`: 错误状态

**响应**:
```json
{
  "status": "updated",
  "message": "Agent status updated successfully",
  "agent_id": "agent-001",
  "new_status": "maintenance"
}
```

### 5. Agent 心跳

**POST** `/api/agents/{id}/heartbeat`

Agent 发送心跳信息。

**请求体**:
```json
{
  "status": "online",
  "system_info": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 23.1
  },
  "tasks": ["task-001", "task-002"]
}
```

**响应**:
```json
{
  "status": "ok",
  "message": "Heartbeat received",
  "agent_id": "agent-001"
}
```

### 6. 删除 Agent

**DELETE** `/api/agents/{id}`

从系统中删除指定的 Agent。

**响应**:
```json
{
  "status": "deleted",
  "message": "Agent deleted successfully",
  "agent_id": "agent-001"
}
```

## 系统 API

### 1. 系统健康检查

**GET** `/api/health`

检查系统健康状态。

**响应**:
```json
{
  "status": "ok",
  "timestamp": 1698507000
}
```

### 2. 系统统计

**GET** `/api/v1/system/stats`

获取系统统计信息。

**响应**:
```json
{
  "stats": {
    "total_agents": 2,
    "online_agents": 1,
    "offline_agents": 1,
    "total_clusters": 1,
    "total_alerts": 0,
    "total_tasks": 0,
    "pending_tasks": 0
  }
}
```

## Agent 安装 API

### 1. 获取安装脚本

**GET** `/api/install?token=<token>`

获取 Agent 安装脚本。

**参数**:
- `token`: 认证令牌

**响应**: 返回 bash 安装脚本

### 2. 下载 Agent 二进制文件

**GET** `/api/download?token=<token>`

下载 Agent 二进制文件。

**参数**:
- `token`: 认证令牌

**响应**: 返回 Agent 二进制文件

## 错误响应

所有 API 在出错时都会返回以下格式的响应：

```json
{
  "error": "错误描述信息"
}
```

**常见 HTTP 状态码**:
- `200`: 成功
- `400`: 请求参数错误
- `401`: 未授权
- `404`: 资源不存在
- `500`: 服务器内部错误

## 使用示例

### 使用 curl 注册 Agent

```bash
curl -X POST http://localhost:8090/api/agents/register \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "test-server",
    "cpu_type": "Intel Core i7",
    "cpu_logic": 8,
    "memory": "16GB",
    "os": "Ubuntu 20.04"
  }'
```

### 使用 curl 获取 Agent 列表

```bash
curl -X GET http://localhost:8090/api/agents
```

### 使用 curl 更新 Agent 状态

```bash
curl -X PUT http://localhost:8090/api/agents/agent-001/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "maintenance",
    "reason": "Scheduled maintenance"
  }'
```

## 注意事项

1. 所有需要认证的 API 都需要在请求头中包含有效的 Bearer Token
2. Agent ID 通常基于 hostname 生成，确保 hostname 的唯一性
3. 心跳间隔建议设置为 30 秒
4. Agent 状态会在 5 分钟内未收到心跳时自动标记为离线

---

**版本**: 1.0.0  
**最后更新**: 2025-10-28
