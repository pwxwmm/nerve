// Package alert provides alert management functionality with rule engine and notifications.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package alert

import (
	"fmt"
	"sync"
	"time"
)

// AlertManager manages alerts and notifications
type AlertManager struct {
	alerts    map[string]*Alert
	rules     map[string]*AlertRule
	mutex     sync.RWMutex
	notifiers map[string]Notifier
}

// Alert represents an alert instance
type Alert struct {
	ID          string                 `json:"id"`
	RuleID      string                 `json:"rule_id"`
	AgentID     string                 `json:"agent_id"`
	ClusterID   string                 `json:"cluster_id,omitempty"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// AlertRule defines alert conditions
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Severity    string                 `json:"severity"`
	Conditions  []AlertCondition       `json:"conditions"`
	Actions     []AlertAction          `json:"actions"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AlertCondition defines a single condition
type AlertCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// AlertAction defines an action to take when alert fires
type AlertAction struct {
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
	Enabled bool                   `json:"enabled"`
}

// Notifier interface for alert notifications
type Notifier interface {
	Send(alert *Alert) error
	Name() string
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts:    make(map[string]*Alert),
		rules:     make(map[string]*AlertRule),
		notifiers: make(map[string]Notifier),
	}
}

// AddAlertRule adds a new alert rule
func (am *AlertManager) AddAlertRule(rule *AlertRule) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if _, exists := am.rules[rule.ID]; exists {
		return fmt.Errorf("alert rule %s already exists", rule.ID)
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	am.rules[rule.ID] = rule

	return nil
}

// GetAlertRule retrieves an alert rule by ID
func (am *AlertManager) GetAlertRule(id string) (*AlertRule, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	rule, exists := am.rules[id]
	if !exists {
		return nil, fmt.Errorf("alert rule %s not found", id)
	}

	return rule, nil
}

// ListAlertRules returns all alert rules
func (am *AlertManager) ListAlertRules() []*AlertRule {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var rules []*AlertRule
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}

	return rules
}

// UpdateAlertRule updates an existing alert rule
func (am *AlertManager) UpdateAlertRule(id string, updates map[string]interface{}) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	rule, exists := am.rules[id]
	if !exists {
		return fmt.Errorf("alert rule %s not found", id)
	}

	// Update fields
	if name, ok := updates["name"].(string); ok {
		rule.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		rule.Description = desc
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		rule.Enabled = enabled
	}
	if severity, ok := updates["severity"].(string); ok {
		rule.Severity = severity
	}

	rule.UpdatedAt = time.Now()

	return nil
}

// DeleteAlertRule removes an alert rule
func (am *AlertManager) DeleteAlertRule(id string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if _, exists := am.rules[id]; !exists {
		return fmt.Errorf("alert rule %s not found", id)
	}

	delete(am.rules, id)
	return nil
}

// EvaluateRules evaluates all enabled alert rules against agent data
func (am *AlertManager) EvaluateRules(agentID string, data map[string]interface{}) error {
	am.mutex.RLock()
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	am.mutex.RUnlock()

	for _, rule := range rules {
		if am.evaluateRule(rule, agentID, data) {
			alert := &Alert{
				ID:        fmt.Sprintf("%s-%d", rule.ID, time.Now().Unix()),
				RuleID:    rule.ID,
				AgentID:   agentID,
				Severity:  rule.Severity,
				Status:    "active",
				Message:   rule.Description,
				Data:      data,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := am.createAlert(alert); err != nil {
				fmt.Printf("Failed to create alert: %v\n", err)
			}

			// Execute actions
			am.executeActions(rule.Actions, alert)
		}
	}

	return nil
}

// evaluateRule checks if a rule condition is met
func (am *AlertManager) evaluateRule(rule *AlertRule, agentID string, data map[string]interface{}) bool {
	for _, condition := range rule.Conditions {
		if !am.evaluateCondition(condition, data) {
			return false
		}
	}
	return true
}

// evaluateCondition checks a single condition
func (am *AlertManager) evaluateCondition(condition AlertCondition, data map[string]interface{}) bool {
	value, exists := data[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "eq":
		return value == condition.Value
	case "ne":
		return value != condition.Value
	case "gt":
		return compareNumbers(value, condition.Value) > 0
	case "gte":
		return compareNumbers(value, condition.Value) >= 0
	case "lt":
		return compareNumbers(value, condition.Value) < 0
	case "lte":
		return compareNumbers(value, condition.Value) <= 0
	case "contains":
		if str, ok := value.(string); ok {
			if target, ok := condition.Value.(string); ok {
				return contains(str, target)
			}
		}
		return false
	default:
		return false
	}
}

// createAlert creates a new alert
func (am *AlertManager) createAlert(alert *Alert) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.alerts[alert.ID] = alert
	return nil
}

// executeActions executes alert actions
func (am *AlertManager) executeActions(actions []AlertAction, alert *Alert) {
	for _, action := range actions {
		if !action.Enabled {
			continue
		}

		switch action.Type {
		case "webhook":
			am.executeWebhookAction(action, alert)
		case "email":
			am.executeEmailAction(action, alert)
		case "slack":
			am.executeSlackAction(action, alert)
		default:
			fmt.Printf("Unknown action type: %s\n", action.Type)
		}
	}
}

// executeWebhookAction executes a webhook action
func (am *AlertManager) executeWebhookAction(action AlertAction, alert *Alert) {
	// TODO: Implement webhook execution
	fmt.Printf("Executing webhook action for alert %s\n", alert.ID)
}

// executeEmailAction executes an email action
func (am *AlertManager) executeEmailAction(action AlertAction, alert *Alert) {
	// TODO: Implement email execution
	fmt.Printf("Executing email action for alert %s\n", alert.ID)
}

// executeSlackAction executes a Slack action
func (am *AlertManager) executeSlackAction(action AlertAction, alert *Alert) {
	// TODO: Implement Slack execution
	fmt.Printf("Executing Slack action for alert %s\n", alert.ID)
}

// ListAlerts returns all alerts
func (am *AlertManager) ListAlerts() []*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var alerts []*Alert
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}

	return alerts
}

// ResolveAlert marks an alert as resolved
func (am *AlertManager) ResolveAlert(alertID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert %s not found", alertID)
	}

	alert.Status = "resolved"
	alert.UpdatedAt = time.Now()
	now := time.Now()
	alert.ResolvedAt = &now

	return nil
}

// RegisterNotifier registers a notification handler
func (am *AlertManager) RegisterNotifier(name string, notifier Notifier) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.notifiers[name] = notifier
}

// Helper functions

func compareNumbers(a, b interface{}) int {
	// Simple numeric comparison
	// TODO: Implement proper numeric comparison
	return 0
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

