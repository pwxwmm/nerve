// Package main provides the Nerve Agent entry point.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nerve/agent/core"
	agentlog "github.com/nerve/agent/pkg/log"
)

var (
	serverURL = flag.String("server", "", "Server URL (e.g., https://nerve-center:8080)")
	token     = flag.String("token", "", "Authentication token")
	interval  = flag.Duration("interval", 30*time.Second, "Heartbeat interval")
	debug     = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()

	// Setup logger
	logger := agentlog.New(*debug)

	if *serverURL == "" {
		logger.Fatal("server URL is required (--server)")
	}
	if *token == "" {
		logger.Fatal("token is required (--token)")
	}

	logger.Infof("Starting Nerve Agent (Server: %s)", *serverURL)

	// Initialize core components
	agent := core.NewAgentWithLogger(*serverURL, *token, *interval, logger)

	// Initial registration
	if err := agent.Register(); err != nil {
		logger.Fatalf("Failed to register: %v", err)
	}
	logger.Info("Successfully registered with server")

	// Start heartbeat in background
	go agent.StartHeartbeat()

	// Start task listener
	go agent.StartTaskListener()

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
	agent.Stop()
}

