// Package security provides audit logging functionality for operation tracking.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package security

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogger manages audit logging
type AuditLogger struct {
	logFile string
	mutex   sync.Mutex
}

// AuditEvent represents an audit event
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id,omitempty"`
	AgentID     string                 `json:"agent_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Result      string                 `json:"result"`
	Details     map[string]interface{} `json:"details"`
	RequestID   string                 `json:"request_id,omitempty"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logFile string) *AuditLogger {
	return &AuditLogger{
		logFile: logFile,
	}
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(event *AuditEvent) error {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Convert to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %v", err)
	}

	// Append to log file
	file, err := os.OpenFile(al.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(append(eventJSON, '\n')); err != nil {
		return fmt.Errorf("failed to write audit event: %v", err)
	}

	return nil
}

// LogAuthentication logs authentication events
func (al *AuditLogger) LogAuthentication(userID, agentID, ipAddress, userAgent, result string) error {
	event := &AuditEvent{
		EventType: "authentication",
		UserID:    userID,
		AgentID:   agentID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Action:    "login",
		Resource:  "system",
		Result:    result,
		Details: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}

	return al.LogEvent(event)
}

// LogAuthorization logs authorization events
func (al *AuditLogger) LogAuthorization(userID, agentID, action, resource, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: "authorization",
		UserID:    userID,
		AgentID:   agentID,
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
	}

	return al.LogEvent(event)
}

// LogDataAccess logs data access events
func (al *AuditLogger) LogDataAccess(userID, agentID, action, resource, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: "data_access",
		UserID:    userID,
		AgentID:   agentID,
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
	}

	return al.LogEvent(event)
}

// LogSystemEvent logs system events
func (al *AuditLogger) LogSystemEvent(eventType, action, resource, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: eventType,
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
	}

	return al.LogEvent(event)
}

// LogTaskExecution logs task execution events
func (al *AuditLogger) LogTaskExecution(userID, agentID, taskID, taskType, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: "task_execution",
		UserID:    userID,
		AgentID:   agentID,
		Action:    "execute_task",
		Resource:  fmt.Sprintf("task/%s", taskID),
		Result:    result,
		Details:   details,
	}

	return al.LogEvent(event)
}

// LogConfigurationChange logs configuration change events
func (al *AuditLogger) LogConfigurationChange(userID, action, resource, result string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: "configuration_change",
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
	}

	return al.LogEvent(event)
}

// AuditMiddleware creates a middleware for audit logging
func AuditMiddleware(auditLogger *AuditLogger) func(c *gin.Context) {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log the request
		duration := time.Since(start)
		status := c.Writer.Status()

		event := &AuditEvent{
			EventType: "api_request",
			IPAddress: c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Action:    c.Request.Method,
			Resource:  c.Request.URL.Path,
			Result:    fmt.Sprintf("%d", status),
			Details: map[string]interface{}{
				"duration_ms": duration.Milliseconds(),
				"request_size": c.Request.ContentLength,
				"response_size": c.Writer.Size(),
			},
		}

		// Add user info if available
		if userID, exists := c.Get("user_id"); exists {
			event.UserID = userID.(string)
		}

		// Add agent info if available
		if agentID, exists := c.Get("agent_id"); exists {
			event.AgentID = agentID.(string)
		}

		// Log the event
		auditLogger.LogEvent(event)
	}
}

// GetAuditLogs reads audit logs (for admin purposes)
func (al *AuditLogger) GetAuditLogs(limit int) ([]*AuditEvent, error) {
	al.mutex.Lock()
	defer al.mutex.Unlock()

	file, err := os.Open(al.logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %v", err)
	}
	defer file.Close()

	var events []*AuditEvent
	decoder := json.NewDecoder(file)

	for decoder.More() && len(events) < limit {
		var event AuditEvent
		if err := decoder.Decode(&event); err != nil {
			continue // Skip malformed entries
		}
		events = append(events, &event)
	}

	return events, nil
}

