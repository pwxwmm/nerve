package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nerve/server/pkg/storage"
)

func main() {
	fmt.Println("Testing Nerve Database Connections...")
	fmt.Println()

	// Test MongoDB
	fmt.Println("1. Testing MongoDB...")
	mongoCfg := storage.MongoDBConfig{
		URI:      "mongodb://root:c2B2h15D1PfHTOEjZd@172.29.228.139/nerve?authSource=admin",
		Database: "nerve",
	}

	mongoStore, err := storage.NewMongoDB(mongoCfg)
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	defer mongoStore.Close()
	
	fmt.Println("   ✓ MongoDB connected successfully")
	
	// Test basic operations
	testAgent := map[string]interface{}{
		"hostname": "test-server",
		"cpu_type": "Intel Xeon",
		"status":   "online",
	}
	
	if err := mongoStore.SaveAgent(testAgent); err != nil {
		fmt.Printf("   ⚠ Save agent test failed: %v\n", err)
	} else {
		fmt.Println("   ✓ Save agent test passed")
	}
	
	fmt.Println()

	// Test Redis (optional)
	fmt.Println("2. Testing Redis...")
	// Note: Redis client would need to be implemented
	fmt.Println("   ⚠ Redis test skipped (not yet implemented)")
	fmt.Println()

	fmt.Println("All database tests completed!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Build the server: cd server && go build -o nerve-center")
	fmt.Println("  2. Start the server: ./nerve-center --config=../server/config/server.yaml --debug")
}

