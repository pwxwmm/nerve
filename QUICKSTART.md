# Nerve 快速开始指南

## 🎯 当前配置状态

✅ **MongoDB**: 172.29.228.139 (已配置)  
✅ **Redis**: 172.29.228.139 (已配置)  
✅ **配置文件**: `server/config/server.yaml`

## 🚀 立即开始（3步完成）

### 第 1 步：初始化数据库

```bash
# 自动创建数据库、集合和索引
./scripts/init-db.sh
```

这会创建：
- ✅ `nerve` 数据库
- ✅ 5 个集合（agents, heartbeats, tasks, system_info, hooks）
- ✅ 所有必要的索引
- ✅ TTL 自动清理（7天心跳，30天系统信息）

### 第 2 步：构建并启动

```bash
# 安装依赖
go mod download

# 构建 Server
cd server && go build -o nerve-center && cd ..

# 启动 Server
./server/nerve-center --addr :8090 --debug
```

### 第 3 步：验证

在新终端测试：

```bash
curl http://localhost:8090/health
# 应该返回: {"status":"ok"}
```

## 📦 数据库信息

**MongoDB:**
```bash
Host: 172.29.228.139
Database: nerve
URI: mongodb://root:password@172.29.228.139/nerve?authSource=admin
```

**Redis:**
```bash
Host: 172.29.228.139:6379
Database: 0
```

## 🔧 测试连接

```bash
# 测试 MongoDB
mongosh "mongodb://root:c2B2h15D1PfHTOEjZd@172.29.228.139/nerve?authSource=admin" \
  --eval "db.stats()"

# 测试 Redis
redis-cli -h 172.29.228.139 -p 6379 -a c2B2h15D1PfHTOEjZd ping
```

## 📦 生产部署

### Server 部署

```bash
# 1. 复制二进制文件
sudo cp nerve-center /usr/local/bin/

# 2. 创建配置文件
sudo mkdir -p /etc/nerve
sudo cp server/config/server.yaml /etc/nerve/

# 3. 创建 systemd 服务
sudo tee /etc/systemd/system/nerve-center.service <<'EOF'
[Unit]
Description=Nerve Center
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nerve-center --addr :8090
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 4. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable --now nerve-center
```

### Agent 安装

#### 一键安装

```bash
curl -fsSL http://your-server:8090/install.sh | \
  sh -s -- --token=YOUR_TOKEN --server=http://your-server:8090
```

#### 手动安装

```bash
# 下载 Agent 二进制
wget http://your-server:8090/download?token=YOUR_TOKEN \
  -O /usr/local/bin/nerve-agent

chmod +x /usr/local/bin/nerve-agent

# 创建 systemd 服务
sudo tee /etc/systemd/system/nerve-agent.service <<'EOF'
[Unit]
Description=Nerve Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nerve-agent \
  --server=http://your-server:8090 \
  --token=YOUR_TOKEN
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable --now nerve-agent
```

## 🔧 测试 Agent 功能

### 1. 查看 Agent 日志

```bash
journalctl -u nerve-agent -f
```

### 2. 检查 Agent 状态

```bash
systemctl status nerve-agent
```

### 3. 查看 Agent 信息

```bash
curl http://localhost:8090/api/agents/list | jq
```

## 🎯 部署到多台机器

### 使用并行部署工具

```bash
# 准备主机列表
cat > hosts.txt <<EOF
host1.example.com
host2.example.com
host3.example.com
EOF

# 并行安装 Agent
parallel-ssh -h hosts.txt \
  'curl -fsSL http://nerve-center:8090/install.sh | \
   sh -s -- --token=YOUR_TOKEN --server=http://nerve-center:8090'
```

### 使用 Ansible

```yaml
- name: Install Nerve Agent
  hosts: all
  tasks:
    - name: Download and install agent
      shell: |
        curl -fsSL http://nerve-center:8090/install.sh | \
          sh -s -- \
          --token={{ nerve_token }} \
          --server=http://nerve-center:8090
```

## 📊 监控和告警

### 查看 Server 指标

```bash
curl http://localhost:8090/metrics
```

### 设置告警规则

```bash
# 检查离线 Agent
curl -s http://localhost:8090/api/agents/list | \
  jq '.[] | select(.status == "offline")'
```

## 🐛 故障排除

### Agent 无法连接

```bash
# 1. 检查网络连接
ping your-server

# 2. 检查防火墙
curl http://your-server:8090/health

# 3. 查看 Agent 日志
journalctl -u nerve-agent --no-pager | tail -20
```

### Server 高负载

```bash
# 1. 检查连接数
ss -an | grep 8080 | wc -l

# 2. 增加心跳间隔（Agent 配置）
# 编辑 /etc/nerve-agent/config.yaml
# heartbeat.interval: 60s
```

## 🔐 安全配置

### 启用 HTTPS

```bash
# 使用反向代理（Nginx）
sudo apt-get install nginx certbot

# 配置 SSL
sudo certbot --nginx -d your-domain.com
```

### 修改认证 Token

```bash
# Server 配置
vim /etc/nerve/server.yaml
# auth.token_secret: "new-secret"

# 重启服务
sudo systemctl restart nerve-center
```

## 📈 下一步

- 阅读 [架构文档](docs/ARCHITECTURE.md)
- 查看 [API 文档](docs/API.md)
- 探索 [Hook 插件系统](docs/HOOK_PLUGIN.md)
- 了解 [部署指南](docs/DEPLOYMENT.md)

## 💡 提示

1. **测试环境**: 先在少量机器上测试
2. **监控**: 关注 Server 的 CPU 和内存使用
3. **备份**: 定期备份 Agent 配置
4. **更新**: 使用滚动更新策略
5. **日志**: 保留日志至少 7 天

## 🆘 获取帮助

- GitHub Issues: https://github.com/nerve/nerve/issues
- 文档: https://github.com/nerve/nerve/tree/main/docs
- 社区: https://github.com/nerve/nerve/discussions

