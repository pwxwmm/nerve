#!/bin/bash
# Nerve Database Initialization Script

set -e

echo "========================================="
echo "  Nerve Database Initialization"
echo "========================================="

MONGO_URI="mongodb://root:c2B2h15D1PfHTOEjZd@172.29.228.139/nerve?authSource=admin"
REDIS_HOST="172.29.228.139"
REDIS_PORT="6379"
REDIS_PASSWORD="c2B2h15D1PfHTOEjZd"

# Test Redis connection
echo ""
echo "Testing Redis connection..."
if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" ping > /dev/null 2>&1; then
    echo "✓ Redis connection successful"
else
    echo "⚠ Redis connection failed (optional for caching)"
fi

# Create database and collections using Python
echo ""
echo "Creating Nerve database and collections using Python..."

python3 << 'PYTHON_SCRIPT'
import sys

try:
    from pymongo import MongoClient
    
    # 连接到 MongoDB
    client = MongoClient("mongodb://root:c2B2h15D1PfHTOEjZd@172.29.228.139/nerve?authSource=admin", 
                        serverSelectionTimeoutMS=5000)
    
    # 测试连接
    client.admin.command('ping')
    print("✓ MongoDB 连接成功")
    
    # 切换到 nerve 数据库
    db = client.nerve
    
    # 创建集合
    collections = ["agents", "heartbeats", "tasks", "system_info", "hooks"]
    existing = db.list_collection_names()
    
    print(f"\n创建集合...")
    for collection in collections:
        if collection not in existing:
            db.create_collection(collection)
            print(f"  ✓ {collection}")
        else:
            print(f"  ⊙ {collection} (已存在)")
    
    # 创建索引
    print(f"\n创建索引...")
    
    # Agents 索引
    try:
        db.agents.create_index("hostname", unique=True)
        print("  ✓ agents.hostname (unique)")
    except:
        pass
    
    try:
        db.agents.create_index("cluster")
        print("  ✓ agents.cluster")
    except:
        pass
    
    try:
        db.agents.create_index("status")
        print("  ✓ agents.status")
    except:
        pass
    
    try:
        db.agents.create_index("last_seen")
        print("  ✓ agents.last_seen")
    except:
        pass
    
    # Heartbeats 索引（带 TTL）
    try:
        db.heartbeats.create_index([("agent_id", 1), ("timestamp", -1)])
        print("  ✓ heartbeats.agent_id+timestamp")
    except:
        pass
    
    try:
        db.heartbeats.create_index("timestamp", expireAfterSeconds=7*24*3600)
        print("  ✓ heartbeats.timestamp (7 days TTL)")
    except:
        pass
    
    # Tasks 索引
    try:
        db.tasks.create_index("task_id", unique=True)
        print("  ✓ tasks.task_id (unique)")
    except:
        pass
    
    try:
        db.tasks.create_index([("agent_id", 1), ("status", 1)])
        print("  ✓ tasks.agent_id+status")
    except:
        pass
    
    try:
        db.tasks.create_index("created_at")
        print("  ✓ tasks.created_at")
    except:
        pass
    
    # System info 索引
    try:
        db.system_info.create_index([("hostname", 1), ("timestamp", -1)])
        print("  ✓ system_info.hostname+timestamp")
    except:
        pass
    
    try:
        db.system_info.create_index("timestamp", expireAfterSeconds=30*24*3600)
        print("  ✓ system_info.timestamp (30 days TTL)")
    except:
        pass
    
    # Hooks 索引
    try:
        db.hooks.create_index("name", unique=True)
        print("  ✓ hooks.name (unique)")
    except:
        pass
    
    # 验证
    print(f"\n✓ 数据库初始化完成")
    print(f"\n数据库信息:")
    print(f"  数据库: nerve")
    print(f"  集合数: {len(db.list_collection_names())}")
    
except ImportError:
    print("✗ pymongo 未安装")
    print("💡 安装命令: pip3 install pymongo")
    sys.exit(1)
except Exception as e:
    print(f"✗ 连接失败: {e}")
    print("💡 请检查:")
    print("  1. 网络连接: ping 172.29.228.139")
    print("  2. 端口开放: telnet 172.29.228.139 27017")
    print("  3. 认证信息是否正确")
    sys.exit(1)

PYTHON_SCRIPT

echo ""
echo "========================================="
echo "  ✓ Database initialization complete!"
echo "========================================="
echo ""
echo "Database: nerve"
echo "MongoDB: 172.29.228.139"
echo ""
echo "Next steps:"
echo "  1. Build the server: make build-server"
echo "  2. Start the server: ./nerve-center"
echo ""

