# Prometheus Integration for Nerve

This document describes how to integrate Nerve with Prometheus for monitoring and alerting.

## Quick Start

### 1. Start Nerve Server with Metrics

```bash
./server/nerve-center --addr :8080 --metrics-addr :9090
```

Metrics will be available at: http://localhost:9090/metrics

### 2. Configure Prometheus

Add to your Prometheus configuration (`prometheus.yml`):

```yaml
scrape_configs:
  - job_name: 'nerve-server'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:9090']
```

### 3. Start Prometheus

```bash
# Using Docker
docker run -d \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Or use Docker Compose
cd deploy/prometheus
docker-compose up -d
```

### 4. Access Prometheus

- **Prometheus UI**: http://localhost:9090
- **Nerve Metrics**: http://localhost:9090/metrics

## Available Metrics

### Agent Metrics

```promql
# Total number of registered agents
nerve_agent_total

# Number of online agents
nerve_agent_online

# Number of offline agents
nerve_agent_offline

# Total heartbeat count
nerve_agent_heartbeat_total

# Heartbeat errors
nerve_agent_heartbeat_errors_total
```

### Task Metrics

```promql
# Total tasks executed
nerve_task_total

# Successful tasks
nerve_task_success_total

# Failed tasks
nerve_task_failed_total

# Task execution duration histogram
nerve_task_duration_seconds
```

### System Metrics

```promql
# System info updates
nerve_system_info_update_total

# System info update errors
nerve_system_info_update_errors_total
```

### API Metrics

```promql
# API requests by method, endpoint, status
nerve_api_requests_total

# API request duration by method, endpoint
nerve_api_request_duration_seconds

# API request errors
nerve_api_request_errors_total
```

### Data Metrics

```promql
# Data write operations
nerve_data_write_total

# Data write errors
nerve_data_write_errors_total

# Data read operations
nerve_data_read_total
```

## Useful Queries

### Agent Status

```promql
# Agent online ratio
nerve_agent_online / nerve_agent_total * 100

# Agent offline count
nerve_agent_offline

# Heartbeat rate per agent
rate(nerve_agent_heartbeat_total[5m])
```

### Task Performance

```promql
# Task success rate
rate(nerve_task_success_total[5m]) / rate(nerve_task_total[5m]) * 100

# Task failure rate
rate(nerve_task_failed_total[5m]) / rate(nerve_task_total[5m]) * 100

# Average task duration
histogram_quantile(0.5, rate(nerve_task_duration_seconds_bucket[5m]))
```

### API Performance

```promql
# API request rate
rate(nerve_api_requests_total[5m])

# API error rate
rate(nerve_api_request_errors_total[5m])

# Average API request duration
rate(nerve_api_request_duration_seconds_sum[5m]) / rate(nerve_api_request_duration_seconds_count[5m])
```

## Alerting Rules

Create `alerts.yml`:

```yaml
groups:
  - name: nerve_alerts
    interval: 30s
    rules:
      # Alert when agents go offline
      - alert: AgentDown
        expr: nerve_agent_offline > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "{{ $value }} agent(s) are offline"
          description: "Check agent status and network connectivity"

      # Alert on high task failure rate
      - alert: HighTaskFailureRate
        expr: rate(nerve_task_failed_total[5m]) / rate(nerve_task_total[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Task failure rate is above 10%"
          description: "Investigate task execution issues"

      # Alert on high heartbeat error rate
      - alert: HighHeartbeatErrorRate
        expr: rate(nerve_agent_heartbeat_errors_total[5m]) / rate(nerve_agent_heartbeat_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Heartbeat error rate is above 5%"
          description: "Network or agent issues detected"

      # Alert on high API error rate
      - alert: HighAPIErrorRate
        expr: rate(nerve_api_request_errors_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API error rate"
          description: "Server issues detected"
```

## Grafana Dashboard

Import the provided Grafana dashboard from `deploy/prometheus/grafana-dashboard.json`.

### Dashboard Panels

1. **Agent Overview**
   - Total Agents
   - Online Agents
   - Offline Agents
   - Agent Online Ratio

2. **Task Performance**
   - Task Success Rate
   - Task Failure Rate
   - Task Duration Distribution
   - Tasks per Minute

3. **API Performance**
   - API Request Rate
   - API Response Time
   - API Error Rate
   - Top Endpoints

4. **System Health**
   - Heartbeat Rate
   - System Info Update Rate
   - Data Write/Read Operations

## Integration Steps

### For Existing Prometheus

1. Add scrape config to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'nerve-server'
    scrape_interval: 15s
    static_configs:
      - targets: ['nerve-server:9090']
```

2. Reload Prometheus:

```bash
curl -X POST http://localhost:9090/-/reload
```

### For New Prometheus Setup

1. Copy configuration files:

```bash
cp deploy/prometheus/prometheus.yml /etc/prometheus/
cp deploy/prometheus/alerts.yml /etc/prometheus/
```

2. Start Prometheus:

```bash
docker run -d \
  -p 9090:9090 \
  -v /etc/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
  -v /etc/prometheus/alerts.yml:/etc/prometheus/alerts.yml \
  prom/prometheus
```

## Monitoring Best Practices

1. **Set up alerting** for critical metrics
2. **Dashboard visualization** for real-time monitoring
3. **Log aggregation** alongside metrics
4. **Regular health checks** on metrics collection
5. **Capacity planning** based on historical data

## Troubleshooting

### Metrics not appearing

1. Check Nerve server is running with `--metrics-addr` flag
2. Verify metrics endpoint is accessible: `curl http://localhost:9090/metrics`
3. Check Prometheus configuration and targets
4. Verify network connectivity

### High cardinality

If you're experiencing high cardinality issues:

1. Reduce label dimensions
2. Use recording rules for aggregations
3. Consider sampling for high-volume metrics

## Next Steps

- Set up AlertManager for notifications
- Configure Grafana dashboards
- Implement custom metrics as needed
- Add business-specific metrics

