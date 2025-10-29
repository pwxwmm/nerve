// Package core provides task execution functionality with timeout protection.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// TaskExecutor handles task execution with timeout
type TaskExecutor struct {
	defaultTimeout time.Duration
}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor(defaultTimeout time.Duration) *TaskExecutor {
	return &TaskExecutor{
		defaultTimeout: defaultTimeout,
	}
}

// ExecuteCommand executes a shell command with timeout
func (e *TaskExecutor) ExecuteCommand(command string, timeout int) (TaskResult, error) {
	var result TaskResult
	
	// Determine timeout
	t := e.defaultTimeout
	if timeout > 0 {
		t = time.Duration(timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	// Execute command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	// Capture output
	output, err := cmd.CombinedOutput()
	
	result.Success = (err == nil)
	result.Output = string(output)
	
	if err != nil {
		if err == context.DeadlineExceeded {
			result.Error = fmt.Sprintf("command timeout after %v", t)
		} else {
			result.Error = err.Error()
		}
	}

	return result, err
}

// ExecuteScript executes a script file with timeout
func (e *TaskExecutor) ExecuteScript(script string, timeout int) (TaskResult, error) {
	var result TaskResult
	
	// Determine timeout
	t := e.defaultTimeout
	if timeout > 0 {
		t = time.Duration(timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	// Write script to temporary file
	tmpFile, err := os.CreateTemp("", "nerve-script-*.sh")
	if err != nil {
		result.Error = err.Error()
		return result, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		result.Error = err.Error()
		return result, err
	}
	tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		result.Error = err.Error()
		return result, err
	}

	// Execute script
	cmd := exec.CommandContext(ctx, "/bin/bash", tmpFile.Name())
	output, err := cmd.CombinedOutput()

	result.Success = (err == nil)
	result.Output = string(output)

	if err != nil {
		if err == context.DeadlineExceeded {
			result.Error = fmt.Sprintf("script timeout after %v", t)
		} else {
			result.Error = err.Error()
		}
	}

	return result, err
}

// ExecuteHook executes a hook plugin with timeout
func (e *TaskExecutor) ExecuteHook(pluginManager *PluginManager, pluginName string, params map[string]interface{}, timeout int) (TaskResult, error) {
	var result TaskResult
	
	// Determine timeout
	t := e.defaultTimeout
	if timeout > 0 {
		t = time.Duration(timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	// Execute plugin in goroutine with timeout
	done := make(chan struct{})
	var pluginResult map[string]interface{}
	var pluginErr error

	go func() {
		defer close(done)
		pluginResult, pluginErr = pluginManager.ExecutePlugin(pluginName, params)
	}()

	select {
	case <-ctx.Done():
		result.Success = false
		result.Error = fmt.Sprintf("plugin timeout after %v", t)
		return result, ctx.Err()
	case <-done:
		result.Success = (pluginErr == nil)
		if pluginErr != nil {
			result.Error = pluginErr.Error()
		} else {
			result.Output = fmt.Sprintf("%+v", pluginResult)
		}
		return result, pluginErr
	}
}

