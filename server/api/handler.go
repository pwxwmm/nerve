package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nerve/server/core"
	"github.com/nerve/server/pkg/log"
)

// Handler handles HTTP requests
type Handler struct {
	registry  *core.Registry
	scheduler *core.Scheduler
	logger    log.Logger
}

// NewHandler creates a new handler
func NewHandler(registry *core.Registry, scheduler *core.Scheduler, logger log.Logger) *Handler {
	return &Handler{
		registry:  registry,
		scheduler: scheduler,
		logger:    logger,
	}
}

// RegisterAgent handles agent registration
func (h *Handler) RegisterAgent(c *gin.Context) {
	var agent core.AgentInfo
	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set registration time
	agent.RegisteredAt = time.Now()
	agent.LastSeen = time.Now()
	agent.Status = "online"

	// Validate token (simple check for now)
	token := c.GetHeader("Authorization")
	if !strings.HasPrefix(token, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Register agent
	id := h.registry.Register(&agent)
	h.logger.Infof("Agent registered: ID=%s, Hostname=%s", id, agent.Hostname)

	c.JSON(http.StatusOK, gin.H{
		"id":    id,
		"status": "registered",
	})
}

// Heartbeat handles agent heartbeat
func (h *Handler) Heartbeat(c *gin.Context) {
	var agentInfo core.AgentInfo
	if err := c.ShouldBindJSON(&agentInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update agent info
	agentInfo.LastSeen = time.Now()
	agentInfo.Status = "online"

	// Update registry
	h.registry.Update(agentInfo.Hostname, &agentInfo)
	h.logger.Debugf("Heartbeat from: %s", agentInfo.Hostname)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetTasks returns pending tasks for an agent
func (h *Handler) GetTasks(c *gin.Context) {
	// TODO: Implement task retrieval based on agent identity
	tasks := []core.Task{
		// Example: no tasks for now
	}

	c.JSON(http.StatusOK, tasks)
}

// SubmitTaskResult handles task execution results
func (h *Handler) SubmitTaskResult(c *gin.Context) {
	var result core.TaskResult
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("Task result: ID=%s, Success=%v", result.TaskID, result.Success)
	
	// TODO: Update task status in scheduler

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// ListAgents returns all registered agents
func (h *Handler) ListAgents(c *gin.Context) {
	agents := h.registry.List()
	c.JSON(http.StatusOK, agents)
}

// GetAgent returns a specific agent
func (h *Handler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	agent := h.registry.Get(id)
	if agent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// InstallScript returns the installation script
func (h *Handler) InstallScript(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	script := generateInstallScript(token)
	c.String(http.StatusOK, script)
}

// DownloadAgent returns the agent binary
func (h *Handler) DownloadAgent(c *gin.Context) {
	// TODO: Implement actual binary download
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}


