/**
 * Nerve MongoDB Initialization Script
 *
 * Author: mmwei3 (2025-10-28)
 * Wethers: cloudWays
 */
// Run with: mongosh <this-file> OR mongosh --file this-file

// Switch to nerve database
use nerve

print("Creating collections...")

// Create collections
db.createCollection("agents")
db.createCollection("heartbeats")
db.createCollection("tasks")
db.createCollection("system_info")
db.createCollection("hooks")

print("Creating indexes for agents...")
db.agents.createIndex({ hostname: 1 }, { unique: true })
db.agents.createIndex({ cluster: 1 })
db.agents.createIndex({ status: 1 })
db.agents.createIndex({ last_seen: 1 })
db.agents.createIndex({ updated_at: -1 })

print("Creating indexes for heartbeats...")
db.heartbeats.createIndex({ agent_id: 1, timestamp: -1 })
db.heartbeats.createIndex({ timestamp: 1 }, { expireAfterSeconds: 7 * 24 * 3600 }) // 7 days TTL
db.heartbeats.createIndex({ hostname: 1, timestamp: -1 })

print("Creating indexes for tasks...")
db.tasks.createIndex({ task_id: 1 }, { unique: true })
db.tasks.createIndex({ agent_id: 1, status: 1 })
db.tasks.createIndex({ created_at: -1 })
db.tasks.createIndex({ type: 1 })

print("Creating indexes for system_info...")
db.system_info.createIndex({ hostname: 1, timestamp: -1 })
db.system_info.createIndex({ timestamp: 1 }, { expireAfterSeconds: 30 * 24 * 3600 }) // 30 days TTL

print("Creating indexes for hooks...")
db.hooks.createIndex({ name: 1 }, { unique: true })
db.hooks.createIndex({ enabled: 1 })
db.hooks.createIndex({ executed_at: -1 })

print("✓ All collections and indexes created successfully")
print("\nCollections:")
db.getCollectionNames().forEach(function(name) {
    print("  - " + name)
})

print("\nVerifying indexes:")
db.agents.getIndexes().forEach(function(index) {
    print("  " + JSON.stringify(index.key))
})

