// Package core provides plugin management functionality for dynamic hook loading.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

// HookPlugin defines the interface for hook plugins
type HookPlugin interface {
	Name() string
	Version() string
	Execute(params map[string]interface{}) (map[string]interface{}, error)
}

// PluginManager manages hook plugins
type PluginManager struct {
	plugins map[string]HookPlugin
	mutex   sync.RWMutex
	path    string
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginPath string) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]HookPlugin),
		path:    pluginPath,
	}
}

// LoadPlugin loads a plugin from file
func (pm *PluginManager) LoadPlugin(pluginFile string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Load the plugin
	p, err := plugin.Open(filepath.Join(pm.path, pluginFile))
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %v", pluginFile, err)
	}

	// Look up the symbol
	symbol, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin %s does not export Plugin symbol: %v", pluginFile, err)
	}

	// Type assert to HookPlugin
	hookPlugin, ok := symbol.(HookPlugin)
	if !ok {
		return fmt.Errorf("plugin %s does not implement HookPlugin interface", pluginFile)
	}

	// Register the plugin
	pm.plugins[hookPlugin.Name()] = hookPlugin
	return nil
}

// LoadPlugins loads all plugins from the plugin directory
func (pm *PluginManager) LoadPlugins() error {
	// Create plugin directory if it doesn't exist
	if err := os.MkdirAll(pm.path, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// Read plugin directory
	files, err := os.ReadDir(pm.path)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %v", err)
	}

	// Load each .so file
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".so" {
			if err := pm.LoadPlugin(file.Name()); err != nil {
				fmt.Printf("Warning: failed to load plugin %s: %v\n", file.Name(), err)
			}
		}
	}

	return nil
}

// ExecutePlugin executes a plugin by name
func (pm *PluginManager) ExecutePlugin(name string, params map[string]interface{}) (map[string]interface{}, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin.Execute(params)
}

// ListPlugins returns a list of loaded plugins
func (pm *PluginManager) ListPlugins() []map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var plugins []map[string]interface{}
	for name, plugin := range pm.plugins {
		plugins = append(plugins, map[string]interface{}{
			"name":    name,
			"version": plugin.Version(),
		})
	}

	return plugins
}

// PluginConfig represents plugin configuration
type PluginConfig struct {
	Name    string                 `json:"name"`
	Version string                 `json:"version"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

// LoadPluginConfig loads plugin configuration from file
func (pm *PluginManager) LoadPluginConfig(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read plugin config: %v", err)
	}

	var configs []PluginConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse plugin config: %v", err)
	}

	// Apply configurations
	for _, config := range configs {
		if plugin, exists := pm.plugins[config.Name]; exists {
			// TODO: Apply configuration to plugin
			_ = plugin
		}
	}

	return nil
}

// SavePluginConfig saves plugin configuration to file
func (pm *PluginManager) SavePluginConfig(configFile string) error {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var configs []PluginConfig
	for name, plugin := range pm.plugins {
		configs = append(configs, PluginConfig{
			Name:    name,
			Version: plugin.Version(),
			Enabled: true,
			Config:  make(map[string]interface{}),
		})
	}

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugin config: %v", err)
	}

	return os.WriteFile(configFile, data, 0644)
}

