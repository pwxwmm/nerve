// Package api provides HTTP API routing and handlers for Nerve Center Server.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nerve/server/core"
	"github.com/nerve/server/pkg/alert"
	"github.com/nerve/server/pkg/cluster"
	"github.com/nerve/server/pkg/metrics"
	"github.com/nerve/server/pkg/websocket"
)

// APIRouter sets up all API routes
type APIRouter struct {
	wsManager     *websocket.WebSocketManager
	clusterMgr    *cluster.ClusterManager
	alertMgr      *alert.AlertManager
	registry      *core.Registry
}

// NewAPIRouter creates a new API router
func NewAPIRouter(wsManager *websocket.WebSocketManager, clusterMgr *cluster.ClusterManager, alertMgr *alert.AlertManager, registry *core.Registry) *APIRouter {
	return &APIRouter{
		wsManager:  wsManager,
		clusterMgr: clusterMgr,
		alertMgr:   alertMgr,
		registry:   registry,
	}
}

// SetupRoutes configures all API routes
func (r *APIRouter) SetupRoutes(router *gin.Engine) {
	// Web UI static files
	router.Static("/web", "../web")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/")
	})

	// WebSocket endpoint
	router.GET("/ws", r.wsManager.HandleWebSocket)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Agent routes
		agents := v1.Group("/agents")
		{
			agents.GET("/list", r.listAgents)
			agents.GET("/:id", r.getAgent)
			agents.POST("/:id/restart", r.restartAgent)
			agents.GET("/:id/tasks", r.getAgentTasks)
		}

		// Task routes
		tasks := v1.Group("/tasks")
		{
			tasks.GET("/list", r.listTasks)
			tasks.POST("/", r.createTask)
			tasks.GET("/:id", r.getTask)
			tasks.POST("/:id/cancel", r.cancelTask)
		}

		// Cluster routes
		clusters := v1.Group("/clusters")
		{
			clusters.GET("/list", r.listClusters)
			clusters.POST("/", r.createCluster)
			clusters.GET("/:id", r.getCluster)
			clusters.PUT("/:id", r.updateCluster)
			clusters.DELETE("/:id", r.deleteCluster)
			clusters.GET("/:id/stats", r.getClusterStats)
			clusters.POST("/:id/agents/:agent_id", r.addAgentToCluster)
			clusters.DELETE("/:id/agents/:agent_id", r.removeAgentFromCluster)
		}

		// Alert routes
		alerts := v1.Group("/alerts")
		{
			alerts.GET("/list", r.listAlerts)
			alerts.POST("/rules", r.createAlertRule)
			alerts.GET("/rules", r.listAlertRules)
			alerts.PUT("/rules/:id", r.updateAlertRule)
			alerts.DELETE("/rules/:id", r.deleteAlertRule)
			alerts.POST("/:id/resolve", r.resolveAlert)
		}

		// Plugin routes
		plugins := v1.Group("/plugins")
		{
			plugins.GET("/list", r.listPlugins)
			plugins.POST("/upload", r.uploadPlugin)
			plugins.DELETE("/:name", r.deletePlugin)
		}

		// System routes
		system := v1.Group("/system")
		{
			system.GET("/stats", r.getSystemStats)
			system.GET("/health", r.getHealth)
		}

		// Token management routes
		tokens := v1.Group("/tokens")
		{
			tokens.POST("/generate", r.generateToken)
			tokens.GET("/list", r.listTokens)
			tokens.DELETE("/:id", r.revokeToken)
		}
	}

	// Legacy API routes (for backward compatibility)
	api := router.Group("/api")
	{
		// Agent management routes
		api.POST("/agents/register", r.registerAgent)
		api.GET("/agents", r.listAgents)
		api.GET("/agents/:id", r.getAgent)
		api.PUT("/agents/:id/status", r.updateAgentStatus)
		api.DELETE("/agents/:id", r.deleteAgent)
		api.POST("/agents/:id/heartbeat", r.agentHeartbeat)
		api.POST("/agents/heartbeat", r.agentHeartbeat) // Token-based heartbeat (no ID required)
		
		// Task routes
		api.POST("/tasks", r.createTask)
		api.GET("/tasks", r.listTasks)
		api.GET("/tasks/:id", r.getTask)
		
		// System routes
		api.GET("/health", r.getHealth)
		api.GET("/install", r.installScript)
		api.GET("/download", r.downloadAgent)
	}
}

