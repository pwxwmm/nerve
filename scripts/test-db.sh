#!/bin/bash
# Test Database Connections

set -e

MONGO_URI="mongodb://root:c2B2h15D1PfHTOEjZd@172.29.228.139/nerve?authSource=admin"
REDIS_HOST="172.29.228.139"
REDIS_PORT="6379"
REDIS_PASSWORD="c2B2h15D1PfHTOEjZd"

echo "Testing Nerve Database Connections..."
echo ""

# Test MongoDB
echo "1. Testing MongoDB..."

# 检查 mongosh 命令是否存在
if ! command -v mongosh &> /dev/null; then
    echo "   ✗ mongosh command not found"
    echo "   💡 Install: yum install mongodb-mongosh-shell or brew install mongosh"
    
    # 尝试使用旧的 mongo 命令
    if command -v mongo &> /dev/null; then
        echo "   → Trying mongo command instead..."
        if mongo "$MONGO_URI" --quiet --eval "db.stats()" > /dev/null 2>&1; then
            echo "   ✓ MongoDB connected (using mongo)"
            COLLECTIONS=$(mongo "$MONGO_URI" --quiet --eval "printjson(db.getCollectionNames())")
            echo "   Collections: $COLLECTIONS"
        else
            echo "   ✗ MongoDB connection failed"
        fi
    fi
else
    # 使用 mongosh 命令并显示详细错误
    if mongosh "$MONGO_URI" --quiet --eval "db.stats()" 2>&1 | grep -v "^$"; then
        echo "   ✓ MongoDB connected"
        
        # Check collections
        COLLECTIONS=$(mongosh "$MONGO_URI" --quiet --eval "db.getCollectionNames()" 2>&1)
        echo "   Collections: $COLLECTIONS"
    else
        echo "   ✗ MongoDB connection failed"
        echo "   尝试手动连接测试: mongosh \"$MONGO_URI\""
    fi
fi

echo ""

# Test Redis
echo "2. Testing Redis..."

# 检查 redis-cli 命令是否存在
if ! command -v redis-cli &> /dev/null; then
    echo "   ✗ redis-cli command not found"
    echo "   💡 Install: yum install redis or apt-get install redis-tools"
else
    # 测试连接并显示详细信息
    REDIS_TEST=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" ping 2>&1)
    
    if [ "$REDIS_TEST" == "PONG" ]; then
        echo "   ✓ Redis connected"
        
        # Test set/get
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" set "nerve:test" "ok" > /dev/null 2>&1
        VALUE=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" get "nerve:test" 2>/dev/null || echo "none")
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" del "nerve:test" > /dev/null 2>&1
        
        if [ "$VALUE" == "ok" ]; then
            echo "   ✓ Redis read/write test passed"
        else
            echo "   ⚠ Redis read/write test failed"
        fi
    else
        echo "   ✗ Redis connection failed"
        echo "   错误信息: $REDIS_TEST"
        echo "   尝试手动测试: redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping"
    fi
fi

echo ""
echo "提示：如果连接失败，请检查："
echo "  1. 网络连接: ping 172.29.228.139"
echo "  2. 端口开放: telnet 172.29.228.139 27017 / 6379"
echo "  3. 认证信息是否正确"
echo ""

echo ""

