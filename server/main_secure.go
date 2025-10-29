// Package main provides the Nerve Center Server entry point with security features.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package main

import (
	"context"
	"flag"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nerve/server/api"
	"github.com/nerve/server/core"
	"github.com/nerve/server/pkg/alert"
	"github.com/nerve/server/pkg/binary"
	"github.com/nerve/server/pkg/cluster"
	"github.com/nerve/server/pkg/log"
	"github.com/nerve/server/pkg/metrics"
	"github.com/nerve/server/pkg/security"
	"github.com/nerve/server/pkg/storage"
	"github.com/nerve/server/pkg/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr         = flag.String("addr", ":8090", "Server address")
	debug        = flag.Bool("debug", false, "Enable debug mode")
	metricsAddr  = flag.String("metrics-addr", "", "Metrics server address (empty to disable)")
	enableTLS    = flag.Bool("tls", false, "Enable TLS/HTTPS")
	certFile     = flag.String("cert", "server.crt", "TLS certificate file")
	keyFile      = flag.String("key", "server.key", "TLS private key file")
	auditLogFile = flag.String("audit-log", "audit.log", "Audit log file")
)

func main() {
	flag.Parse()

	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize security components
	tlsServer := security.NewTLSServer(*certFile, *keyFile)
	tokenManager := security.NewTokenManager(24*time.Hour, 7*24*time.Hour) // 24h rotation, 7d expiration
	auditLogger := security.NewAuditLogger(*auditLogFile)
	permManager := security.NewPermissionManager()

	// Setup TLS if enabled
	if *enableTLS {
		if err := tlsServer.SetupTLS(); err != nil {
			stdlog.Fatalf("Failed to setup TLS: %v", err)
		}
		fmt.Println("TLS/HTTPS enabled")
	}

	// Initialize logger
	logger := log.New(*debug)

	// Initialize storage and registry
	var store storage.Storage
	// For now, use in-memory storage
	store = storage.NewInMemory()
	
	// Create registry
	registry := core.NewRegistry(store, logger)

	// Initialize other components
	wsManager := websocket.NewWebSocketManager()
	clusterMgr := cluster.NewClusterManager()
	alertMgr := alert.NewAlertManager()
	metricsCollector := metrics.NewMetricsCollector()
	binaryMgr := binary.NewAgentBinaryManager("./binaries")

	// Start WebSocket manager
	go wsManager.Run()

	// Start metrics collector
	go startMetricsServer(metricsCollector)

	// Setup HTTP router
	router := gin.Default()

	// Add security middleware
	router.Use(security.AuditMiddleware(auditLogger))

	// Setup API routes with security
	apiRouter := api.NewAPIRouter(wsManager, clusterMgr, alertMgr, registry)
	apiRouter.SetupRoutes(router)

	// Setup security routes
	setupSecurityRoutes(router, tokenManager, permManager, auditLogger)

	// Setup metrics routes
	metricsHandler := api.NewMetricsHandler(metricsCollector)
	router.GET("/metrics", metricsHandler)

	// Setup binary routes
	binaryMgr.SetupBinaryRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:    *addr,
		Handler: router,
	}

	// Add TLS configuration if enabled
	if *enableTLS {
		srv.TLSConfig = tlsServer.GetTLSConfig()
	}

	// Start HTTP server
	go func() {
		var err error
		if *enableTLS {
			err = srv.ListenAndServeTLS(*certFile, *keyFile)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			stdlog.Fatalf("Server failed to start: %v", err)
		}
	}()

	protocol := "http"
	if *enableTLS {
		protocol = "https"
	}

	fmt.Printf("Nerve Center started at %s://localhost%s\n", protocol, *addr)
	if *metricsAddr != "" {
		fmt.Printf("Metrics endpoint: http://localhost%s/metrics\n", *metricsAddr)
	} else {
		fmt.Printf("Metrics endpoint: %s://localhost%s/metrics\n", protocol, *addr)
	}
	fmt.Printf("Web UI: %s://localhost%s/web/\n", protocol, *addr)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		stdlog.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting")
}

// setupSecurityRoutes sets up security-related routes
func setupSecurityRoutes(router *gin.Engine, tokenManager *security.TokenManager, permManager *security.PermissionManager, auditLogger *security.AuditLogger) {
	// Authentication routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			// TODO: Implement login logic
			c.JSON(http.StatusOK, gin.H{"token": "dummy-token"})
		})
		auth.POST("/logout", func(c *gin.Context) {
			// TODO: Implement logout logic
			c.JSON(http.StatusOK, gin.H{"message": "logged out"})
		})
	}

	// Token management routes
	tokens := router.Group("/api/tokens")
	{
		tokens.GET("/", func(c *gin.Context) {
			tokenList := tokenManager.ListTokens()
			c.JSON(http.StatusOK, gin.H{"tokens": tokenList})
		})
		tokens.POST("/generate", func(c *gin.Context) {
			var req struct {
				AgentID     string   `json:"agent_id"`
				Permissions []string `json:"permissions"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			token, err := tokenManager.GenerateToken(req.AgentID, req.Permissions)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"token": token})
		})
		tokens.POST("/rotate", func(c *gin.Context) {
			var req struct {
				OldToken string `json:"old_token"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			newToken, err := tokenManager.RotateToken(req.OldToken)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"new_token": newToken})
		})
	}

	// Role management routes
	roles := router.Group("/api/roles")
	{
		roles.GET("/", func(c *gin.Context) {
			roleList := permManager.ListRoles()
			c.JSON(http.StatusOK, gin.H{"roles": roleList})
		})
		roles.POST("/", func(c *gin.Context) {
			var role security.Role
			if err := c.ShouldBindJSON(&role); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := permManager.AddRole(&role); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "role created"})
		})
	}

	// User management routes
	users := router.Group("/api/users")
	{
		users.GET("/", func(c *gin.Context) {
			userList := permManager.ListUsers()
			c.JSON(http.StatusOK, gin.H{"users": userList})
		})
		users.POST("/", func(c *gin.Context) {
			var user security.User
			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := permManager.AddUser(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "user created"})
		})
	}

	// Audit log routes
	audit := router.Group("/api/audit")
	{
		audit.GET("/logs", func(c *gin.Context) {
			limit := 100
			if limitStr := c.Query("limit"); limitStr != "" {
				fmt.Sscanf(limitStr, "%d", &limit)
			}

			logs, err := auditLogger.GetAuditLogs(limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"logs": logs})
		})
	}
}

// startMetricsServer starts a separate metrics server
func startMetricsServer(collector *metrics.MetricsCollector) {
	// Skip if metrics address is empty
	if *metricsAddr == "" {
		stdlog.Println("Metrics server disabled (no address specified)")
		return
	}

	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	srv := &http.Server{
		Addr:         *metricsAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		stdlog.Printf("Metrics server failed to start: %v", err)
	}
}