// Agent handlers
func (r *APIRouter) listAgents(c *gin.Context) {
	if r.registry == nil {
		c.JSON(http.StatusOK, gin.H{
			"agents": []gin.H{},
			"total":  0,
		})
		return
	}
	
	// Get agents from registry
	agentInfos := r.registry.List()
	agents := make([]gin.H, 0, len(agentInfos))
	
	for _, agent := range agentInfos {
		agents = append(agents, gin.H{
			"id":            agent.ID,
			"hostname":      agent.Hostname,
			"status":        agent.Status,
			"cpu_type":      agent.CPUType,
			"cpu_logic":     agent.CPULogic,
			"memory":        agent.Memory,
			"os":            agent.OS,
			"manageip":      agent.ManageIP,
			"gpu_num":       agent.GPUNum,
			"gpu_type":      agent.GPUType,
			"last_seen":     agent.LastSeen,
			"registered_at": agent.RegisteredAt,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"total":  len(agents),
	})
}

func (r *APIRouter) getAgent(c *gin.Context) {
	agentID := c.Param("id")
	
	if r.registry == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	
	agent := r.registry.Get(agentID)
	if agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"agent": gin.H{
			"id":            agent.ID,
			"hostname":      agent.Hostname,
			"status":        agent.Status,
			"cpu_type":      agent.CPUType,
			"cpu_logic":     agent.CPULogic,
			"memory":        agent.Memory,
			"os":            agent.OS,
			"sn":            agent.SN,
			"product":       agent.Product,
			"brand":         agent.Brand,
			"netcard":       agent.Netcard,
			"basearch":      agent.Basearch,
			"gpu_num":       agent.GPUNum,
			"gpu_type":      agent.GPUType,
			"last_seen":     agent.LastSeen,
			"registered_at": agent.RegisteredAt,
		},
	})
}

func (r *APIRouter) restartAgent(c *gin.Context) {
	agentID := c.Param("id")
	// TODO: Implement agent restart
	c.JSON(http.StatusOK, gin.H{
		"message": "Restart command sent",
		"agent_id": agentID,
	})
}

func (r *APIRouter) getAgentTasks(c *gin.Context) {
	agentID := c.Param("id")
	// TODO: Implement agent task retrieval
	c.JSON(http.StatusOK, gin.H{
		"tasks": []gin.H{},
		"agent_id": agentID,
	})
}

// Task handlers
func (r *APIRouter) listTasks(c *gin.Context) {
	// TODO: Implement task listing
	c.JSON(http.StatusOK, gin.H{
		"tasks": []gin.H{},
		"total": 0,
	})
}

