package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

// Plugin interface defines the contract for all plugins
type Plugin interface {
	// Core plugin information
	GetName() string
	GetVersion() string
	GetDescription() string
	GetAuthor() string

	// Lifecycle hooks
	Initialize(config map[string]interface{}) error
	Execute(ctx *PluginContext) error
	Cleanup() error

	// Capabilities
	GetSupportedFrameworks() []string
	GetSupportedEvents() []PluginEventType
	GetDependencies() []PluginDependency

	// Configuration
	GetConfigSchema() map[string]interface{}
	ValidateConfig(config map[string]interface{}) error
}

// PluginEventType represents different plugin event types
type PluginEventType string

const (
	EventBeforeScan    PluginEventType = "before_scan"
	EventAfterScan     PluginEventType = "after_scan"
	EventBeforeGen     PluginEventType = "before_generation"
	EventAfterGen      PluginEventType = "after_generation"
	EventRouteGenerated PluginEventType = "route_generated"
	EventValidationError PluginEventType = "validation_error"
	EventMiddleware     PluginEventType = "middleware"
	EventCustom        PluginEventType = "custom"
)

// PluginContext provides context for plugin execution
type PluginContext struct {
	EventType     PluginEventType          `json:"event_type"`
	Generator     *APIGenerator            `json:"generator"`
	Package       *PackageInfo             `json:"package"`
	Struct        *StructInfo              `json:"struct,omitempty"`
	Route         *APIRoute                `json:"route,omitempty"`
	Validation    *ValidationResult        `json:"validation,omitempty"`
	Config        map[string]interface{}   `json:"config"`
	Data          map[string]interface{}   `json:"data"`
	Metadata      map[string]interface{}   `json:"metadata"`
	RequestID     string                   `json:"request_id"`
	Timestamp     int64                    `json:"timestamp"`
}

// PluginDependency represents a plugin dependency
type PluginDependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
}

// PluginConfig represents plugin configuration
type PluginConfig struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	Priority    int                    `json:"priority"`
	Order       int                    `json:"order"`
	Constraints []string               `json:"constraints"`
}

// PluginManager manages plugin loading, execution, and lifecycle
type PluginManager struct {
	plugins    map[string]Plugin
	configs    map[string]*PluginConfig
	hooks      map[PluginEventType][]Plugin
	config     *PluginManagerConfig
	mu         sync.RWMutex
	initialized bool
}

// PluginManagerConfig contains configuration for the plugin manager
type PluginManagerConfig struct {
	PluginDir       string   `json:"plugin_dir"`
	AutoLoad        bool     `json:"auto_load"`
	EnabledPlugins  []string `json:"enabled_plugins"`
	DisabledPlugins []string `json:"disabled_plugins"`
	SecurityMode    bool     `json:"security_mode"`
	MaxPlugins      int      `json:"max_plugins"`
	SandboxMode     bool     `json:"sandbox_mode"`
}

// PluginMetadata contains plugin metadata from plugin files
type PluginMetadata struct {
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Description      string            `json:"description"`
	Author           string            `json:"author"`
	Homepage         string            `json:"homepage"`
	Repository       string            `json:"repository"`
	License          string            `json:"license"`
	MainFile         string            `json:"main_file"`
	Dependencies     []PluginDependency `json:"dependencies"`
	SupportedFrameworks []string       `json:"supported_frameworks"`
	SupportedEvents []PluginEventType   `json:"supported_events"`
	ConfigSchema     map[string]interface{} `json:"config_schema"`
	Tags             []string          `json:"tags"`
	Category         string            `json:"category"`
}

// NewPluginManager creates a new plugin manager instance
func NewPluginManager(config *PluginManagerConfig) *PluginManager {
	pm := &PluginManager{
		plugins: make(map[string]Plugin),
		configs: make(map[string]*PluginConfig),
		hooks:   make(map[PluginEventType][]Plugin),
		config:  config,
	}

	// Initialize hooks for all event types
	eventTypes := []PluginEventType{
		EventBeforeScan,
		EventAfterScan,
		EventBeforeGen,
		EventAfterGen,
		EventRouteGenerated,
		EventValidationError,
		EventMiddleware,
		EventCustom,
	}

	for _, eventType := range eventTypes {
		pm.hooks[eventType] = []Plugin{}
	}

	return pm
}

