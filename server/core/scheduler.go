package core

import (
	"sync"
	"time"

	"github.com/nerve/server/pkg/log"
)

// Scheduler manages task scheduling
type Scheduler struct {
	mu       sync.RWMutex
	registry *Registry
	logger   log.Logger
	tasks    map[string]*Task
}

// NewScheduler creates a new scheduler
func NewScheduler(registry *Registry, logger log.Logger) *Scheduler {
	return &Scheduler{
		registry: registry,
		logger:   logger,
		tasks:    make(map[string]*Task),
	}
}

// SubmitTask submits a task for execution
func (s *Scheduler) SubmitTask(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.Status = "pending"
	s.tasks[task.ID] = task
	
	s.logger.Infof("Task submitted: ID=%s, AgentID=%s, Type=%s", 
		task.ID, task.AgentID, task.Type)
}

// GetPendingTasks returns pending tasks for an agent
func (s *Scheduler) GetPendingTasks(agentID string) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*Task
	for _, task := range s.tasks {
		if task.AgentID == agentID && task.Status == "pending" {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// MarkTaskDone marks a task as completed
func (s *Scheduler) MarkTaskDone(taskID string, success bool, output string, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return
	}

	task.Status = "completed"
	if success {
		s.logger.Infof("Task completed: %s", taskID)
	} else {
		s.logger.Errorf("Task failed: %s - %s", taskID, errMsg)
	}
}

// GetTasksByStatus returns tasks filtered by status
func (s *Scheduler) GetTasksByStatus(status string) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*Task
	for _, task := range s.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// ScheduleHook schedules a hook execution
func (s *Scheduler) ScheduleHook(agentID, plugin string, params map[string]interface{}) {
	task := &Task{
		ID:      generateTaskID(),
		AgentID: agentID,
		Type:    "hook",
		Plugin:  plugin,
		Params:  params,
		Status:  "pending",
	}

	s.SubmitTask(task)
}

// ScheduleCommand schedules a command execution
func (s *Scheduler) ScheduleCommand(agentID, command string, timeout int) {
	task := &Task{
		ID:      generateTaskID(),
		AgentID: agentID,
		Type:    "command",
		Command: command,
		Timeout: timeout,
		Status:  "pending",
	}

	s.SubmitTask(task)
}

func generateTaskID() string {
	return time.Now().Format("20060102150405")
}