func (r *APIRouter) createTask(c *gin.Context) {
	var taskRequest struct {
		Type         string   `json:"type"`
		TargetAgents []string `json:"target_agents"`
		Content      string   `json:"content"`
		Timeout      int      `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement task creation
	c.JSON(http.StatusOK, gin.H{
		"message": "Task created successfully",
		"task":    taskRequest,
	})
}

func (r *APIRouter) getTask(c *gin.Context) {
	taskID := c.Param("id")
	// TODO: Implement task retrieval
	c.JSON(http.StatusOK, gin.H{
		"task": gin.H{
			"id": taskID,
		},
	})
}

func (r *APIRouter) cancelTask(c *gin.Context) {
	taskID := c.Param("id")
	// TODO: Implement task cancellation
	c.JSON(http.StatusOK, gin.H{
		"message": "Task cancelled",
		"task_id": taskID,
	})
}

// Cluster handlers
func (r *APIRouter) listClusters(c *gin.Context) {
	clusters := r.clusterMgr.ListClusters()
	c.JSON(http.StatusOK, gin.H{
		"clusters": clusters,
		"total":    len(clusters),
	})
}

func (r *APIRouter) createCluster(c *gin.Context) {
	var cluster cluster.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.clusterMgr.AddCluster(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cluster created successfully",
		"cluster": cluster,
	})
}

func (r *APIRouter) getCluster(c *gin.Context) {
	clusterID := c.Param("id")
	cluster, err := r.clusterMgr.GetCluster(clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cluster": cluster,
	})
}

func (r *APIRouter) updateCluster(c *gin.Context) {
	clusterID := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.clusterMgr.UpdateCluster(clusterID, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cluster updated successfully",
	})
}

func (r *APIRouter) deleteCluster(c *gin.Context) {
	clusterID := c.Param("id")
	if err := r.clusterMgr.DeleteCluster(clusterID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cluster deleted successfully",
	})
}

func (r *APIRouter) getClusterStats(c *gin.Context) {
	clusterID := c.Param("id")
	stats, err := r.clusterMgr.GetClusterStats(clusterID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

func (r *APIRouter) addAgentToCluster(c *gin.Context) {
	clusterID := c.Param("id")
	agentID := c.Param("agent_id")
	
	if err := r.clusterMgr.AddAgentToCluster(clusterID, agentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent added to cluster successfully",
	})
}

func (r *APIRouter) removeAgentFromCluster(c *gin.Context) {
	clusterID := c.Param("id")
	agentID := c.Param("agent_id")
	
	if err := r.clusterMgr.RemoveAgentFromCluster(clusterID, agentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent removed from cluster successfully",
	})
}

// Alert handlers
func (r *APIRouter) listAlerts(c *gin.Context) {
	alerts := r.alertMgr.ListAlerts()
	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

func (r *APIRouter) createAlertRule(c *gin.Context) {
	var rule alert.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.alertMgr.AddAlertRule(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert rule created successfully",
		"rule":    rule,
	})
}

func (r *APIRouter) listAlertRules(c *gin.Context) {
	rules := r.alertMgr.ListAlertRules()
	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"total": len(rules),
	})
}

func (r *APIRouter) updateAlertRule(c *gin.Context) {
	ruleID := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.alertMgr.UpdateAlertRule(ruleID, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert rule updated successfully",
	})
}

func (r *APIRouter) deleteAlertRule(c *gin.Context) {
	ruleID := c.Param("id")
	if err := r.alertMgr.DeleteAlertRule(ruleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert rule deleted successfully",
	})
}

func (r *APIRouter) resolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	if err := r.alertMgr.ResolveAlert(alertID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert resolved successfully",
	})
}

// Plugin handlers
func (r *APIRouter) listPlugins(c *gin.Context) {
	// TODO: Implement plugin listing
	c.JSON(http.StatusOK, gin.H{
		"plugins": []gin.H{},
		"total":   0,
	})
}

func (r *APIRouter) uploadPlugin(c *gin.Context) {
	// TODO: Implement plugin upload
	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin upload not yet implemented",
	})
}

func (r *APIRouter) deletePlugin(c *gin.Context) {
	pluginName := c.Param("name")
	// TODO: Implement plugin deletion
	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin deleted",
		"name":    pluginName,
	})
}

// System handlers
func (r *APIRouter) getSystemStats(c *gin.Context) {
	// Get real statistics from registry
	totalAgents := 0
	onlineAgents := 0
	offlineAgents := 0
	
	if r.registry != nil {
		agents := r.registry.List()
		totalAgents = len(agents)
		for _, agent := range agents {
			if agent.Status == "online" {
				onlineAgents++
			} else {
				offlineAgents++
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"total_agents":   totalAgents,
			"online_agents":  onlineAgents,
			"offline_agents": offlineAgents,
			"total_clusters": len(r.clusterMgr.ListClusters()),
			"total_alerts":   len(r.alertMgr.ListAlerts()),
			"total_tasks":    0,
			"pending_tasks":  0,
		},
	})
}

func (r *APIRouter) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"timestamp": time.Now().Unix(),
	})
}

// Agent registration handler
func (r *APIRouter) registerAgent(c *gin.Context) {
	var agentInfo struct {
		Hostname     string                 `json:"hostname" binding:"required"`
		CPUType      string                 `json:"cpu_type"`
		CPULogic     int                    `json:"cpu_logic"`
		Memsum       int64                  `json:"memsum"`
		Memory       string                 `json:"memory"`
		SN           string                 `json:"sn"`
		Product      string                 `json:"product"`
		Brand        string                 `json:"brand"`
		Netcard      []string               `json:"netcard"`
		Basearch     string                 `json:"basearch"`
		Disk         map[string]interface{} `json:"disk"`
		Raid         string                 `json:"raid"`
		IPMIIP       string                 `json:"ipmi_ip"`
		ManageIP     string                 `json:"manageip"`
		StorageIP    string                 `json:"storageip"`
		ParamIP      string                 `json:"paramip"`
		OS           string                 `json:"os"`
		GPUNum       int                    `json:"gpu_num"`
		GPUType      string                 `json:"gpu_type"`
		GPUVendors   []string               `json:"gpu_vendors"`
		DiskInfo     []map[string]interface{} `json:"disk_info"`
		MemoryInfo   []map[string]interface{} `json:"memory_info"`
		CPUInfo      map[string]interface{} `json:"cpu_info"`
		GPUInfo      []map[string]interface{} `json:"gpu_info"`
		NetworkInfo  []map[string]interface{} `json:"network_info"`
		AgentVersion string                 `json:"agent_version"`
	}

	if err := c.ShouldBindJSON(&agentInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Validate token from Authorization header
	token := c.GetHeader("Authorization")
	if token == "" {
		// Try getting token from query parameter as fallback
		token = c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			return
		}
	}

	// Register agent with registry
	if r.registry != nil {
		agentID := agentInfo.Hostname + "-" + generateRandomID(8)
		
		// Create AgentInfo from request
		info := &core.AgentInfo{
			ID:           agentID,
			Hostname:     agentInfo.Hostname,
			CPUType:      agentInfo.CPUType,
			CPULogic:     agentInfo.CPULogic,
			Memsum:       agentInfo.Memsum,
			Memory:       agentInfo.Memory,
			SN:           agentInfo.SN,
			Product:      agentInfo.Product,
			Brand:        agentInfo.Brand,
			Netcard:      agentInfo.Netcard,
			Basearch:     agentInfo.Basearch,
			Disk:         agentInfo.Disk,
			Raid:         agentInfo.Raid,
			IPMIIP:       agentInfo.IPMIIP,
			ManageIP:     agentInfo.ManageIP,
			StorageIP:    agentInfo.StorageIP,
			ParamIP:      agentInfo.ParamIP,
			OS:           agentInfo.OS,
			Status:       "online",
			GPUNum:       agentInfo.GPUNum,
			GPUType:      agentInfo.GPUType,
			GPUVendors:   agentInfo.GPUVendors,
			DiskInfo:     agentInfo.DiskInfo,
			MemoryInfo:   agentInfo.MemoryInfo,
			CPUInfo:      agentInfo.CPUInfo,
			GPUInfo:      agentInfo.GPUInfo,
			NetworkInfo:  agentInfo.NetworkInfo,
			UpdateTime:   time.Now().Format("2006-01-02 15:04:05"),
			AgentVersion: agentInfo.AgentVersion,
			RegisteredAt: time.Now(),
			LastSeen:     time.Now(),
		}
		
		// Register the agent
		id := r.registry.Register(info)
		
		c.JSON(http.StatusOK, gin.H{
			"id":      id,
			"status":  "registered",
			"message": "Agent registered successfully",
		})
		return
	}
	
	// Fallback if registry is not available
	agentID := agentInfo.Hostname + "-" + generateRandomID(8)
	c.JSON(http.StatusOK, gin.H{
		"id":      agentID,
		"status":  "registered",
		"message": "Agent registered successfully (registry not available)",
	})
}

// Agent heartbeat handler
func (r *APIRouter) agentHeartbeat(c *gin.Context) {
	agentID := c.Param("id")
	
	var heartbeatData struct {
		Status      string                 `json:"status"`
		SystemInfo  map[string]interface{} `json:"system_info,omitempty"`
		Tasks       []string               `json:"tasks,omitempty"`
	}

	if err := c.ShouldBindJSON(&heartbeatData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update agent heartbeat in registry
	if r.registry != nil {
		var agent *core.AgentInfo
		
		if agentID != "" {
			// Use provided agent ID
			agent = r.registry.Get(agentID)
		} else {
			// Token-based heartbeat: try to find agent by hostname from system_info
			if heartbeatData.SystemInfo != nil {
				if hostname, ok := heartbeatData.SystemInfo["hostname"].(string); ok && hostname != "" {
					// Try to find agent by hostname (registry uses hostname as ID base)
					agents := r.registry.List()
					for _, a := range agents {
						if a.Hostname == hostname {
							agent = a
							agentID = a.ID
							break
						}
					}
				}
			}
		}
		
		if agent != nil {
			agent.LastSeen = time.Now()
			if heartbeatData.Status != "" {
				agent.Status = heartbeatData.Status
			} else {
				agent.Status = "online"
			}
			// Update system info if provided
			if heartbeatData.SystemInfo != nil {
				// Update relevant fields from system_info
				if hostname, ok := heartbeatData.SystemInfo["hostname"].(string); ok {
					agent.Hostname = hostname
				}
				if cpuType, ok := heartbeatData.SystemInfo["cpu_type"].(string); ok {
					agent.CPUType = cpuType
				}
				if cpuLogic, ok := heartbeatData.SystemInfo["cpu_logic"].(float64); ok {
					agent.CPULogic = int(cpuLogic)
				}
				if memory, ok := heartbeatData.SystemInfo["memory"].(string); ok {
					agent.Memory = memory
				}
			}
			r.registry.Update(agentID, agent)
		}
		// If agent not found, still return success (may not be registered yet)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Heartbeat received",
		"agent_id": agentID,
	})
}

// Update agent status handler
func (r *APIRouter) updateAgentStatus(c *gin.Context) {
	agentID := c.Param("id")
	
	var statusUpdate struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status values
	validStatuses := []string{"online", "offline", "maintenance", "error"}
	if !contains(validStatuses, statusUpdate.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid status. Must be one of: online, offline, maintenance, error",
		})
		return
	}

	// TODO: Update agent status in registry
	// For now, return success
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "updated",
		"message": "Agent status updated successfully",
		"agent_id": agentID,
		"new_status": statusUpdate.Status,
	})
}

// Delete agent handler
func (r *APIRouter) deleteAgent(c *gin.Context) {
	agentID := c.Param("id")
	
	// TODO: Remove agent from registry
	// For now, return success
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "deleted",
		"message": "Agent deleted successfully",
		"agent_id": agentID,
	})
}

// Install script handler
func (r *APIRouter) installScript(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	script := generateInstallScript(token)
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, script)
}

// Download agent binary handler
func (r *APIRouter) downloadAgent(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	// TODO: Validate token
	// For now, check if token is not empty

	// Get current working directory for better path resolution
	wd, err := os.Getwd()
	if err != nil {
		// Fallback paths
		wd = ""
	}

	// Try multiple possible paths for the binary
	possiblePaths := []string{
		filepath.Join(wd, "../agent/nerve-agent"),      // Relative from server directory
		filepath.Join(wd, "./agent/nerve-agent"),       // Relative from project root
		filepath.Join(wd, "agent/nerve-agent"),          // Alternative relative path
		"../agent/nerve-agent",                          // Relative from server directory (fallback)
		"./agent/nerve-agent",                            // Relative from project root (fallback)
		"agent/nerve-agent",                              // Alternative (fallback)
		"/usr/local/bin/nerve-agent",                     // System path
	}

	var binaryPath string
	var found bool

	for _, path := range possiblePaths {
		// Try to get absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		
		if fileInfo, err := os.Stat(absPath); err == nil {
			// Check if it's a regular file and not a directory
			if !fileInfo.Mode().IsRegular() {
				continue
			}
			binaryPath = absPath
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent binary not found. Please build it first: cd agent && go build -o nerve-agent",
			"hint":  "Checked paths: " + strings.Join(possiblePaths, ", "),
			"cwd":   wd,
		})
		return
	}

	// Set headers for file download
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=nerve-agent")
	
	// Send file
	c.File(binaryPath)
}

// Helper functions
func generateRandomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateInstallScript(token string) string {
	return `#!/bin/bash
set -e

# Parse arguments
SERVER_URL="${SERVER_URL:-http://localhost:8090}"
TOKEN=""
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --token) TOKEN="$2"; shift ;;
    --server) SERVER_URL="$2"; shift ;;
  esac
  shift
