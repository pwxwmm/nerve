// Package cluster provides cluster management functionality for multi-cluster support.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package cluster

import (
	"fmt"
	"sync"
	"time"
)

// ClusterManager manages multiple clusters
type ClusterManager struct {
	clusters map[string]*Cluster
	mutex    sync.RWMutex
}

// Cluster represents a cluster configuration
type Cluster struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Agents      []string               `json:"agents"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NewClusterManager creates a new cluster manager
func NewClusterManager() *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*Cluster),
	}
}

// AddCluster adds a new cluster
func (cm *ClusterManager) AddCluster(cluster *Cluster) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.clusters[cluster.ID]; exists {
		return fmt.Errorf("cluster %s already exists", cluster.ID)
	}

	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()
	cm.clusters[cluster.ID] = cluster

	return nil
}

// GetCluster retrieves a cluster by ID
func (cm *ClusterManager) GetCluster(id string) (*Cluster, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cluster, exists := cm.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", id)
	}

	return cluster, nil
}

// ListClusters returns all clusters
func (cm *ClusterManager) ListClusters() []*Cluster {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var clusters []*Cluster
	for _, cluster := range cm.clusters {
		clusters = append(clusters, cluster)
	}

	return clusters
}

// UpdateCluster updates an existing cluster
func (cm *ClusterManager) UpdateCluster(id string, updates map[string]interface{}) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cluster, exists := cm.clusters[id]
	if !exists {
		return fmt.Errorf("cluster %s not found", id)
	}

	// Update fields
	if name, ok := updates["name"].(string); ok {
		cluster.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		cluster.Description = desc
	}
	if config, ok := updates["config"].(map[string]interface{}); ok {
		cluster.Config = config
	}
	if agents, ok := updates["agents"].([]string); ok {
		cluster.Agents = agents
	}

	cluster.UpdatedAt = time.Now()

	return nil
}

// DeleteCluster removes a cluster
func (cm *ClusterManager) DeleteCluster(id string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.clusters[id]; !exists {
		return fmt.Errorf("cluster %s not found", id)
	}

	delete(cm.clusters, id)
	return nil
}

// AddAgentToCluster adds an agent to a cluster
func (cm *ClusterManager) AddAgentToCluster(clusterID, agentID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cluster, exists := cm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	// Check if agent already exists
	for _, agent := range cluster.Agents {
		if agent == agentID {
			return fmt.Errorf("agent %s already in cluster %s", agentID, clusterID)
		}
	}

	cluster.Agents = append(cluster.Agents, agentID)
	cluster.UpdatedAt = time.Now()

	return nil
}

// RemoveAgentFromCluster removes an agent from a cluster
func (cm *ClusterManager) RemoveAgentFromCluster(clusterID, agentID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cluster, exists := cm.clusters[clusterID]
	if !exists {
		return fmt.Errorf("cluster %s not found", clusterID)
	}

	// Find and remove agent
	for i, agent := range cluster.Agents {
		if agent == agentID {
			cluster.Agents = append(cluster.Agents[:i], cluster.Agents[i+1:]...)
			cluster.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("agent %s not found in cluster %s", agentID, clusterID)
}

// GetClusterStats returns statistics for a cluster
func (cm *ClusterManager) GetClusterStats(clusterID string) (map[string]interface{}, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cluster, exists := cm.clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}

	// TODO: Get actual agent statistics
	stats := map[string]interface{}{
		"total_agents":   len(cluster.Agents),
		"online_agents":  0, // TODO: Calculate from agent status
		"offline_agents": 0, // TODO: Calculate from agent status
		"total_tasks":    0, // TODO: Calculate from task history
		"last_activity":  cluster.UpdatedAt,
	}

	return stats, nil
}

// GetAgentClusters returns clusters that contain the specified agent
func (cm *ClusterManager) GetAgentClusters(agentID string) []*Cluster {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var clusters []*Cluster
	for _, cluster := range cm.clusters {
		for _, agent := range cluster.Agents {
			if agent == agentID {
				clusters = append(clusters, cluster)
				break
			}
		}
	}

	return clusters
}

