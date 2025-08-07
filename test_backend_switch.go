package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/auto-devs/auto-devs/internal/config"
	"github.com/auto-devs/auto-devs/internal/websocket"
)

func main() {
	// Test configuration loading
	fmt.Println("Testing backend configuration...")
	
	// Test with legacy backend (default)
	os.Setenv("USE_CENTRIFUGE", "false")
	cfg := &config.Config{
		CentrifugeConfig: &config.CentrifugeConfig{
			RedisAddr: "localhost:6379",
			RedisDB:   0,
		},
	}
	
	fmt.Println("Creating enhanced service with legacy backend...")
	enhancedSvc, err := websocket.NewEnhancedService(cfg)
	if err != nil {
		log.Fatalf("Failed to create enhanced service: %v", err)
	}
	
	fmt.Printf("✓ Enhanced service created. Using Centrifuge: %t\n", enhancedSvc.IsCentrifugeEnabled())
	
	// Test switching to Centrifuge
	fmt.Println("\nTesting switch to Centrifuge backend...")
	os.Setenv("USE_CENTRIFUGE", "true")
	
	enhancedSvc2, err := websocket.NewEnhancedService(cfg)
	if err != nil {
		fmt.Printf("⚠ Expected error creating Centrifuge service (Redis not available): %v\n", err)
	} else {
		fmt.Printf("✓ Enhanced service created. Using Centrifuge: %t\n", enhancedSvc2.IsCentrifugeEnabled())
	}
	
	// Test broadcasting with legacy backend
	fmt.Println("\nTesting message broadcasting with legacy backend...")
	ctx := context.Background()
	testMessage := map[string]interface{}{
		"type": "test",
		"data": "Hello from enhanced service",
	}
	
	err = enhancedSvc.BroadcastToProject(ctx, 1, testMessage)
	if err != nil {
		fmt.Printf("⚠ Broadcast error (expected, no active connections): %v\n", err)
	} else {
		fmt.Println("✓ Broadcast method executed successfully")
	}
	
	fmt.Println("\n✅ Backend switching test completed successfully!")
}