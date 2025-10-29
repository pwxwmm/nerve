# Nerve 数据库设计与使用

## 📊 使用场景

Nerve 需要存储的数据类型：

1. **Agent 注册信息**
   - 机器基本属性（hostname, IP, cluster）
   - 硬件信息（CPU, Memory, GPU, Disk, Network）
   - 注册时间和状态

2. **心跳数据**
   - 每台机器 30 秒一次心跳
   - 6000 台机器 = 每分钟 12,000 次写入
   - 需要快速写入，查询较少

3. **系统信息历史**
   - CPU、内存、GPU 使用率变化
   - 磁盘空间变化
   - 网络流量统计
   - 需要保存一段时间用于趋势分析

4. **任务执行记录**
   - 任务类型、参数、结果
   - 执行时间、耗时
   - 错误信息

5. **Hook 插件数据**
   - 插件配置
   - 执行结果
   - 自定义字段

## 🆚 数据库对比

### MongoDB

**优势：**
- ✅ 文档型，直接存储 JSON（无需序列化）
- ✅ 灵活的 Schema（新增字段自动适应）
- ✅ 水平扩展能力强（分片）
- ✅ 写入性能高（适合大批量心跳）
- ✅ 适合嵌套数据（GPU 数组、磁盘数组）
- ✅ TTL 索引自动清理过期数据
- ✅ 时间序列支持（未来扩展）

**劣势：**
- ❌ 需要学习 NoSQL 查询语法
- ❌ 事务支持较弱（4.0+ 后改善）
- ❌ 运维复杂度稍高

**适用场景：**
```javascript
// Agent 信息直接存 JSON
{
  hostname: "server01",
  cpu_info: { model: "Intel Xeon", cores: 32 },
  gpu_info: [{ type: "NVIDIA A100", memory: "80GB" }],
  disk_info: [{ device: "/dev/sda", capacity: "2TB" }],
  heartbeat_at: ISODate("2024-01-01T12:00:00Z")
}
```

### PostgreSQL

**优势：**
- ✅ 成熟稳定，生态丰富
- ✅ ACID 事务保证
- ✅ SQL 查询灵活强大
- ✅ JSONB 支持半结构化数据
- ✅ 关系查询方便（Agent ↔ Cluster）
- ✅ 运维工具成熟
- ✅ 全文搜索支持

**劣势：**
- ❌ JSONB 性能不如 MongoDB 原生文档
- ❌ 需要设计表结构（灵活性稍低）
- ❌ 水平扩展需要 sharding
- ❌ 高频写入需要优化

**适用场景：**
```sql
-- 表结构设计
CREATE TABLE agents (
  id SERIAL PRIMARY KEY,
  hostname VARCHAR(255),
  cpu_info JSONB,
  gpu_info JSONB[],
  disk_info JSONB[],
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- 查询示例
SELECT hostname, cpu_info->>'model' as cpu_model
FROM agents
WHERE gpu_info @> '[{"type": "NVIDIA"}]';
```

### MySQL

**优势：**
- ✅ 使用最广泛，社区活跃
- ✅ 运维经验丰富
- ✅ SQL 标准支持好

**劣势：**
- ❌ 对 JSON 支持较弱（5.7+ 才支持）
- ❌ 性能不如 PostgreSQL
- ❌ 不适合半结构化数据

**结论：** 不太适合 Nerve

## 🎯 推荐方案

### 方案 1：MongoDB（推荐）

**理由：**
1. **数据结构匹配** - 系统信息采集就是 JSON
2. **写入性能** - 心跳数据高频写入，MongoDB 优化到位
3. **灵活扩展** - 新增字段无需改表
4. **水平扩展** - 6000+ 机器场景，MongoDB 分片简单

**实现示例：**
```go
// agent_doc.go
type Agent struct {
    ID         primitive.ObjectID `bson:"_id"`
    Hostname   string            `bson:"hostname"`
    CPUInfo    map[string]interface{} `bson:"cpu_info"`
    GPUInfo    []map[string]interface{} `bson:"gpu_info"`
    Heartbeat  time.Time         `bson:"heartbeat_at"`
    Created    time.Time         `bson:"created_at"`
}
```

### 方案 2：PostgreSQL + JSONB（平衡之选）

**理由：**
1. **团队熟悉** - 如果团队更熟悉 SQL
2. **事务需求** - 如果需要复杂事务
3. **关系查询** - 需要 Agent、Cluster、Task 关联查询

**实现示例：**
```sql
CREATE TABLE agents (
    id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) UNIQUE,
    system_info JSONB,
    cluster_id INT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 索引优化
CREATE INDEX idx_system_info ON agents USING GIN (system_info);
CREATE INDEX idx_heartbeat ON agents (updated_at);
```

### 方案 3：混合方案（最佳实践）

**推荐架构：**

