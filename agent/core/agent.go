// Package core provides the core agent functionality for Nerve.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/nerve/agent/pkg/log"
	"github.com/nerve/agent/pkg/sysinfo"
)

// Agent represents the nerve agent
type Agent struct {
	serverURL   string
	token       string
	agentID     string
	interval    time.Duration
	client      *http.Client
	logger      log.Logger
	stopChan    chan struct{}
	wg          sync.WaitGroup
	registered  bool
	mu          sync.RWMutex
}

// SystemInfo represents collected system information
type SystemInfo struct {
	Hostname       string                 `json:"hostname"`
	CPUType        string                 `json:"cpu_type"`
	CPULogic       int                    `json:"cpu_logic"`
	Memsum         int64                  `json:"memsum"`
	Memory         string                 `json:"memory"`
	SN             string                 `json:"sn"`
	Product        string                 `json:"product"`
	Brand          string                 `json:"brand"`
	Netcard        []string               `json:"netcard"`
	Basearch       string                 `json:"basearch"`
	Disk           map[string]interface{} `json:"disk"`
	Raid           string                 `json:"raid"`
	IPMIIP         string                 `json:"ipmi_ip"`
	ManageIP       string                 `json:"manageip"`
	StorageIP      string                 `json:"storageip"`
	ParamIP        string                 `json:"paramip"`
	OS             string                 `json:"os"`
	Status         int                    `json:"status"`
	GPUNum         int                    `json:"gpu_num"`
	GPUType        string                 `json:"gpu_type"`
	GPUVendors     []string               `json:"gpu_vendors"`
	DiskInfo       []map[string]interface{} `json:"disk_info"`
	MemoryInfo     []map[string]interface{} `json:"memory_info"`
	CPUInfo        map[string]interface{} `json:"cpu_info"`
	GPUInfo        []map[string]interface{} `json:"gpu_info"`
	NetworkInfo    []map[string]interface{} `json:"network_info"`
	UpdateTime     string                 `json:"update_time"`
	AgentVersion   string                 `json:"agent_version"`
}

// Task represents a task from the server
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Command     string                 `json:"command,omitempty"`
	Script      string                 `json:"script,omitempty"`
	Plugin      string                 `json:"plugin,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
}

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID  string `json:"task_id"`
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

const (
	DefaultTimeout = 30 * time.Second
	UserAgent      = "Nerve-Agent/1.0"
)

// NewAgent creates a new agent instance (deprecated, use NewAgentWithLogger)
func NewAgent(serverURL, token string, interval time.Duration, logger log.Logger) *Agent {
	return NewAgentWithLogger(serverURL, token, interval, logger)
}

// NewAgentWithLogger creates a new agent instance with a logger
func NewAgentWithLogger(serverURL, token string, interval time.Duration, logger log.Logger) *Agent {
	return &Agent{
		serverURL: serverURL,
		token:     token,
		interval:  interval,
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Register registers the agent with the server
func (a *Agent) Register() error {
	info := a.collectSystemInfo()
	
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal system info: %w", err)
	}

	req, err := http.NewRequest("POST", a.serverURL+"/api/agents/register", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	a.setAuthHeaders(req)
	
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse response to get agent ID
	var registerResp struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err == nil && registerResp.ID != "" {
		a.mu.Lock()
		a.agentID = registerResp.ID
		a.registered = true
		a.mu.Unlock()
		a.logger.Infof("Registered successfully: ID=%s, Hostname=%s", registerResp.ID, info.Hostname)
	} else {
		a.mu.Lock()
		a.registered = true
		a.mu.Unlock()
		a.logger.Infof("Registered successfully: Hostname=%s", info.Hostname)
	}

	return nil
}

// Collect system information
func (a *Agent) collectSystemInfo() SystemInfo {
	// Collect real system information
	hostname := sysinfo.Hostname()
	cpuType, cpuLogic := sysinfo.GetCPUData()
	memsum, memory := sysinfo.GetMemory()
	sn := sysinfo.GetSN()
	product := sysinfo.GetProduct()
	brand := sysinfo.GetBrand()
	netcard := sysinfo.GetNetcard()
	basearch := sysinfo.Basearch()
	disk := sysinfo.Disk()
	raid := sysinfo.Raid()
	ipmiIP := sysinfo.IPMI()
	osInfo := sysinfo.OS()
	gpuInfo := sysinfo.GPUInfo()
	
	// Extract GPU information
	gpuNum := 0
	gpuType := ""
	gpuVendors := []string{}
	if count, ok := gpuInfo["count"].(int); ok {
		gpuNum = count
	}
	if gpuTypeStr, ok := gpuInfo["type"].(string); ok && gpuTypeStr != "" {
		gpuType = gpuTypeStr
	}
	if vendors, ok := gpuInfo["vendors"].([]string); ok {
		gpuVendors = vendors
	}
	
	return SystemInfo{
		Hostname:     hostname,
		CPUType:      cpuType,
		CPULogic:     cpuLogic,
		Memsum:       memsum,
		Memory:       memory,
		SN:           sn,
		Product:      product,
		Brand:        brand,
		Netcard:      netcard,
		Basearch:     basearch,
		Disk:         disk,
		Raid:         raid,
		IPMIIP:       ipmiIP,
		ManageIP:     sysinfo.ManagerIP(),
		StorageIP:    "",
		ParamIP:      sysinfo.ParamIP(),
		OS:           osInfo,
		Status:       0,
		GPUNum:       gpuNum,
		GPUType:      gpuType,
		GPUVendors:   gpuVendors,
		DiskInfo:     sysinfo.GetDiskInfo(),
		MemoryInfo:   sysinfo.GetMemoryInfo(),
		CPUInfo:      sysinfo.GetCPUInfo(),
		GPUInfo:      sysinfo.GetGPUInfos(),
		NetworkInfo:  sysinfo.GetNetworkInfo(),
		UpdateTime:   time.Now().Format("2006-01-02 15:04:05"),
		AgentVersion: "1.0.0",
	}
}

// StartHeartbeat starts the heartbeat goroutine
func (a *Agent) StartHeartbeat() {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		ticker := time.NewTicker(a.interval)
		defer ticker.Stop()

		for {
			select {
			case <-a.stopChan:
				return
			case <-ticker.C:
				if err := a.heartbeat(); err != nil {
					a.logger.Errorf("Heartbeat failed: %v", err)
				}
			}
		}
	}()
}

// heartbeat sends heartbeat to server
func (a *Agent) heartbeat() error {
	a.mu.RLock()
	registered := a.registered
	a.mu.RUnlock()

	if !registered {
		return nil
	}

	info := a.collectSystemInfo()
	
	// Format heartbeat data according to backend expectations
	heartbeatData := map[string]interface{}{
		"status":     "online",
		"system_info": info,
	}
	
	data, err := json.Marshal(heartbeatData)
	if err != nil {
		return err
	}

	// Use agent ID if available, otherwise try without ID (backend may support token-based heartbeat)
	a.mu.RLock()
	agentID := a.agentID
	a.mu.RUnlock()
	
	heartbeatURL := a.serverURL + "/api/agents/heartbeat"
	if agentID != "" {
		heartbeatURL = a.serverURL + "/api/agents/" + agentID + "/heartbeat"
	}
	
	req, err := http.NewRequest("POST", heartbeatURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	a.setAuthHeaders(req)
	
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat returned %d", resp.StatusCode)
	}

	a.logger.Debugf("Heartbeat sent successfully")
	return nil
}

// StartTaskListener starts listening for tasks from server
func (a *Agent) StartTaskListener() {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-a.stopChan:
				return
			case <-ticker.C:
				tasks := a.fetchTasks()
				for _, task := range tasks {
					go a.executeTask(task)
				}
			}
		}
	}()
}

// fetchTasks fetches pending tasks from server
func (a *Agent) fetchTasks() []Task {
	req, err := http.NewRequest("GET", a.serverURL+"/api/tasks", nil)
	if err != nil {
		a.logger.Errorf("Create request: %v", err)
		return nil
	}

	a.setAuthHeaders(req)
	
	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.Errorf("Fetch tasks: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	// Backend returns {tasks: [], total: 0} format
	var response struct {
		Tasks []Task `json:"tasks"`
		Total int    `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		a.logger.Errorf("Decode tasks: %v", err)
		return nil
	}

	return response.Tasks
}

