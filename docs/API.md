# Nerve API Documentation

## Base URL

```
http://nerve-center:8080/api
```

## Authentication

All endpoints require Bearer token authentication:

```
Authorization: Bearer <token>
```

## Endpoints

### 1. Agent Registration

Register a new agent with the center.

**Request**
```
POST /agents/register
Content-Type: application/json

{
  "hostname": "server01.example.com",
  "cpu_type": "Intel Xeon E5-2680",
  "cpu_logic": 32,
  "memsum": 34359738368,
  "memory": "32 GB",
  "os": "Linux x86_64",
  ...
}
```

**Response**
```json
{
  "id": "server01.example.com",
  "status": "registered"
}
```

### 2. Agent Heartbeat

Update agent status and system information.

**Request**
```
POST /agents/heartbeat
Content-Type: application/json

{
  "hostname": "server01.example.com",
  ...
}
```

**Response**
```json
{
  "status": "ok"
}
```

### 3. Get Pending Tasks

Retrieve pending tasks for an agent.

**Request**
```
GET /agents/tasks
Authorization: Bearer <token>
```

**Response**
```json
[
  {
    "id": "task-001",
    "type": "command",
    "command": "echo 'Hello World'",
    "timeout": 30
  },
  {
    "id": "task-002",
    "type": "hook",
    "plugin": "custom-script",
    "params": {"arg1": "value1"}
  }
]
```

### 4. Submit Task Result

Submit task execution result.

**Request**
```
POST /agents/tasks/result
Content-Type: application/json

{
  "task_id": "task-001",
  "success": true,
  "output": "Hello World\n"
}
```

**Response**
```json
{
  "status": "received"
}
```

### 5. List Agents

List all registered agents.

**Request**
```
GET /agents/list
```

**Response**
```json
[
  {
    "id": "server01.example.com",
    "hostname": "server01.example.com",
    "cpu_type": "Intel Xeon E5-2680",
    "status": "online",
    "last_seen": "2024-01-01T12:00:00Z"
  },
  ...
]
```

### 6. Get Agent

Get specific agent details.

**Request**
```
GET /agents/:id
```

**Response**
```json
{
  "id": "server01.example.com",
  "hostname": "server01.example.com",
  "cpu_info": {...},
  "memory_info": [...],
  "gpu_info": [...],
  "status": "online",
  "last_seen": "2024-01-01T12:00:00Z"
}
```

### 7. Health Check

Check server health.

**Request**
```
GET /health
```

**Response**
```json
{
  "status": "ok"
}
```

### 8. Installation Script

Get agent installation script.

**Request**
```
GET /install.sh?token=<token>
```

**Response**
```bash
#!/bin/bash
# Installation script...
```

### 9. Download Agent Binary

Download agent binary.

**Request**
```
GET /download?token=<token>
```

**Response**
Binary file (application/octet-stream)

## Error Responses

### 400 Bad Request
```json
{
  "error": "invalid request"
}
```

### 401 Unauthorized
```json
{
  "error": "invalid token"
}
```

### 404 Not Found
```json
{
  "error": "agent not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

## Task Types

### Command Task
Execute a shell command.

```json
{
  "type": "command",
  "command": "df -h",
  "timeout": 30
}
```

### Script Task
Execute a script.

```json
{
  "type": "script",
  "script": "#!/bin/bash\necho 'Hello'",
  "timeout": 60
}
```

### Hook Task
Execute a hook plugin.

```json
{
  "type": "hook",
  "plugin": "custom-hook",
  "params": {
    "arg1": "value1"
  },
  "timeout": 120
}
```

## Usage Examples

### Register Agent (cURL)

```bash
curl -X POST http://nerve-center:8080/api/agents/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d @agent-info.json
```

### Submit Task Result (cURL)

```bash
curl -X POST http://nerve-center:8080/api/agents/tasks/result \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "task_id": "task-001",
    "success": true,
    "output": "Command executed successfully"
  }'
```

### List Agents (cURL)

```bash
curl http://nerve-center:8080/api/agents/list \
  -H "Authorization: Bearer your-token"
```

### Install Agent (Shell)

```bash
curl -fsSL http://nerve-center:8080/install.sh | \
  sh -s -- --token=your-token --server=http://nerve-center:8080
```

