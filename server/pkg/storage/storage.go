package storage

import (
	"sync"
	"time"
)

// Storage defines storage interface
type Storage interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Delete(key string) error
	List() map[string]interface{}
}

// InMemory is an in-memory storage implementation
type InMemory struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewInMemory creates a new in-memory storage
func NewInMemory() Storage {
	return &InMemory{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value
func (s *InMemory) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	value, ok := s.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	
	return value, nil
}

// Set stores a value
func (s *InMemory) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.data[key] = value
	return nil
}

// Delete removes a value
func (s *InMemory) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.data, key)
	return nil
}

// List returns all key-value pairs
func (s *InMemory) List() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make(map[string]interface{})
	for k, v := range s.data {
		result[k] = v
	}
	
	return result
}

// AgentRecord represents an agent record in storage
type AgentRecord struct {
	ID        string
	Info      interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "not found"
}

