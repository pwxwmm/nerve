// Package storage provides MongoDB storage implementation for Nerve Center Server.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBStorage implements Storage using MongoDB
type MongoDBStorage struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDB creates a new MongoDB storage instance
func NewMongoDB(cfg MongoDBConfig) (*MongoDBStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(cfg.Database)

	// Create indexes
	createIndexes(db)

	return &MongoDBStorage{
		client:   client,
		database: db,
	}, nil
}

// createIndexes creates necessary indexes for optimal query performance
func createIndexes(db *mongo.Database) {
	ctx := context.Background()

	// Agents collection
	agentsCol := db.Collection("agents")
	agentsCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "hostname", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "cluster", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "last_seen", Value: 1}}},
	})

	// Heartbeats collection with TTL
	heartbeatsCol := db.Collection("heartbeats")
	heartbeatsCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "agent_id", Value: 1}, {Key: "timestamp", Value: -1}}},
		{Keys: bson.D{{Key: "timestamp", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(7 * 24 * 3600)}, // 7 days TTL
	})

	// Tasks collection
	tasksCol := db.Collection("tasks")
	tasksCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "agent_id", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
}

// Get retrieves a value from storage
func (m *MongoDBStorage) Get(key string) (interface{}, error) {
	ctx := context.Background()
	
	var result bson.M
	err := m.database.Collection("agents").FindOne(ctx, bson.M{"_id": key}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	return result, err
}

// Set stores a value in storage
func (m *MongoDBStorage) Set(key string, value interface{}) error {
	ctx := context.Background()
	
	_, err := m.database.Collection("data").UpdateOne(
		ctx,
		bson.M{"_id": key},
		bson.M{"$set": bson.M{"value": value, "updated_at": time.Now()}},
		options.Update().SetUpsert(true),
	)
	return err
}

// Delete removes a value from storage
func (m *MongoDBStorage) Delete(key string) error {
	ctx := context.Background()
	
	_, err := m.database.Collection("data").DeleteOne(ctx, bson.M{"_id": key})
	return err
}

// List returns all key-value pairs
func (m *MongoDBStorage) List() map[string]interface{} {
	ctx := context.Background()
	
	cursor, err := m.database.Collection("data").Find(ctx, bson.M{})
	if err != nil {
		return make(map[string]interface{})
	}
	defer cursor.Close(ctx)

	result := make(map[string]interface{})
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		if id, ok := doc["_id"].(string); ok {
			result[id] = doc["value"]
		}
	}
	
	return result
}

// SaveAgent saves agent information
func (m *MongoDBStorage) SaveAgent(agent interface{}) error {
	ctx := context.Background()
	
	_, err := m.database.Collection("agents").UpdateOne(
		ctx,
		bson.M{"hostname": getHostname(agent)},
		bson.M{
			"$set": agent,
			"$setOnInsert": bson.M{
				"created_at": time.Now(),
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

// SaveHeartbeat saves heartbeat data
func (m *MongoDBStorage) SaveHeartbeat(agentID string, heartbeat interface{}) error {
	ctx := context.Background()
	
	doc := bson.M{
		"agent_id":  agentID,
		"timestamp": time.Now(),
		"heartbeat": heartbeat,
	}
	
	_, err := m.database.Collection("heartbeats").InsertOne(ctx, doc)
	return err
}

// GetAgents retrieves all agents
func (m *MongoDBStorage) GetAgents(filter interface{}) ([]interface{}, error) {
	ctx := context.Background()
	
	cursor, err := m.database.Collection("agents").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []interface{}
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		results = append(results, doc)
	}
	
	return results, nil
}

// Close closes the MongoDB connection
func (m *MongoDBStorage) Close() error {
	ctx := context.Background()
	return m.client.Disconnect(ctx)
}

func getHostname(agent interface{}) string {
	// Helper function to extract hostname
	if m, ok := agent.(map[string]interface{}); ok {
		if hostname, ok := m["hostname"].(string); ok {
			return hostname
		}
	}
	return ""
}