// RegisterPlugin registers a built-in plugin directly
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if already registered
	name := plugin.GetName()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin already registered: %s", name)
	}

	// Store plugin
	pm.plugins[name] = plugin

	return nil
}

// LoadPlugin loads a plugin from a file or directory
func (pm *PluginManager) LoadPlugin(path string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if already loaded
	if _, exists := pm.plugins[path]; exists {
		return fmt.Errorf("plugin already loaded: %s", path)
	}

	// Load plugin metadata
	metadata, err := pm.loadPluginMetadata(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin metadata: %v", err)
	}

	// Check dependencies
	if err := pm.checkDependencies(metadata.Dependencies); err != nil {
		return fmt.Errorf("dependency check failed: %v", err)
	}

	// Load the plugin
	plugin, err := pm.loadPluginFromFile(path, metadata)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %v", err)
	}

	// Store plugin
	pm.plugins[metadata.Name] = plugin

	// Register hooks
	for _, eventType := range metadata.SupportedEvents {
		pm.hooks[eventType] = append(pm.hooks[eventType], plugin)
	}

	// Create default config if not exists
	if _, exists := pm.configs[metadata.Name]; !exists {
		pm.configs[metadata.Name] = &PluginConfig{
			Name:    metadata.Name,
			Enabled: true,
			Config:  make(map[string]interface{}),
			Order:   len(pm.plugins),
		}
	}

	return nil
}

// loadPluginMetadata loads plugin metadata from plugin.json file
func (pm *PluginManager) loadPluginMetadata(path string) (*PluginMetadata, error) {
	metadataFile := filepath.Join(path, "plugin.json")
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		// Try plugin.yaml
		metadataFile = filepath.Join(path, "plugin.yaml")
		if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("no plugin metadata file found in %s", path)
		}
	}

	data, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, err
	}

	var metadata PluginMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// loadPluginFromFile loads a Go plugin from a .so file
func (pm *PluginManager) loadPluginFromFile(path string, metadata *PluginMetadata) (Plugin, error) {
	// Determine plugin file path
	pluginFile := filepath.Join(path, metadata.MainFile)
	if filepath.Ext(pluginFile) != ".so" {
		pluginFile += ".so"
	}

	// Open the plugin
	p, err := plugin.Open(pluginFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin file: %v", err)
	}

	// Look up the plugin symbol
	sym, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("failed to find NewPlugin symbol: %v", err)
	}

	// Assert the symbol to the correct type
	newPluginFunc, ok := sym.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("unexpected type from module symbol")
	}

	// Create plugin instance
	plugin := newPluginFunc()

	return plugin, nil
}

// checkDependencies checks if all plugin dependencies are satisfied
func (pm *PluginManager) checkDependencies(dependencies []PluginDependency) error {
	for _, dep := range dependencies {
		if dep.Type == "plugin" {
			if _, exists := pm.plugins[dep.Name]; !exists {
				return fmt.Errorf("missing plugin dependency: %s", dep.Name)
			}
		}
		// Add more dependency checks (go modules, external services, etc.)
	}
	return nil
}

// ExecutePlugins executes all plugins registered for a specific event type
func (pm *PluginManager) ExecutePlugins(eventType PluginEventType, ctx *PluginContext) error {
	pm.mu.RLock()
	plugins, exists := pm.hooks[eventType]
	pm.mu.RUnlock()

	if !exists {
		return nil
	}

	ctx.EventType = eventType

	// Sort plugins by order/priority
	sortedPlugins := pm.sortPluginsByPriority(plugins)

	for _, plugin := range sortedPlugins {
		pluginName := plugin.GetName()

		// Check if plugin is enabled
		config, exists := pm.configs[pluginName]
		if exists && !config.Enabled {
			continue
		}

		// Create plugin context with plugin-specific config
		pluginCtx := *ctx
		if config != nil {
			pluginCtx.Config = config.Config
		}

		// Execute plugin
		if err := plugin.Execute(&pluginCtx); err != nil {
			if pm.config.SandboxMode {
				// Log error but continue in sandbox mode
				fmt.Printf("Plugin %s execution failed: %v\n", pluginName, err)
				continue
			}
			return fmt.Errorf("plugin %s execution failed: %v", pluginName, err)
		}

		// Update context with plugin data
		if pluginCtx.Data != nil {
			for k, v := range pluginCtx.Data {
				ctx.Data[k] = v
			}
		}
	}

	return nil
}

