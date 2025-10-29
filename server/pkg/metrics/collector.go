// Package metrics provides Prometheus metrics collection and exposure functionality.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector collects and exposes metrics
type MetricsCollector struct {
	// Agent metrics
	agentTotal           prometheus.Gauge
	agentOnline          prometheus.Gauge
	agentOffline         prometheus.Gauge
	agentHeartbeatTotal  prometheus.Counter
	agentHeartbeatErrors prometheus.Counter

	// Task metrics
	taskTotal          prometheus.Counter
	taskSuccess        prometheus.Counter
	taskFailed         prometheus.Counter
	taskDuration       prometheus.Histogram

	// System metrics
	systemInfoUpdateTotal prometheus.Counter
	systemInfoUpdateErrors prometheus.Counter

	// Performance metrics
	apiRequestTotal     *prometheus.CounterVec
	apiRequestDuration  *prometheus.HistogramVec
	apiRequestErrors    *prometheus.CounterVec

	// Data metrics
	dataWriteTotal  prometheus.Counter
	dataWriteErrors prometheus.Counter
	dataReadTotal   prometheus.Counter

	mu sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		agentTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "nerve_agent_total",
			Help: "Total number of registered agents",
		}),
		agentOnline: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "nerve_agent_online",
			Help: "Number of online agents",
		}),
		agentOffline: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "nerve_agent_offline",
			Help: "Number of offline agents",
		}),
		agentHeartbeatTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_agent_heartbeat_total",
			Help: "Total number of agent heartbeats",
		}),
		agentHeartbeatErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_agent_heartbeat_errors_total",
			Help: "Total number of agent heartbeat errors",
		}),
		taskTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_task_total",
			Help: "Total number of tasks executed",
		}),
		taskSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_task_success_total",
			Help: "Total number of successful tasks",
		}),
		taskFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_task_failed_total",
			Help: "Total number of failed tasks",
		}),
		taskDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "nerve_task_duration_seconds",
			Help:    "Task execution duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		}),
		systemInfoUpdateTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_system_info_update_total",
			Help: "Total number of system info updates",
		}),
		systemInfoUpdateErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_system_info_update_errors_total",
			Help: "Total number of system info update errors",
		}),
		apiRequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nerve_api_requests_total",
				Help: "Total number of API requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		apiRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "nerve_api_request_duration_seconds",
				Help:    "API request duration in seconds",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
			},
			[]string{"method", "endpoint"},
		),
		apiRequestErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nerve_api_request_errors_total",
				Help: "Total number of API request errors",
			},
			[]string{"method", "endpoint"},
		),
		dataWriteTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_data_write_total",
			Help: "Total number of data write operations",
		}),
		dataWriteErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_data_write_errors_total",
			Help: "Total number of data write errors",
		}),
		dataReadTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "nerve_data_read_total",
			Help: "Total number of data read operations",
		}),
	}
}

// UpdateAgentMetrics updates agent metrics
func (mc *MetricsCollector) UpdateAgentMetrics(total, online, offline int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.agentTotal.Set(float64(total))
	mc.agentOnline.Set(float64(online))
	mc.agentOffline.Set(float64(offline))
}

// RecordHeartbeat records a heartbeat event
func (mc *MetricsCollector) RecordHeartbeat(success bool) {
	mc.agentHeartbeatTotal.Inc()
	if !success {
		mc.agentHeartbeatErrors.Inc()
	}
}

// RecordTask records a task execution
func (mc *MetricsCollector) RecordTask(success bool, duration time.Duration) {
	mc.taskTotal.Inc()
	if success {
		mc.taskSuccess.Inc()
	} else {
		mc.taskFailed.Inc()
	}
	mc.taskDuration.Observe(duration.Seconds())
}

// RecordSystemInfoUpdate records a system info update
func (mc *MetricsCollector) RecordSystemInfoUpdate(success bool) {
	mc.systemInfoUpdateTotal.Inc()
	if !success {
		mc.systemInfoUpdateErrors.Inc()
	}
}

// RecordAPIRequest records an API request
func (mc *MetricsCollector) RecordAPIRequest(method, endpoint string, status int, duration time.Duration) {
	mc.apiRequestTotal.WithLabelValues(method, endpoint, string(rune(status))).Inc()
	mc.apiRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

	if status >= 400 {
		mc.apiRequestErrors.WithLabelValues(method, endpoint).Inc()
	}
}

// RecordDataWrite records a data write operation
func (mc *MetricsCollector) RecordDataWrite(success bool) {
	mc.dataWriteTotal.Inc()
	if !success {
		mc.dataWriteErrors.Inc()
	}
}

// RecordDataRead records a data read operation
func (mc *MetricsCollector) RecordDataRead() {
	mc.dataReadTotal.Inc()
}

// AgentMetrics represents agent-specific metrics
type AgentMetrics struct {
	CPUUsage        float64
	MemoryUsage     float64
	DiskUsage       float64
	NetworkRxBytes  int64
	NetworkTxBytes  int64
	Uptime          time.Duration
	LastHeartbeat   time.Time
}

// CollectAgentMetrics collects metrics from an agent
func (mc *MetricsCollector) CollectAgentMetrics(agentID string, metrics AgentMetrics) {
	// TODO: Store agent-specific metrics in a time series database
	// For now, we'll use Prometheus Gauge vectors
}

// GetMetricsSnapshot returns a snapshot of current metrics
func (mc *MetricsCollector) GetMetricsSnapshot() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return map[string]interface{}{
		"agent_total":     getGaugeValue(mc.agentTotal),
		"agent_online":   getGaugeValue(mc.agentOnline),
		"agent_offline":   getGaugeValue(mc.agentOffline),
		"heartbeat_total": getCounterValue(mc.agentHeartbeatTotal),
		"task_total":      getCounterValue(mc.taskTotal),
		"task_success":    getCounterValue(mc.taskSuccess),
		"task_failed":     getCounterValue(mc.taskFailed),
	}
}

// Helper functions
func getGaugeValue(gauge prometheus.Gauge) float64 {
	// TODO: Implement actual gauge value reading
	return 0
}

func getCounterValue(counter prometheus.Counter) float64 {
	// TODO: Implement actual counter value reading
	return 0
}