// executeTask executes a task and reports results
func (a *Agent) executeTask(task Task) {
	a.logger.Infof("Executing task: %s (type=%s)", task.ID, task.Type)

	var result TaskResult
	result.TaskID = task.ID
	result.Success = false

	// Execute based on task type
	switch task.Type {
	case "command":
		result = a.executeCommand(task)
	case "script":
		result = a.executeScript(task)
	case "hook":
		result = a.executeHook(task)
	default:
		result.Error = fmt.Sprintf("unknown task type: %s", task.Type)
	}

	// Report result back to server
	a.reportTaskResult(result)
}

// executeCommand executes a shell command
func (a *Agent) executeCommand(task Task) TaskResult {
	// TODO: Implement command execution
	return TaskResult{
		TaskID:  task.ID,
		Success: false,
		Error:   "command execution not implemented",
	}
}

// executeScript executes a script
func (a *Agent) executeScript(task Task) TaskResult {
	// TODO: Implement script execution
	return TaskResult{
		TaskID:  task.ID,
		Success: false,
		Error:   "script execution not implemented",
	}
}

// executeHook executes a hook plugin
func (a *Agent) executeHook(task Task) TaskResult {
	// TODO: Implement hook execution
	return TaskResult{
		TaskID:  task.ID,
		Success: false,
		Error:   "hook execution not implemented",
	}
}

// reportTaskResult reports task execution result to server
func (a *Agent) reportTaskResult(result TaskResult) {
	data, err := json.Marshal(result)
	if err != nil {
		a.logger.Errorf("Marshal result: %v", err)
		return
	}

	req, err := http.NewRequest("POST", a.serverURL+"/api/tasks/"+result.TaskID+"/result", bytes.NewReader(data))
	if err != nil {
		a.logger.Errorf("Create request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	a.setAuthHeaders(req)
	
	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.Errorf("Report result: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		a.logger.Infof("Task result reported: %s", result.TaskID)
	}
}

// setAuthHeaders sets authentication headers
func (a *Agent) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json")
}

// Stop stops the agent
func (a *Agent) Stop() {
	close(a.stopChan)
	a.wg.Wait()
	a.logger.Info("Agent stopped")
}

