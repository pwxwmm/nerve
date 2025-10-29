# Nerve Hook Plugin System

## Overview

The hook plugin system allows you to extend Nerve agents with custom functionality. Plugins can be registered and executed remotely from the center.

## Plugin Structure

A hook plugin is defined as a YAML configuration file:

```yaml
name: custom-monitor
version: 1.0.0
description: Custom monitoring script

trigger:
  type: scheduled  # scheduled, on-demand, event
  interval: 300    # seconds (for scheduled)

execution:
  type: script     # script, binary, http
  path: /opt/scripts/monitor.sh
  
parameters:
  - name: threshold
    type: number
    default: 80
    
output:
  format: json
```

## Plugin Types

### 1. Script Plugins

Execute shell scripts.

```yaml
name: disk-cleanup
execution:
  type: script
  path: /opt/scripts/cleanup.sh
  timeout: 300
```

### 2. Binary Plugins

Execute compiled binaries.

```yaml
name: custom-collector
execution:
  type: binary
  path: /usr/local/bin/custom-collector
```

### 3. HTTP Plugins

Call HTTP endpoints.

```yaml
name: webhook
execution:
  type: http
  url: https://api.example.com/hook
  method: POST
  headers:
    X-API-Key: secret-key
```

## Registering Plugins

### On Agent

```bash
# Copy plugin to agent
scp my-plugin.yaml agent:/etc/nerve/plugins/

# Restart agent
systemctl restart nerve-agent
```

### Via API

```bash
curl -X POST http://nerve-center:8080/api/plugins/register \
  -H "Content-Type: application/yaml" \
  -d @my-plugin.yaml
```

## Plugin Execution

Plugins are executed when:

1. **Scheduled Trigger**: Time-based execution
2. **On-Demand**: Manual trigger from center
3. **Event Trigger**: Based on system events

### Trigger from Center

```bash
curl -X POST http://nerve-center:8080/api/agents/{hostname}/hooks/execute \
  -H "Content-Type: application/json" \
  -d '{
    "plugin": "custom-monitor",
    "params": {"threshold": 90}
  }'
```

## Plugin Results

Plugin execution results are reported back to center:

```json
{
  "plugin": "custom-monitor",
  "success": true,
  "output": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Plugin Development

### Example Python Plugin

```python
#!/usr/bin/env python3
import sys
import json

def main():
    params = json.loads(sys.argv[1])
    
    # Plugin logic
    result = {
        "status": "ok",
        "data": {}
    }
    
    print(json.dumps(result))

if __name__ == "__main__":
    main()
```

### Example Shell Plugin

```bash
#!/bin/bash

# Parse parameters
PARAMS="$1"

# Execute logic
result=$(some-command)

# Output result
echo "{\"status\": \"ok\", \"result\": \"$result\"}"
```

## Built-in Plugins

### 1. System Info Collector

```yaml
name: system-info
execution:
  type: builtin
  handler: SystemInfoCollector
```

### 2. Log Collector

```yaml
name: log-collector
execution:
  type: builtin
  handler: LogCollector
```

### 3. Metric Exporter

```yaml
name: prometheus-exporter
execution:
  type: builtin
  handler: PrometheusExporter
```

## Best Practices

1. **Idempotency**: Plugins should be idempotent
2. **Timeout**: Always set appropriate timeouts
3. **Logging**: Log plugin execution details
4. **Error Handling**: Handle errors gracefully
5. **Security**: Validate inputs and outputs

## Security Considerations

1. **Sandboxing**: Plugins run in isolated context
2. **Permissions**: Use least privilege principle
3. **Validation**: Validate plugin code before execution
4. **Audit**: Log all plugin executions

## Monitoring

View plugin executions:

```bash
curl http://nerve-center:8080/api/agents/{hostname}/hooks/history
```

Response:
```json
[
  {
    "plugin": "custom-monitor",
    "status": "success",
    "executed_at": "2024-01-01T12:00:00Z",
    "duration": 1.23
  }
]
```

