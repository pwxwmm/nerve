# Nerve Deployment Guide

## Prerequisites

- Go 1.21+ installed
- Linux machines for agents
- Server machine with public IP

## Building

### Build Agent

```bash
cd agent
go build -o nerve-agent
```

### Build Server

```bash
cd server
go build -o nerve-center
```

## Server Deployment

### Option 1: Direct Deployment

```bash
# Copy binary
scp nerve-center user@server:/usr/local/bin/

# Start server
nerve-center --addr :8090 --debug

# Or with systemd
cp deploy/nerve-agent.service /etc/systemd/system/
systemctl enable --now nerve-center
```

### Option 2: Docker

```bash
# Build image
docker build -t nerve-center:latest .

# Run
docker-compose up -d
```

## Agent Deployment

### Simple Installation

On any Linux machine:

```bash
curl -fsSL https://your-server:8090/install.sh | sh -s -- \
  --token=YOUR_TOKEN \
  --server=https://your-server:8090
```

### Manual Installation

```bash
# Download agent binary
wget https://your-server:8090/download -O /usr/local/bin/nerve-agent
chmod +x /usr/local/bin/nerve-agent

# Create systemd service
cat > /etc/systemd/system/nerve-agent.service <<'EOF'
[Unit]
Description=Nerve Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nerve-agent --server=https://your-server:8090 --token=YOUR_TOKEN
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Start service
systemctl daemon-reload
systemctl enable --now nerve-agent
```

## Configuration

### Agent Configuration

Edit `/etc/nerve-agent/config.yaml`:

```yaml
server:
  url: "https://nerve-center:8090"
  timeout: 30s

auth:
  token: "your-token-here"

heartbeat:
  interval: 30s

collection:
  cpu: true
  memory: true
  disk: true
  network: true
  gpu: true
```

### Server Configuration

Edit `/etc/nerve-center/server.yaml`:

```yaml
server:
  addr: ":8090"
  
auth:
  method: token
  token_secret: "change-this"

registry:
  cleanup_interval: 1m
  offline_threshold: 5m

storage:
  type: postgres
  postgres:
    host: "localhost"
    database: "nerve"
```

## Verification

### Check Agent Status

```bash
# Agent logs
journalctl -u nerve-agent -f

# Agent info
systemctl status nerve-agent
```

### Check Server Status

```bash
# Server logs
journalctl -u nerve-center -f

# List agents
curl http://localhost:8090/api/agents/list

# Health check
curl http://localhost:8090/health
```

## Scaling

### Multi-Server Setup

Deploy multiple server instances with load balancer:

```nginx
upstream nerve-backend {
    server nerve1:8090;
    server nerve2:8090;
    server nerve3:8090;
}

server {
    listen 443 ssl;
    
    location / {
        proxy_pass http://nerve-backend;
    }
}
```

### Database

Use PostgreSQL for persistent storage:

```bash
# Create database
createdb nerve

# Update config
storage:
  type: postgres
  postgres:
    host: "localhost"
    database: "nerve"
    user: "nerve"
    password: "secure-password"
```

## Troubleshooting

### Agent Not Connecting

1. Check network connectivity
2. Verify token is correct
3. Check firewall rules
4. Review server logs

### Agent Stuck

```bash
# Restart agent
systemctl restart nerve-agent

# Check logs
journalctl -u nerve-agent --since "5 minutes ago"
```

### Server High Load

1. Increase agent heartbeat interval
2. Use PostgreSQL for storage
3. Add more server instances
4. Enable Redis caching

## Security Best Practices

1. **Use HTTPS** - Always use TLS encryption
2. **Rotate Tokens** - Regularly rotate authentication tokens
3. **Firewall Rules** - Restrict server access to internal networks
4. **Binary Verification** - Verify agent binary signatures
5. **Audit Logging** - Enable audit logs for compliance

## Performance Tuning

### Agent

```yaml
heartbeat:
  interval: 60s  # Increase for less frequent updates
  
collection:
  gpu: false  # Disable if no GPUs
  ipmi: false  # Disable if no IPMI
```

### Server

```yaml
server:
  read_timeout: 30s
  write_timeout: 30s
  
scheduler:
  max_concurrent_tasks: 500
```

## Monitoring

### Metrics Endpoint

```
curl http://localhost:8090/metrics
```

### Agent List

```bash
curl http://localhost:8090/api/agents/list | jq
```

### Agent Details

```bash
curl http://localhost:8090/api/agents/:hostname | jq
```