```
┌─────────────────────────────────────────┐
│              Nerve Center               │
├─────────────────────────────────────────┤
│  Hot Data          │  Cold Data         │
│  ──────────        │  ──────────        │
│  Redis/Memory      │  MongoDB           │
│  - 在线 Agent       │  - Agent 历史     │
│  - 心跳数据         │  - 心跳历史        │
│  - 任务状态         │  - 任务历史        │
│                    │  - 系统信息历史     │
└─────────────────────────────────────────┘
```

**数据流动：**
1. **实时数据** → Redis（在线 Agent 列表、任务队列）
2. **热数据** → MongoDB（最近 7 天的心跳和系统信息）
3. **冷数据** → PostgreSQL 或归档存储（长期历史数据）

## 📊 写入性能对比

以 6000 台机器，30 秒心跳为例：

### 写入频率
- 每分钟：12,000 条记录
- 每小时：720,000 条记录
- 每天：17,280,000 条记录

### 存储需求（估算）

单条心跳记录：
```json
{
  "hostname": "server01",
  "timestamp": "2024-01-01T12:00:00Z",
  "cpu_usage": 45.2,
  "memory_usage": 67.8,
  "disk_free": 1234567890,
  "network_rx": 1024000,
  "network_tx": 2048000
}
```
大小约：500 字节

### 存储计算

**一天数据：**
- 6000 台 × 2880 心跳/天 × 500 字节
- ≈ 8.64 GB/天

**一个月数据：**
- ≈ 259 GB/月

**建议存储策略：**
- Redis/Memory: 最近 1 小时（热数据）
- MongoDB: 最近 7 天（温数据）
- 压缩归档: 1 个月以上（冷数据）

## 🔧 实现建议

### MongoDB 实现

```go
// storage/mongodb.go
type MongoDBStorage struct {
    client *mongo.Client
    db     *mongo.Database
}

func (m *MongoDBStorage) SaveAgent(agent *core.AgentInfo) error {
    ctx := context.TODO()
    _, err := m.db.Collection("agents").UpdateOne(
        ctx,
        bson.M{"hostname": agent.Hostname},
        bson.M{"$set": agent, "$setOnInsert": bson.M{"created_at": time.Now()}},
        options.Update().SetUpsert(true),
    )
    return err
}

func (m *MongoDBStorage) SaveHeartbeat(agentID string, heartbeat *Heartbeat) error {
    ctx := context.TODO()
    _, err := m.db.Collection("heartbeats").InsertOne(ctx, bson.M{
        "agent_id": agentID,
        "timestamp": heartbeat.Timestamp,
        "metrics": heartbeat.Metrics,
    })
    return err
}
```

### PostgreSQL 实现

```go
// storage/postgres.go
type PostgresStorage struct {
    db *sql.DB
}

func (p *PostgresStorage) SaveAgent(agent *core.AgentInfo) error {
    query := `
        INSERT INTO agents (hostname, system_info, updated_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (hostname) 
        DO UPDATE SET system_info = EXCLUDED.system_info, updated_at = EXCLUDED.updated_at
    `
    _, err := p.db.Exec(query, agent.Hostname, agent.SystemInfo, time.Now())
    return err
}
```

## 🎯 最终推荐

**生产环境推荐：**

1. **主要数据库**：MongoDB
   - 存储 Agent 信息和心跳数据
   - TTL 自动清理（保留 7-30 天）

2. **缓存层**：Redis
   - 在线 Agent 列表
   - 任务队列
   - 热点数据

3. **长期归档**：PostgreSQL（可选）
   - 如果需要关系查询
   - 如果需要复杂分析

**理由总结：**
- MongoDB：性能和灵活性最好
- PostgreSQL：团队熟悉或有复杂查询需求时可用
- MySQL：不建议，对 JSON 支持弱
- Redis：必选，做缓存和实时数据

## 🚀 实际配置（已配置）

**您的 Nerve 系统已连接到：**

- **MongoDB**: `mongodb://root:password@172.29.228.139/nerve?authSource=admin`
- **Redis**: `172.29.228.139:6379`

**配置文件**: `server/config/server.yaml`

### 初始化数据库

```bash
./scripts/init-db.sh
```

### 测试连接

```bash
mongosh "mongodb://root:c2B2h15D1PfHTOEjZd@172.29.228.139/nerve?authSource=admin" \
  --eval "db.stats()"

redis-cli -h 172.29.228.139 -p 6379 -a c2B2h15D1PfHTOEjZd ping
```

## 📚 参考资料

- [MongoDB 性能基准测试](https://www.mongodb.com/compare/mongodb-performance)
- [PostgreSQL JSONB vs MongoDB](https://www.postgresql.org/docs/current/datatype-json.html)
- [数据库选择指南](https://www.mongodb.com/compare)