done

if [ -z "$TOKEN" ]; then
  echo "Error: --token is required"
  exit 1
fi

echo "Installing Nerve Agent..."

# Download agent binary
curl -fSL "$SERVER_URL/api/download?token=$TOKEN" -o /usr/local/bin/nerve-agent
chmod +x /usr/local/bin/nerve-agent

# Create systemd service
cat > /etc/systemd/system/nerve-agent.service << EOF
[Unit]
Description=Nerve Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nerve-agent --server=$SERVER_URL --token=$TOKEN
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable nerve-agent
systemctl start nerve-agent

echo "Nerve Agent installed successfully!"
`
}

// NewMetricsHandler creates a metrics handler for Prometheus
func NewMetricsHandler(collector *metrics.MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement Prometheus metrics endpoint
		c.JSON(http.StatusOK, gin.H{
			"message": "Metrics endpoint not yet implemented",
		})
	}
}

// Token management handlers
func (r *APIRouter) generateToken(c *gin.Context) {
	var tokenRequest struct {
		Name      string `json:"name" binding:"required"`
		ExpiresIn int    `json:"expires_in"` // seconds
	}

	if err := c.ShouldBindJSON(&tokenRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a random token
	token := generateRandomToken(32)
	
	// TODO: Store token in database with expiration
	// For now, return the token directly
	
	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"name":       tokenRequest.Name,
		"expires_at": time.Now().Add(time.Duration(tokenRequest.ExpiresIn) * time.Second),
		"created_at": time.Now(),
	})
}

func (r *APIRouter) listTokens(c *gin.Context) {
	// TODO: Get tokens from database
	// For now, return mock data
	tokens := []gin.H{
		{
			"id":         "token-001",
			"name":       "Agent安装Token_2025-01-28T10:30:00",
			"token":      "nerve_abc123...",
			"created_at": time.Now().Add(-2 * time.Hour),
			"expires_at": time.Now().Add(22 * time.Hour),
			"status":     "active",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
		"total":  len(tokens),
	})
}

func (r *APIRouter) revokeToken(c *gin.Context) {
	tokenID := c.Param("id")
	
	// TODO: Revoke token in database
	// For now, return success
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Token revoked successfully",
		"token_id": tokenID,
	})
}

// Helper function to generate random token
func generateRandomToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return "nerve_" + string(b)
}

