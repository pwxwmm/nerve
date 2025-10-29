# Nerve 安全功能使用指南

## 🔒 安全功能概览

Nerve 现在包含完整的企业级安全功能：

- ✅ **TLS/HTTPS 加密通信**
- ✅ **Token 轮换机制**
- ✅ **审计日志系统**
- ✅ **细粒度权限控制**

## 🚀 快速开始

### 1. 启动安全模式 Server

```bash
# 启用 TLS/HTTPS
./server/nerve-center --tls --cert server.crt --key server.key --addr :8443

# 启用审计日志
./server/nerve-center --audit-log audit.log --tls

# 完整安全配置
./server/nerve-center \
  --tls \
  --cert server.crt \
  --key server.key \
  --audit-log audit.log \
  --addr :8443 \
  --debug
```

### 2. 访问安全服务

- **HTTPS Server**: https://localhost:8443
- **Web UI**: https://localhost:8443/web/
- **API**: https://localhost:8443/api/

## 🔐 TLS/HTTPS 配置

### 自动生成证书（开发环境）

```bash
# Server 会自动生成自签名证书
./server/nerve-center --tls
```

### 使用现有证书（生产环境）

```bash
# 使用 Let's Encrypt 或其他 CA 证书
./server/nerve-center --tls --cert /path/to/cert.pem --key /path/to/key.pem
```

### 证书文件格式

```bash
# 证书文件 (server.crt)
-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKoK/Ovj8F5TMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
...
-----END CERTIFICATE-----

# 私钥文件 (server.key)
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC7VJTUt9Us8cKB
...
-----END PRIVATE KEY-----
```

## 🎫 Token 管理

### 生成 Token

```bash
curl -X POST https://localhost:8443/api/tokens/generate \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-001",
    "permissions": ["read", "execute"]
  }'
```

### Token 轮换

```bash
curl -X POST https://localhost:8443/api/tokens/rotate \
  -H "Content-Type: application/json" \
  -d '{
    "old_token": "old-token-here"
  }'
```

### Token 验证

```bash
curl -H "Authorization: Bearer your-token" \
  https://localhost:8443/api/agents/list
```

## 👥 权限管理

### 默认角色

1. **admin** - 完全系统访问
2. **operator** - 系统操作权限
3. **agent** - Agent 操作权限
4. **viewer** - 只读权限

### 创建自定义角色

```bash
curl -X POST https://localhost:8443/api/roles \
  -H "Content-Type: application/json" \
  -d '{
    "id": "custom-role",
    "name": "Custom Role",
    "description": "Custom role description",
    "permissions": [
      {
        "resource": "agents",
        "actions": ["read", "update"]
      },
      {
        "resource": "tasks",
        "actions": ["read", "create"]
      }
    ]
  }'
```

### 创建用户

```bash
curl -X POST https://localhost:8443/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-001",
    "username": "john.doe",
    "email": "john@example.com",
    "roles": ["operator", "viewer"],
    "is_active": true
  }'
```

### 权限检查

```bash
# 检查用户权限
curl -H "Authorization: Bearer user-token" \
  https://localhost:8443/api/agents/list
```

## 📝 审计日志

### 查看审计日志

```bash
curl https://localhost:8443/api/audit/logs?limit=50
```

### 审计事件类型

1. **authentication** - 认证事件
2. **authorization** - 授权事件
3. **data_access** - 数据访问事件
4. **task_execution** - 任务执行事件
5. **configuration_change** - 配置变更事件
6. **api_request** - API 请求事件

### 审计日志格式

```json
{
  "timestamp": "2025-01-28T10:30:00Z",
  "event_type": "api_request",
  "user_id": "user-001",
  "agent_id": "agent-001",
  "ip_address": "192.168.1.100",
  "user_agent": "curl/7.68.0",
  "action": "GET",
  "resource": "/api/agents/list",
  "result": "200",
  "details": {
    "duration_ms": 45,
    "request_size": 0,
    "response_size": 1024
  }
}
```

## 🔧 配置示例

### 生产环境配置

```bash
#!/bin/bash
# 生产环境启动脚本

# 设置环境变量
export NERVE_TLS_ENABLED=true
export NERVE_CERT_FILE=/etc/ssl/certs/nerve.crt
export NERVE_KEY_FILE=/etc/ssl/private/nerve.key
export NERVE_AUDIT_LOG=/var/log/nerve/audit.log
export NERVE_ADDR=:8443

# 启动服务
./server/nerve-center \
  --tls \
  --cert $NERVE_CERT_FILE \
  --key $NERVE_KEY_FILE \
  --audit-log $NERVE_AUDIT_LOG \
  --addr $NERVE_ADDR
```

### Docker 配置

```yaml
version: '3.8'
services:
  nerve-server:
    image: nerve-server:latest
    ports:
      - "8443:8443"
    environment:
      - NERVE_TLS_ENABLED=true
      - NERVE_CERT_FILE=/certs/server.crt
      - NERVE_KEY_FILE=/certs/server.key
      - NERVE_AUDIT_LOG=/logs/audit.log
    volumes:
      - ./certs:/certs
      - ./logs:/logs
    command: [
      "--tls",
      "--cert", "/certs/server.crt",
      "--key", "/certs/server.key",
      "--audit-log", "/logs/audit.log",
      "--addr", ":8443"
    ]
```

## 🛡️ 安全最佳实践

### 1. 证书管理

- 使用有效的 CA 证书（Let's Encrypt、DigiCert 等）
- 定期更新证书
- 使用强加密算法（TLS 1.2+）

### 2. Token 安全

- 定期轮换 Token
- 使用强随机 Token
- 设置合理的过期时间
- 监控 Token 使用情况

### 3. 权限控制

- 遵循最小权限原则
- 定期审查用户权限
- 使用角色基础访问控制（RBAC）
- 监控权限变更

### 4. 审计日志

- 定期备份审计日志
- 监控异常活动
- 设置日志保留策略
- 使用日志分析工具

## 🔍 监控和告警

### Prometheus 指标

```promql
# TLS 连接数
nerve_tls_connections_total

# Token 使用情况
nerve_token_usage_total

# 权限检查失败
nerve_permission_denied_total

# 审计事件数
nerve_audit_events_total
```

### 告警规则

```yaml
groups:
  - name: nerve_security_alerts
    rules:
      - alert: HighPermissionDeniedRate
        expr: rate(nerve_permission_denied_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High permission denied rate"

      - alert: TokenExpirationWarning
        expr: nerve_token_expires_in_hours < 24
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Token expires soon"
```

## 🚨 故障排除

### 常见问题

1. **TLS 证书错误**
   ```bash
   # 检查证书有效性
   openssl x509 -in server.crt -text -noout
   
   # 检查私钥
   openssl rsa -in server.key -check
   ```

2. **Token 验证失败**
   ```bash
   # 检查 Token 格式
   echo "your-token" | base64 -d
   
   # 验证 Token 有效性
   curl -H "Authorization: Bearer your-token" \
     https://localhost:8443/api/auth/validate
   ```

3. **权限检查失败**
   ```bash
   # 检查用户角色
   curl https://localhost:8443/api/users/user-id
   
   # 检查角色权限
   curl https://localhost:8443/api/roles/role-id
   ```

## 📞 支持

- **文档**: 查看 `docs/` 目录
- **问题**: GitHub Issues
- **安全报告**: security@nerve.example.com

---

**Nerve** - 安全的基础设施管理平台 🔒