// sortPluginsByPriority sorts plugins by their priority and order
func (pm *PluginManager) sortPluginsByPriority(plugins []Plugin) []Plugin {
	// Create a slice of plugin-info pairs
	type pluginInfo struct {
		plugin   Plugin
		priority int
		order    int
	}

	var infos []pluginInfo
	for _, p := range plugins {
		config, exists := pm.configs[p.GetName()]
		priority := 0
		order := 0
		if exists {
			priority = config.Priority
			order = config.Order
		}
		infos = append(infos, pluginInfo{plugin: p, priority: priority, order: order})
	}

	// Sort by priority (descending), then by order (ascending)
	for i := 0; i < len(infos); i++ {
		for j := i + 1; j < len(infos); j++ {
			if infos[i].priority < infos[j].priority ||
				(infos[i].priority == infos[j].priority && infos[i].order > infos[j].order) {
				infos[i], infos[j] = infos[j], infos[i]
			}
		}
	}

	// Extract sorted plugins
	result := make([]Plugin, len(infos))
	for i, info := range infos {
		result[i] = info.plugin
	}

	return result
}

// EnablePlugin enables a plugin
func (pm *PluginManager) EnablePlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	config, exists := pm.configs[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	config.Enabled = true
	return nil
}

// DisablePlugin disables a plugin
func (pm *PluginManager) DisablePlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	config, exists := pm.configs[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	config.Enabled = false
	return nil
}

// ConfigurePlugin configures a plugin
func (pm *PluginManager) ConfigurePlugin(name string, config map[string]interface{}) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Validate config
	if err := plugin.ValidateConfig(config); err != nil {
		return fmt.Errorf("config validation failed: %v", err)
	}

	// Store config
	if pm.configs[name] == nil {
		pm.configs[name] = &PluginConfig{Name: name}
	}
	pm.configs[name].Config = config

	return nil
}

// GetPlugin returns a plugin by name
func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[name]
	return plugin, exists
}

// ListPlugins returns all loaded plugins
func (pm *PluginManager) ListPlugins() map[string]Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]Plugin)
	for k, v := range pm.plugins {
		result[k] = v
	}
	return result
}

// LoadAllPlugins loads all plugins from the plugin directory
func (pm *PluginManager) LoadAllPlugins() error {
	if pm.config.PluginDir == "" {
		return nil
	}

	return filepath.Walk(pm.config.PluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that are not plugin directories
		if !info.IsDir() || path == pm.config.PluginDir {
			return nil
		}

		// Check if this is a plugin directory (has plugin.json)
		if _, err := os.Stat(filepath.Join(path, "plugin.json")); os.IsNotExist(err) {
			if _, err := os.Stat(filepath.Join(path, "plugin.yaml")); os.IsNotExist(err) {
				return nil
			}
		}

		// Load plugin
		if err := pm.LoadPlugin(path); err != nil {
			fmt.Printf("Warning: Failed to load plugin from %s: %v\n", path, err)
		}

		return nil
	})
}

// InitializePlugins initializes all loaded plugins
func (pm *PluginManager) InitializePlugins() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, plugin := range pm.plugins {
		config, exists := pm.configs[name]
		if !exists || !config.Enabled {
			continue
		}

		pluginConfig := make(map[string]interface{})
		if config != nil {
			pluginConfig = config.Config
		}

		if err := plugin.Initialize(pluginConfig); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %v", name, err)
		}
	}

	pm.initialized = true
	return nil
}

