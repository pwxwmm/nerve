package core

import (
	"sync"
	"time"

	"github.com/nerve/server/pkg/log"
	"github.com/nerve/server/pkg/storage"
)

// AgentInfo represents agent information
type AgentInfo struct {
	ID           string                 `json:"id"`
	Hostname     string                 `json:"hostname"`
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
	Status       string                 `json:"status"`
	GPUNum       int                    `json:"gpu_num"`
	GPUType      string                 `json:"gpu_type"`
	GPUVendors   []string               `json:"gpu_vendors"`
	DiskInfo     []map[string]interface{} `json:"disk_info"`
	MemoryInfo   []map[string]interface{} `json:"memory_info"`
	CPUInfo      map[string]interface{} `json:"cpu_info"`
	GPUInfo      []map[string]interface{} `json:"gpu_info"`
	NetworkInfo  []map[string]interface{} `json:"network_info"`
	UpdateTime   string                 `json:"update_time"`
	AgentVersion string                 `json:"agent_version"`
	RegisteredAt time.Time              `json:"registered_at"`
	LastSeen     time.Time              `json:"last_seen"`
}

// Task represents a task
type Task struct {
	ID      string                 `json:"id"`
	AgentID string                 `json:"agent_id"`
	Type    string                 `json:"type"`
	Command string                 `json:"command,omitempty"`
	Script  string                 `json:"script,omitempty"`
	Plugin  string                 `json:"plugin,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Timeout int                    `json:"timeout,omitempty"`
	Status  string                 `json:"status"`
}

// TaskResult represents task execution result
type TaskResult struct {
	TaskID  string `json:"task_id"`
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Registry manages agent registry
type Registry struct {
	mu    sync.RWMutex
	store storage.Storage
	agents map[string]*AgentInfo
	logger log.Logger
}

// NewRegistry creates a new registry
func NewRegistry(store storage.Storage, logger log.Logger) *Registry {
	registry := &Registry{
		store:  store,
		agents: make(map[string]*AgentInfo),
		logger: logger,
	}

	// Start cleanup goroutine
	go registry.cleanupStaleAgents()

	return registry
}

// Register registers an agent
func (r *Registry) Register(agent *AgentInfo) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := agent.Hostname // Use hostname as ID for now
	agent.ID = id

	r.agents[id] = agent
	r.logger.Infof("Registered agent: %s", id)

	return id
}

// Update updates agent information
func (r *Registry) Update(id string, agent *AgentInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.agents[id]; ok {
		*existing = *agent
		existing.ID = id
	}
}

// Get retrieves an agent by ID
func (r *Registry) Get(id string) *AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.agents[id]
}

// List returns all agents
func (r *Registry) List() []*AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentInfo, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents
}

// cleanupStaleAgents removes agents that haven't been seen for 5 minutes
func (r *Registry) cleanupStaleAgents() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		now := time.Now()
		for id, agent := range r.agents {
			if now.Sub(agent.LastSeen) > 5*time.Minute {
				agent.Status = "offline"
				r.logger.Infof("Agent marked as offline: %s", id)
			}
		}
		r.mu.Unlock()
	}
}