// CleanupPlugins cleans up all loaded plugins
func (pm *PluginManager) CleanupPlugins() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var errors []error
	for name, plugin := range pm.plugins {
		if err := plugin.Cleanup(); err != nil {
			errors = append(errors, fmt.Errorf("plugin %s cleanup failed: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// GeneratePluginCode generates boilerplate code for creating a new plugin
func (pm *PluginManager) GeneratePluginCode(name, description, author string) error {
	pluginDir := filepath.Join(pm.config.PluginDir, name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	// Generate plugin.go file
	pluginGo := fmt.Sprintf(`package main

import (
	"fmt"
)

type %sPlugin struct {
	config map[string]interface{}
}

func NewPlugin() Plugin {
	return &%sPlugin{}
}

func (p *%sPlugin) GetName() string {
	return "%s"
}

func (p *%sPlugin) GetVersion() string {
	return "1.0.0"
}

func (p *%sPlugin) GetDescription() string {
	return "%s"
}

func (p *%sPlugin) GetAuthor() string {
	return "%s"
}

func (p *%sPlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	return nil
}

func (p *%sPlugin) Execute(ctx *PluginContext) error {
	switch ctx.EventType {
	case EventBeforeScan:
		return p.handleBeforeScan(ctx)
	case EventAfterScan:
		return p.handleAfterScan(ctx)
	case EventBeforeGen:
		return p.handleBeforeGeneration(ctx)
	case EventAfterGen:
		return p.handleAfterGeneration(ctx)
	default:
		return nil
	}
}

func (p *%sPlugin) Cleanup() error {
	return nil
}

func (p *%sPlugin) GetSupportedFrameworks() []string {
	return []string{"gin", "echo", "chi", "fiber"}
}

func (p *%sPlugin) GetSupportedEvents() []PluginEventType {
	return []PluginEventType{
		EventBeforeScan,
		EventAfterScan,
		EventBeforeGen,
		EventAfterGen,
	}
}

func (p *%sPlugin) GetDependencies() []PluginDependency {
	return []PluginDependency{}
}

func (p *%sPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"enabled": map[string]interface{}{
				"type": "boolean",
				"default": true,
			},
		},
	}
}

func (p *%sPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func (p *%sPlugin) handleBeforeScan(ctx *PluginContext) error {
	fmt.Printf("%%s: Before scan hook\n", p.GetName())
	return nil
}

func (p *%sPlugin) handleAfterScan(ctx *PluginContext) error {
	fmt.Printf("%%s: After scan hook\n", p.GetName())
	return nil
}

func (p *%sPlugin) handleBeforeGeneration(ctx *PluginContext) error {
	fmt.Printf("%%s: Before generation hook\n", p.GetName())
	return nil
}

func (p *%sPlugin) handleAfterGeneration(ctx *PluginContext) error {
	fmt.Printf("%%s: After generation hook\n", p.GetName())
	return nil
}
`, name, name, name, name, name, name, name, description, author, name, name, name, name, name)

	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.go"), []byte(pluginGo), 0644); err != nil {
		return err
	}

	// Generate plugin.json metadata
	metadata := map[string]interface{}{
		"name":                name,
		"version":             "1.0.0",
		"description":         description,
		"author":              author,
		"main_file":           "plugin.so",
		"supported_frameworks": []string{"gin", "echo", "chi", "fiber"},
		"supported_events":    []string{"before_scan", "after_scan", "before_generation", "after_generation"},
		"config_schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"enabled": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		"category": "custom",
		"tags":     []string{"custom", "generated"},
	}

	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.json"), metadataBytes, 0644); err != nil {
		return err
	}

	// Generate Makefile for building the plugin
	makefile := fmt.Sprintf(`# Makefile for %s plugin
.PHONY: build clean install

PLUGIN_NAME=%s
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

build:
	go build -buildmode=plugin -o $(PLUGIN_NAME).so plugin.go

clean:
	rm -f $(PLUGIN_NAME).so

install: build
	cp $(PLUGIN_NAME).so ../../plugins/

dev: clean build
	cp $(PLUGIN_NAME).so ../../plugins/
`, name, name)

	if err := os.WriteFile(filepath.Join(pluginDir, "Makefile"), []byte(makefile), 0644); err != nil {
		return err
	}

	// Generate README.md
	readme := fmt.Sprintf(`# %s Plugin

%s

## Author
%s

## Installation

1. Build the plugin:
   ` + "```bash" + `
   make build
   ` + "```" + `

2. Install to plugins directory:
   ` + "```bash" + `
   make install
   ` + "```" + `

## Configuration

This plugin supports the following configuration options:

- ` + "`enabled`" + ` (boolean): Enable or disable the plugin (default: true)

## Events

This plugin responds to the following events:

- before_scan
- after_scan
- before_generation
- after_generation

## Development

To rebuild the plugin:

` + "```bash" + `
make dev
` + "```" + `

To clean build artifacts:

` + "```bash" + `
make clean
` + "```" + `
`, name, description, author)

	if err := os.WriteFile(filepath.Join(pluginDir, "README.md"), []byte(readme), 0644); err != nil {
		return err
	}

	return nil
}

// Built-in plugins

type LoggingPlugin struct {
	config map[string]interface{}
}

func NewLoggingPlugin() Plugin {
	return &LoggingPlugin{}
}

func (p *LoggingPlugin) GetName() string { return "logging" }
func (p *LoggingPlugin) GetVersion() string { return "1.0.0" }
func (p *LoggingPlugin) GetDescription() string { return "Built-in logging plugin" }
func (p *LoggingPlugin) GetAuthor() string { return "GoFastAPI" }

func (p *LoggingPlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	return nil
}

func (p *LoggingPlugin) Execute(ctx *PluginContext) error {
	enabled, _ := p.config["enabled"].(bool)
	if !enabled {
		return nil
	}

	fmt.Printf("[PLUGIN] %s: %s event\n", p.GetName(), string(ctx.EventType))
	return nil
}

func (p *LoggingPlugin) Cleanup() error { return nil }

func (p *LoggingPlugin) GetSupportedFrameworks() []string {
	return []string{"gin", "echo", "chi", "fiber"}
}

func (p *LoggingPlugin) GetSupportedEvents() []PluginEventType {
	return []PluginEventType{
		EventBeforeScan,
		EventAfterScan,
		EventBeforeGen,
		EventAfterGen,
		EventRouteGenerated,
	}
}

func (p *LoggingPlugin) GetDependencies() []PluginDependency { return nil }

func (p *LoggingPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"enabled": map[string]interface{}{
				"type":    "boolean",
				"default": true,
			},
			"level": map[string]interface{}{
				"type":    "string",
				"default": "info",
			},
		},
	}
}

func (p *LoggingPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// MetricsPlugin for collecting API generation metrics
type MetricsPlugin struct {
	config map[string]interface{}
	metrics map[string]interface{}
}

func NewMetricsPlugin() Plugin {
	return &MetricsPlugin{
		metrics: make(map[string]interface{}),
	}
}

func (p *MetricsPlugin) GetName() string { return "metrics" }
func (p *MetricsPlugin) GetVersion() string { return "1.0.0" }
func (p *MetricsPlugin) GetDescription() string { return "Built-in metrics collection plugin" }
func (p *MetricsPlugin) GetAuthor() string { return "GoFastAPI" }

func (p *MetricsPlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	return nil
}

func (p *MetricsPlugin) Execute(ctx *PluginContext) error {
	// Collect metrics based on event type
	switch ctx.EventType {
	case EventAfterScan:
		p.metrics["packages_scanned"] = ctx.Metadata["package_count"]
	case EventAfterGen:
		p.metrics["routes_generated"] = ctx.Metadata["route_count"]
	case EventRouteGenerated:
		if count, ok := p.metrics["routes_generated"].(int); ok {
			p.metrics["routes_generated"] = count + 1
		} else {
			p.metrics["routes_generated"] = 1
		}
	}

	// Store metrics in context
	if ctx.Data == nil {
		ctx.Data = make(map[string]interface{})
	}
	ctx.Data["metrics"] = p.metrics

	return nil
}

func (p *MetricsPlugin) Cleanup() error {
	p.metrics = nil
	return nil
}

func (p *MetricsPlugin) GetSupportedFrameworks() []string {
	return []string{"gin", "echo", "chi", "fiber"}
}

func (p *MetricsPlugin) GetSupportedEvents() []PluginEventType {
	return []PluginEventType{
		EventAfterScan,
		EventAfterGen,
		EventRouteGenerated,
	}
}

func (p *MetricsPlugin) GetDependencies() []PluginDependency { return nil }

func (p *MetricsPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"enabled": map[string]interface{}{
				"type":    "boolean",
				"default": true,
			},
			"output_format": map[string]interface{}{
				"type":    "string",
				"default": "json",
			},
		},
	}
}

func (p *MetricsPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// Global plugin manager instance
var globalPluginManager *PluginManager

// GetPluginManager returns the global plugin manager instance
func GetPluginManager() *PluginManager {
	if globalPluginManager == nil {
		config := &PluginManagerConfig{
			PluginDir:       "./plugins",
			AutoLoad:        true,
			SecurityMode:    true,
			MaxPlugins:      50,
			SandboxMode:     true,
			EnabledPlugins:  []string{"logging", "metrics"},
			DisabledPlugins: []string{},
		}
		globalPluginManager = NewPluginManager(config)

		// Register built-in plugins
		globalPluginManager.RegisterPlugin(NewLoggingPlugin())
		globalPluginManager.RegisterPlugin(NewMetricsPlugin())

		// Auto-load plugins from directory
		if config.AutoLoad {
			globalPluginManager.LoadAllPlugins()
		}
	}
	return globalPluginManager
}