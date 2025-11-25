package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestSuite is the main test suite for GoFastAPI
type TestSuite struct {
	suite.Suite
	generator   *APIGenerator
	testDataDir string
	tempDir     string
	config      *GeneratorConfig
}

// SetupSuite runs once before all tests
func (suite *TestSuite) SetupSuite() {
	suite.config = &GeneratorConfig{
		IncludePatterns: []string{"*.go"},
		ExcludePatterns: []string{"*_test.go", "vendor/*"},
		ScanAnnotations: true,
		AutoCRUD:        true,
		SmartMapping:    true,
		OutputDir:       "./test-output",
		PackageName:     "test-api",
	}

	suite.generator = NewAPIGenerator(suite.config)

	// Create temporary directory for test outputs
	tempDir, err := os.MkdirTemp("", "gofastapi-test-*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// Set up test data directory
	suite.testDataDir = "./test-data"
	if err := os.MkdirAll(suite.testDataDir, 0755); err != nil {
		suite.T().Skipf("Cannot create test data directory: %v", err)
	}
}

// TearDownSuite runs once after all tests
func (suite *TestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// SetupTest runs before each test
func (suite *TestSuite) SetupTest() {
	suite.generator.pkgs = make(map[string]*PackageInfo)
}

// TestScanningDirectory tests basic directory scanning functionality
func (suite *TestSuite) TestScanningDirectory() {
	// Create test Go files
	testFiles := map[string]string{
		"user.go": `package models

import "time"

type User struct {
	ID        string    ` + "`json:\"id\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	Email     string    ` + "`json:\"email\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

type UserService struct {
	users map[string]User
}

func (us *UserService) GetUser(id string) (*User, error) {
	if user, exists := us.users[id]; exists {
		return &user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (us *UserService) CreateUser(user *User) (*User, error) {
	user.ID = generateUUID()
	user.CreatedAt = time.Now()
	us.users[user.ID] = *user
	return user, nil
}

func (us *UserService) UpdateUser(id string, user *User) (*User, error) {
	if _, exists := us.users[id]; !exists {
		return nil, fmt.Errorf("user not found")
	}
	user.ID = id
	user.UpdatedAt = time.Now()
	us.users[id] = *user
	return user, nil
}

func (us *UserService) DeleteUser(id string) error {
	if _, exists := us.users[id]; !exists {
		return fmt.Errorf("user not found")
	}
	delete(us.users, id)
	return nil
}

func (us *UserService) ListUsers() ([]User, error) {
	var users []User
	for _, user := range us.users {
		users = append(users, user)
	}
	return users, nil
}

func (us *UserService) SearchUsers(query string) ([]User, error) {
	var results []User
	for _, user := range us.users {
		if strings.Contains(strings.ToLower(user.Name), strings.ToLower(query)) {
			results = append(results, user)
		}
	}
	return results, nil
}

func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}`,
		"product.go": `package models

type Product struct {
	ID          string  ` + "`json:\"id\"`" + `
	Name        string  ` + "`json:\"name\"`" + `
	Description string  ` + "`json:\"description\"`" + `
	Price       float64 ` + "`json:\"price\"`" + `
	Quantity    int     ` + "`json:\"quantity\"`" + `
	Active      bool    ` + "`json:\"active\"`" + `
}

type ProductService struct {
	products map[string]Product
}

func (ps *ProductService) GetProduct(id string) (*Product, error) {
	if product, exists := ps.products[id]; exists {
		return &product, nil
	}
	return nil, fmt.Errorf("product not found")
}

func (ps *ProductService) CreateProduct(product *Product) (*Product, error) {
	product.ID = generateUUID()
	ps.products[product.ID] = *product
	return product, nil
}

func (ps *ProductService) BulkCreateProducts(products []Product) (int, error) {
	count := 0
	for _, product := range products {
		product.ID = generateUUID()
		ps.products[product.ID] = product
		count++
	}
	return count, nil
}

func (ps *ProductService) ActivateProduct(id string) error {
	if product, exists := ps.products[id]; !exists {
		return fmt.Errorf("product not found")
	}
	ps.products[id] = product
	return nil
}

func (ps *ProductService) DeactivateProduct(id string) error {
	if product, exists := ps.products[id]; !exists {
		return fmt.Errorf("product not found")
	}
	ps.products[id] = product
	return nil
}`,
	}

	// Write test files
	for filename, content := range testFiles {
		filePath := filepath.Join(suite.testDataDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(suite.T(), err)
	}

	// Scan directory
	err := suite.generator.ScanDirectory(suite.testDataDir)
	require.NoError(suite.T(), err)

	// Verify packages were found
	assert.Greater(suite.T(), len(suite.generator.pkgs), 0, "Should have found packages")

	// Verify structs and methods
	for _, pkg := range suite.generator.pkgs {
		assert.Greater(suite.T(), len(pkg.Structs), 0, "Should have found structs")
		assert.Greater(suite.T(), len(pkg.Functions), 0, "Should have found functions")

		for _, structInfo := range pkg.Structs {
			assert.Greater(suite.T(), len(structInfo.Methods), 0,
				fmt.Sprintf("Struct %s should have methods", structInfo.Name))
		}
	}
}

// TestSmartMethodMapping tests intelligent method mapping functionality
func (suite *TestSuite) TestSmartMethodMapping() {
	testCases := []struct {
		methodName   string
		structName   string
		expectedOp   string
		expectedPath string
		shouldMatch  bool
	}{
		{"GetUser", "UserService", "get", "/users/{id}", true},
		{"CreateUser", "UserService", "create", "/users", true},
		{"UpdateUser", "UserService", "update", "/users/{id}", true},
		{"DeleteUser", "UserService", "delete", "/users/{id}", true},
		{"ListUsers", "UserService", "list", "/users", true},
		{"SearchUsers", "UserService", "search", "/users/search", true},
		{"GetUserByEmail", "UserService", "get_by", "/users/by/email", true},
		{"BulkCreateProducts", "ProductService", "bulk_create", "/products/bulk", true},
		{"ActivateProduct", "ProductService", "activate", "/products/{id}/activate", true},
		{"DeactivateProduct", "ProductService", "deactivate", "/products/{id}/deactivate", true},
		{"RandomMethod", "TestService", "custom", "/testservice/randommethod", false},
	}

	for _, tc := range testCases {
		mapping, found := suite.generator.SmartMethodMapping(tc.methodName, tc.structName)

		if tc.shouldMatch {
			assert.True(suite.T(), found,
				fmt.Sprintf("Method %s should match pattern", tc.methodName))
			assert.Equal(suite.T(), tc.expectedOp, mapping.Operation,
				fmt.Sprintf("Operation mismatch for %s", tc.methodName))

			// Check that path contains expected elements
			assert.Contains(suite.T(), mapping.Path, strings.ToLower(tc.structName),
				fmt.Sprintf("Path should contain struct name for %s", tc.methodName))
		} else {
			assert.False(suite.T(), found,
				fmt.Sprintf("Method %s should not match any pattern", tc.methodName))
		}
	}
}

// TestRouteGeneration tests API route generation
func (suite *TestSuite) TestRouteGeneration() {
	// Set up test data with methods that should generate routes
	suite.generator.pkgs["test"] = &PackageInfo{
		Name: "test",
		Structs: []StructInfo{
			{
				Name: "UserService",
				Methods: []MethodInfo{
					{Name: "GetUser", Receiver: "*UserService"},
					{Name: "CreateUser", Receiver: "*UserService"},
					{Name: "UpdateUser", Receiver: "*UserService"},
					{Name: "DeleteUser", Receiver: "*UserService"},
					{Name: "ListUsers", Receiver: "*UserService"},
				},
			},
		},
	}

	routes := suite.generator.GenerateAPIRoutes()

	assert.Greater(suite.T(), len(routes), 0, "Should generate routes")

	// Verify route structure
	for _, route := range routes {
		assert.NotEmpty(suite.T(), route.Method, "Route should have HTTP method")
		assert.NotEmpty(suite.T(), route.Path, "Route should have path")
		assert.NotEmpty(suite.T(), route.Function, "Route should have function name")
		assert.NotEmpty(suite.T(), route.Struct, "Route should have struct name")

		// Verify auto-generated metadata
		assert.True(suite.T(), route.Metadata["auto_generated"].(bool),
			"Route should be marked as auto-generated")
	}
}

// TestValidationEngine tests the validation engine
func (suite *TestSuite) TestValidationEngine() {
	config := &ValidationConfig{
		StopOnFirstError: false,
		StrictMode:       true,
		DefaultRules:     []string{"required", "string", "email"},
	}
	engine := NewValidationEngine(config)

	// Test required validator
	result := engine.ValidateField("email", "test@example.com", []string{"required"})
	assert.True(suite.T(), result.Valid, "Valid email should pass validation")

	result = engine.ValidateField("email", "", []string{"required"})
	assert.False(suite.T(), result.Valid, "Empty email should fail required validation")
	assert.Len(suite.T(), result.Errors, 1, "Should have one validation error")

	// Test email validator
	result = engine.ValidateField("email", "invalid-email", []string{"email"})
	assert.False(suite.T(), result.Valid, "Invalid email should fail validation")

	// Test string validator with length constraints
	result = engine.ValidateField("name", "short", []string{"string"})
	assert.True(suite.T(), result.Valid, "Short string should pass")

	// Test numeric validator
	result = engine.ValidateField("age", 25, []string{"numeric"})
	assert.True(suite.T(), result.Valid, "Valid number should pass")

	result = engine.ValidateField("age", "not-a-number", []string{"numeric"})
	assert.False(suite.T(), result.Valid, "Invalid number should fail")
}

// TestPluginSystem tests the plugin system
func (suite *TestSuite) TestPluginSystem() {
	config := &PluginManagerConfig{
		PluginDir:    "./test-plugins",
		AutoLoad:     false,
		SecurityMode: true,
		MaxPlugins:   10,
		SandboxMode:  true,
	}
	manager := NewPluginManager(config)

	// Test built-in plugins
	loggingPlugin := NewLoggingPlugin()
	assert.Equal(suite.T(), "logging", loggingPlugin.GetName())

	metricsPlugin := NewMetricsPlugin()
	assert.Equal(suite.T(), "metrics", metricsPlugin.GetName())

	// Register plugins
	manager.RegisterPlugin(loggingPlugin)
	manager.RegisterPlugin(metricsPlugin)

	// Verify plugin registration
	plugin, exists := manager.GetPlugin("logging")
	assert.True(suite.T(), exists, "Logging plugin should be registered")
	assert.Equal(suite.T(), loggingPlugin, plugin)

	plugin, exists = manager.GetPlugin("metrics")
	assert.True(suite.T(), exists, "Metrics plugin should be registered")
	assert.Equal(suite.T(), metricsPlugin, plugin)

	// Test plugin configuration
	err := manager.ConfigurePlugin("logging", map[string]interface{}{
		"enabled": true,
		"level":   "debug",
	})
	assert.NoError(suite.T(), err, "Plugin configuration should succeed")

	// Test plugin execution
	ctx := &PluginContext{
		EventType: EventAfterScan,
		Config:    map[string]interface{}{},
		Data:      make(map[string]interface{}),
		Metadata:  map[string]interface{}{"package_count": 5},
	}

	err = manager.ExecutePlugins(EventAfterScan, ctx)
	assert.NoError(suite.T(), err, "Plugin execution should succeed")
}

// TestFrameworkGenerators tests framework-specific code generation
func (suite *TestSuite) TestFrameworkGenerators() {
	registry := GetFrameworkRegistry()

	// Test all supported frameworks
	frameworks := registry.ListFrameworks()
	assert.Contains(suite.T(), frameworks, FrameworkGin)
	assert.Contains(suite.T(), frameworks, FrameworkEcho)
	assert.Contains(suite.T(), frameworks, FrameworkChi)
	assert.Contains(suite.T(), frameworks, FrameworkFiber)

	testRoutes := []APIRoute{
		{
			Method:    "GET",
			Path:      "/users/{id}",
			Function:  "GetUser",
			Struct:    "UserService",
			Package:   "models",
			Parameter: []Parameter{{Name: "id", Type: "string"}},
			Response:  []Parameter{{Type: "User"}},
		},
	}

	// Test each framework generator
	for _, frameworkType := range []FrameworkType{FrameworkGin, FrameworkEcho, FrameworkChi, FrameworkFiber} {
		generator, err := registry.GetGenerator(frameworkType)
		require.NoError(suite.T(), err)

		assert.Equal(suite.T(), frameworkType, generator.GetType())
		assert.NotEmpty(suite.T(), generator.GetName())

		// Test code generation
		mainContent, err := generator.GenerateMainFile(testRoutes, generator.GetDefaultConfig())
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), mainContent)

		handlersContent, err := generator.GenerateHandlers(testRoutes, generator.GetDefaultConfig())
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), handlersContent)

		routesContent, err := generator.GenerateRoutes(testRoutes, generator.GetDefaultConfig())
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), routesContent)
	}
}

// TestPerformance tests performance characteristics
func (suite *TestSuite) TestPerformance() {
	// Create a large test file with many structs and methods
	var largeFile strings.Builder
	largeFile.WriteString("package performance\n\nimport \"fmt\"\n\n")

	// Generate 100 structs with methods
	for i := 0; i < 100; i++ {
		structName := fmt.Sprintf("TestStruct%d", i)
		largeFile.WriteString(fmt.Sprintf("type %s struct {\n", structName))
		largeFile.WriteString("    ID string `json:\"id\"`\n")
		largeFile.WriteString("    Name string `json:\"name\"`\n")
		largeFile.WriteString("}\n\n")

		// Generate 10 methods per struct
		for j := 0; j < 10; j++ {
		methodName := fmt.Sprintf("Method%d", j)
			largeFile.WriteString(fmt.Sprintf("func (ts *%s) %s() error {\n", structName, methodName))
			largeFile.WriteString("    return fmt.Errorf(\"not implemented\")\n")
			largeFile.WriteString("}\n\n")
		}
	}

	// Write large file
	largeFilePath := filepath.Join(suite.testDataDir, "large.go")
	err := os.WriteFile(largeFilePath, []byte(largeFile.String()), 0644)
	require.NoError(suite.T(), err)

	// Measure scanning performance
	start := time.Now()
	err = suite.generator.ScanDirectory(suite.testDataDir)
	require.NoError(suite.T(), err)
	duration := time.Since(start)

	suite.T().Logf("Scanning large file took: %v", duration)
	assert.Less(suite.T(), duration, 5*time.Second, "Scanning should complete in reasonable time")

	// Measure route generation performance
	start = time.Now()
	routes := suite.generator.GenerateAPIRoutes()
	duration = time.Since(start)

	suite.T().Logf("Route generation took: %v for %d routes", duration, len(routes))
	assert.Less(suite.T(), duration, 1*time.Second, "Route generation should be fast")
	assert.Greater(suite.T(), len(routes), 1000, "Should generate many routes from large file")
}

// TestErrorHandling tests error handling scenarios
func (suite *TestSuite) TestErrorHandling() {
	// Test invalid Go file
	invalidFile := filepath.Join(suite.testDataDir, "invalid.go")
	err := os.WriteFile(invalidFile, []byte("package invalid\n\nfunc invalid() {"), 0644)
	require.NoError(suite.T(), err)

	// Should handle invalid file gracefully
	err = suite.generator.ScanDirectory(suite.testDataDir)
	assert.NoError(suite.T(), err, "Should handle invalid Go files gracefully")

	// Test validation engine with invalid config
	invalidConfig := &ValidationConfig{
		DefaultRules: []string{"nonexistent_validator"},
	}
	engine := NewValidationEngine(invalidConfig)

	result := engine.ValidateField("test", "value", []string{"nonexistent_validator"})
	// Should not panic, just return result as-is
	assert.NotNil(suite.T(), result)
}

// TestConcurrentAccess tests thread safety
func (suite *TestSuite) TestConcurrentAccess() {
	// Create multiple goroutines accessing the generator
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- true }()

			// Each goroutine performs operations
			routes := suite.generator.GenerateAPIRoutes()
			assert.NotNil(suite.T(), routes)

			// Test validation engine
			engine := GetValidationEngine()
			result := engine.ValidateField("test", "value", []string{"required"})
			assert.NotNil(suite.T(), result)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

// TestMemoryUsage tests memory usage and leaks
func (suite *TestSuite) TestMemoryUsage() {
	// Get initial memory usage
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Perform operations that might allocate memory
	for i := 0; i < 100; i++ {
		config := &ValidationConfig{StrictMode: true}
		engine := NewValidationEngine(config)

		result := engine.ValidateField("test", "value", []string{"required", "string"})
		_ = result
	}

	// Force garbage collection
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Memory growth should be reasonable (less than 10MB)
	memoryGrowth := m2.Alloc - m1.Alloc
	suite.T().Logf("Memory growth: %d bytes", memoryGrowth)
	assert.Less(suite.T(), memoryGrowth, 10*1024*1024, "Memory usage should be reasonable")
}

// TestIntegration tests end-to-end integration
func (suite *TestSuite) TestIntegration() {
	// Create a comprehensive test scenario
	testScenario := map[string]string{
		"models/user.go": `package models

import "time"

// @api.route("/users")
type User struct {
	ID        string    ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	Name      string    ` + "`json:\"name\" gorm:\"not null\"`" + `
	Email     string    ` + "`json:\"email\" gorm:\"uniqueIndex\"`" + `
	Password  string    ` + "`json:\"-\" gorm:\"not null\"`" + `
	Role      string    ` + "`json:\"role\" gorm:\"default:'user'\"`" + `
	Active    bool      ` + "`json:\"active\" gorm:\"default:true\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

type UserService struct {
	db *sql.DB
}

// @api.endpoint GET /users/{id} auth=required
func (us *UserService) GetUser(id string) (*User, error) {
	var user User
	err := us.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// @api.endpoint POST /users auth=required
func (us *UserService) CreateUser(user *User) (*User, error) {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()

	result, err := us.db.Exec(
		"INSERT INTO users (id, name, email, password, role, active, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user.ID, user.Name, user.Email, user.Password, user.Role, user.Active, user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// @api.endpoint GET /users auth=required
func (us *UserService) ListUsers() ([]User, error) {
	rows, err := us.db.Query("SELECT * FROM users WHERE active = ? ORDER BY created_at DESC", true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}`,
	}

	// Write test scenario files
	for filePath, content := range testScenario {
		fullPath := filepath.Join(suite.testDataDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			suite.T().Skipf("Cannot create directory: %v", err)
			continue
		}

		err := os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(suite.T(), err)
	}

	// Scan and generate
	err := suite.generator.ScanDirectory(suite.testDataDir)
	require.NoError(suite.T(), err)

	routes := suite.generator.GenerateAPIRoutes()
	assert.Greater(suite.T(), len(routes), 0, "Should generate routes from integration test")

	// Test each framework
	registry := GetFrameworkRegistry()
	for _, frameworkType := range []FrameworkType{FrameworkGin, FrameworkEcho, FrameworkChi, FrameworkFiber} {
		generator, err := registry.GetGenerator(frameworkType)
		require.NoError(suite.T(), err)

		config := generator.GetDefaultConfig()
		config.Type = frameworkType
		config.Auth = &AuthConfig{Required: true, Type: "jwt"}
		config.Validation = &ValidationConfig{StrictMode: true}

		// Generate full API
		err = registry.GenerateForFramework(frameworkType, routes, suite.generator.pkgs, config)
		assert.NoError(suite.T(), err,
			fmt.Sprintf("Should generate %s API successfully", frameworkType))

		// Verify output directory exists
		outputDir := fmt.Sprintf("./generated-%s-api", frameworkType)
		assert.DirExists(suite.T(), outputDir,
			fmt.Sprintf("%s output directory should exist", frameworkType))

		// Verify key files exist
		expectedFiles := []string{"main.go", "go.mod", ".env.example"}
		for _, file := range expectedFiles {
			filePath := filepath.Join(outputDir, file)
			assert.FileExists(suite.T(), filePath,
				fmt.Sprintf("%s should exist in %s output", file, frameworkType))
		}
	}
}

// BenchmarkRouteGeneration benchmarks route generation performance
func BenchmarkRouteGeneration(b *testing.B) {
	config := &GeneratorConfig{
		SmartMapping: true,
		AutoCRUD:     true,
	}
	generator := NewAPIGenerator(config)

	// Create test package with many structs and methods
	pkg := &PackageInfo{
		Name: "benchmark",
		Structs: make([]StructInfo, 100),
	}

	for i := 0; i < 100; i++ {
		structName := fmt.Sprintf("BenchmarkStruct%d", i)
		methods := make([]MethodInfo, 10)

		for j := 0; j < 10; j++ {
			methodName := fmt.Sprintf("Method%d", j)
			methods[j] = MethodInfo{
				Name:     methodName,
				Receiver: fmt.Sprintf("*%s", structName),
			}
		}

		pkg.Structs[i] = StructInfo{
			Name:    structName,
			Methods: methods,
		}
	}

	generator.pkgs["benchmark"] = pkg

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		routes := generator.GenerateAPIRoutes()
		_ = routes
	}
}

// TestMain is the test entry point
func TestMain(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// Helper function to create directory
func createDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// Helper function to write file
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// Helper function to assert directory exists
func assertDirExists(t *testing.T, path string, msgAndArgs ...interface{}) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("Directory does not exist: %s", path)
			return
		}
		t.Errorf("Error checking directory: %v", err)
		return
	}
	if !info.IsDir() {
		t.Errorf("Path is not a directory: %s", path)
		return
	}
}

// Helper function to assert file exists
func assertFileExists(t *testing.T, path string, msgAndArgs ...interface{}) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("File does not exist: %s", path)
			return
		}
		t.Errorf("Error checking file: %v", err)
		return
	}
	if info.IsDir() {
		t.Errorf("Path is a directory, not a file: %s", path)
		return
	}
}

// Mock implementations for testing
type MockValidator struct {
	name     string
	fail     bool
	errorMsg string
}

func (m *MockValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}
	if m.fail {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "MOCK_ERROR", Message: m.errorMsg},
		}
	}
	return result
}

func (m *MockValidator) GetName() string { return m.name }
func (m *MockValidator) GetType() string { return "mock" }

type MockPlugin struct {
	name      string
	initialized bool
	executed   bool
}

func (m *MockPlugin) GetName() string { return m.name }
func (m *MockPlugin) GetVersion() string { return "1.0.0" }
func (m *MockPlugin) GetDescription() string { return "Mock plugin for testing" }
func (m *MockPlugin) GetAuthor() string { return "Test" }

func (m *MockPlugin) Initialize(config map[string]interface{}) error {
	m.initialized = true
	return nil
}

func (m *MockPlugin) Execute(ctx *PluginContext) error {
	m.executed = true
	return nil
}

func (m *MockPlugin) Cleanup() error {
	return nil
}

func (m *MockPlugin) GetSupportedFrameworks() []string {
	return []string{"gin", "echo", "chi", "fiber"}
}

func (m *MockPlugin) GetSupportedEvents() []PluginEventType {
	return []PluginEventType{EventBeforeScan, EventAfterScan}
}

func (m *MockPlugin) GetDependencies() []PluginDependency {
	return nil
}

func (m *MockPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"enabled": map[string]interface{}{
				"type":    "boolean",
				"default": true,
			},
		},
	}
}

func (m *MockPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

func NewMockPlugin(name string) Plugin {
	return &MockPlugin{name: name}
}

// Test utilities
func generateTestUUID() string {
	return fmt.Sprintf("test-uuid-%d", time.Now().UnixNano())
}

func createTestHTTPRequest(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func parseJSONResponse(resp *httptest.ResponseRecorder, target interface{}) error {
	return json.NewDecoder(resp.Body).Decode(target)
}

func assertJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedCode int, target interface{}) {
	assert.Equal(t, expectedCode, resp.Code, "Response status code should match")

	var response map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err, "Response should be valid JSON")

	if target != nil {
		expected, err := json.Marshal(target)
		assert.NoError(t, err, "Expected target should be valid JSON")

		actual, err := json.Marshal(response)
		assert.NoError(t, err, "Actual response should be valid JSON")

		assert.JSONEq(t, string(expected), string(actual), "JSON response should match expected")
	}
}

// Coverage test helper functions
func measureCoverage(funcName string, fn func() error) (time.Duration, error) {
	start := time.Now()
	defer func() {
		fmt.Printf("Function %s took %v\n", funcName, time.Since(start))
	}()

	err := fn()
	return time.Since(start), err
}

func validateCodeStructure(content string) error {
	// Basic Go syntax validation
	if !strings.Contains(content, "package ") {
		return fmt.Errorf("missing package declaration")
	}

	if !strings.Contains(content, "func ") {
		return fmt.Errorf("no function declarations found")
	}

	if strings.Contains(content, "panic(") && !strings.Contains(content, "recover()") {
		return fmt.Errorf("found panic without recover")
	}

	return nil
}

func extractImports(content string) []string {
	importRegex := regexp.MustCompile(`import\s*\((.*?)\)`)
	matches := importRegex.FindStringSubmatch(content)

	if len(matches) < 2 {
		return nil
	}

	imports := strings.Split(matches[1], "\n")
	var result []string

	for _, imp := range imports {
		imp = strings.TrimSpace(imp)
		if imp != "" && !strings.HasPrefix(imp, "//") {
			result = append(result, imp)
		}
	}

	return result
}

func extractFunctionNames(content string) []string {
	funcRegex := regexp.MustCompile(`func\s+\w+\s*\(`)
	matches := funcRegex.FindAllString(content, -1)

	var result []string
	for _, match := range matches {
		name := strings.TrimPrefix(match, "func ")
		name = strings.TrimSuffix(name, "(")
		name = strings.TrimSpace(name)
		if name != "" {
			result = append(result, name)
		}
	}

	return result
}

// Test data generators
func generateTestStructs(count int) []StructInfo {
	structs := make([]StructInfo, count)

	for i := 0; i < count; i++ {
		structs[i] = StructInfo{
			Name: fmt.Sprintf("TestStruct%d", i),
			Fields: []FieldInfo{
				{Name: "ID", Type: "string"},
				{Name: "Name", Type: "string"},
				{Name: "Value", Type: "int"},
			},
			Methods: generateTestMethods(5),
		}
	}

	return structs
}

func generateTestMethods(count int) []MethodInfo {
	methods := make([]MethodInfo, count)

	for i := 0; i < count; i++ {
		methods[i] = MethodInfo{
			Name: fmt.Sprintf("TestMethod%d", i),
			Parameters: []Parameter{
				{Name: "input", Type: "string"},
			},
			Returns: []Parameter{
				{Name: "output", Type: "string"},
			},
		}
	}

	return methods
}

func generateTestRoutes(count int) []APIRoute {
	routes := make([]APIRoute, count)

	for i := 0; i < count; i++ {
		routes[i] = APIRoute{
			Method:    "GET",
			Path:      fmt.Sprintf("/test/%d/{id}", i),
			Function:  fmt.Sprintf("GetTest%d", i),
			Struct:    fmt.Sprintf("TestStruct%d", i),
			Package:   "test",
			Parameter: []Parameter{{Name: "id", Type: "string"}},
			Response:  []Parameter{{Type: fmt.Sprintf("TestStruct%d", i)}},
		}
	}

	return routes
}

// Configuration test helpers
func createTestConfig() *GeneratorConfig {
	return &GeneratorConfig{
		IncludePatterns: []string{"*.go"},
		ExcludePatterns: []string{"*_test.go"},
		ScanAnnotations: true,
		AutoCRUD:        true,
		SmartMapping:    true,
		OutputDir:       "./test-output",
		PackageName:     "test-api",
	}
}

func createTestValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		StopOnFirstError: false,
		StrictMode:       true,
		DefaultRules:     []string{"required", "string"},
	}
}

func createTestFrameworkConfig(frameworkType FrameworkType) *FrameworkConfig {
	return &FrameworkConfig{
		Type:     frameworkType,
		Version:  "1.0.0",
		Features: []string{"middleware", "validation", "cors"},
		CORS: &CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		},
		Validation: &ValidationConfig{
			StrictMode: true,
		},
		Docs: &DocumentationConfig{
			Enabled: true,
			Path:    "/docs",
			Format:  "swagger",
		},
		Testing: &TestingConfig{
			Enabled:   true,
			Framework: "testify",
			Coverage:  true,
		},
	}
}

// Performance monitoring
type PerformanceMetrics struct {
	ScanDuration     time.Duration
	RouteGenDuration time.Duration
	ValidationDuration time.Duration
	MemoryUsage      uint64
	RouteCount       int
	PackageCount     int
	StructCount      int
	MethodCount      int
}

func (suite *TestSuite) collectPerformanceMetrics() *PerformanceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	routes := suite.generator.GenerateAPIRoutes()

	var structCount, methodCount int
	for _, pkg := range suite.generator.pkgs {
		structCount += len(pkg.Structs)
		for _, st := range pkg.Structs {
			methodCount += len(st.Methods)
		}
	}

	return &PerformanceMetrics{
		RouteCount:   len(routes),
		PackageCount:  len(suite.generator.pkgs),
		StructCount:   structCount,
		MethodCount:   methodCount,
		MemoryUsage:   m.Alloc,
	}
}

func (suite *TestSuite) assertPerformanceMetrics(metrics *PerformanceMetrics) {
	suite.T().Logf("Performance Metrics:")
	suite.T().Logf("  Routes: %d", metrics.RouteCount)
	suite.T().Logf("  Packages: %d", metrics.PackageCount)
	suite.T().Logf("  Structs: %d", metrics.StructCount)
	suite.T().Logf("  Methods: %d", metrics.MethodCount)
	suite.T().Logf("  Memory: %d MB", metrics.MemoryUsage/1024/1024)

	// Performance assertions
	assert.Less(suite.T(), metrics.MemoryUsage, 100*1024*1024, "Memory usage should be under 100MB")
	assert.Greater(suite.T(), metrics.RouteCount, 0, "Should generate routes")
}

// End-to-end test scenarios
func (suite *TestSuite) TestE2EScenario_RealWorldAPI() {
	// Simulate a real-world API scenario with complex models and relationships
	testFiles := map[string]string{
		"models/user.go": `
package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID        string         ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	Username  string         ` + "`json:\"username\" gorm:\"uniqueIndex;not null\"`" + `
	Email     string         ` + "`json:\"email\" gorm:\"uniqueIndex;not null\"`" + `
	Password  string         ` + "`json:\"-\" gorm:\"not null\"`" + `
	FirstName string         ` + "`json:\"first_name\"`" + `
	LastName  string         ` + "`json:\"last_name\"`" + `
	Avatar    string         ` + "`json:\"avatar\"`" + `
	Bio       string         ` + "`json:\"bio\"`" + `
	Active    bool           ` + "`json:\"active\" gorm:\"default:true\"`" + `
	Role      string         ` + "`json:\"role\" gorm:\"default:'user'\"`" + `
	Settings  UserSettings   ` + "`json:\"settings\" gorm:\"embedded\"`" + `
	Posts     []Post          ` + "`json:\"posts\" gorm:\"foreignKey:AuthorID\"`" + `
	Profile   UserProfile    ` + "`json:\"profile\" gorm:\"foreignKey:UserID\"`" + `
	CreatedAt time.Time      ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`json:\"updated_at\"`" + `
	DeletedAt gorm.DeletedAt ` + "`json:\"deleted_at,omitempty\"`" + `
}

type UserSettings struct {
	Theme           string ` + "`json:\"theme\"`" + `
	Language        string ` + "`json:\"language\"`" + `
	Notifications   bool   ` + "`json:\"notifications\"`" + `
	Privacy         bool   ` + "`json:\"privacy\"`" + `
	EmailVerified   bool   ` + "`json:\"email_verified\"`" + `
}

type UserProfile struct {
	UserID     string    ` + "`json:\"user_id\"`" + `
	Bio        string    ` + "`json:\"bio\"`" + `
	Location   string    ` + "`json:\"location\"`" + `
	Website    string    ` + "`json:\"website\"`" + `
	Social     SocialLinks ` + "`json:\"social\" gorm:\"embedded\"`" + `
	Skills     []Skill   ` + "`json:\"skills\"`" + `
	Experience []Experience ` + "`json:\"experience\"`" + `
	Education []Education ` + "`json:\"education\"`" + `
}

type SocialLinks struct {
	Twitter   string ` + "`json:\"twitter\"`" + `
	LinkedIn  string ` + "`json:\"linkedin\"`" + `
	GitHub    string ` + "`json:\"github\"`" + `
	Instagram string ` + "`json:\"instagram\"`" + `
}

type Skill struct {
	Name      string ` + "`json:\"name\"`" + `
	Level     string ` + "`json:\"level\"`" + `
	Category  string ` + "`json:\"category\"`" + `
}

type UserService struct {
	db *gorm.DB
}

// API methods with annotations
func (us *UserService) GetUser(id string) (*User, error) {
	var user User
	result := us.db.Preload("Profile").Preload("Posts").First(&user, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (us *UserService) CreateUser(user *User) (*User, error) {
	result := us.db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (us *UserService) UpdateUser(id string, updates *User) (*User, error) {
	var user User
	if err := us.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	result := us.db.Model(&user).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (us *UserService) DeleteUser(id string) error {
	return us.db.Delete(&User{}, "id = ?", id).Error
}

func (us *UserService) ListUsers(page, limit int) ([]User, int64, error) {
	var users []User
	var total int64

	offset := (page - 1) * limit

	us.db.Model(&User{}).Count(&total)
	result := us.db.Preload("Profile").Offset(offset).Limit(limit).Find(&users)

	return users, total, result.Error
}

func (us *UserService) SearchUsers(query string, filters map[string]interface{}) ([]User, error) {
	var users []User
	db := us.db.Model(&User{}).Preload("Profile")

	if query != "" {
		db = db.Where("username LIKE ? OR first_name LIKE ? OR last_name LIKE ? OR email LIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	for key, value := range filters {
		db = db.Where(key, value)
	}

	result := db.Find(&users)
	return users, result.Error
}

func (us *UserService) GetUserByEmail(email string) (*User, error) {
	var user User
	result := us.db.First(&user, "email = ?", email)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (us *UserService) ChangePassword(userID string, oldPassword, newPassword string) error {
	var user User
	if err := us.db.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	// Verify old password
	if !bcrypt.CheckPasswordHash(oldPassword, user.Password) {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return us.db.Model(&user).Update("password", hashedPassword).Error
}

func (us *UserService) ActivateUser(id string) error {
	return us.db.Model(&User{}).Where("id = ?", id).Update("active", true).Error
}

func (us *UserService) DeactivateUser(id string) error {
	return us.db.Model(&User{}).Where("id = ?", id).Update("active", false).Error
}

func (us *UserService) BulkUpdateUserStatus(userIDs []string, active bool) error {
	return us.db.Model(&User{}).Where("id IN ?", userIDs).Update("active", active).Error
}`,
		"models/post.go": `
package models

import (
	"time"
	"gorm.io/gorm"
)

type Post struct {
	ID          string      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	Title       string      ` + "`json:\"title\" gorm:\"not null\"`" + `
	Content     string      ` + "`json:\"content\" gorm:\"type:text\"`" + `
	Excerpt     string      ` + "`json:\"excerpt\"`" + `
	Slug        string      ` + "`json:\"slug\" gorm:\"uniqueIndex\"`" + `
	Status      string      ` + "`json:\"status\" gorm:\"default:'draft'\"`" + `
	Type        string      ` + "`json:\"type\" gorm:\"default:'post'\"`" + `
	AuthorID    string      ` + "`json:\"author_id\"`" + `
	CategoryID  string      ` + "`json:\"category_id\"`" + `
	Tags        []Tag       ` + "`json:\"tags\" gorm:\"many2many:post_tags\"`" + `
	Meta        PostMeta    ` + "`json:\"meta\" gorm:\"embedded\"`" + `
	SEO         SEO         ` + "`json:\"seo\" gorm:\"embedded\"`" + `
	Featured    bool        ` + "`json:\"featured\" gorm:\"default:false\"`" + `
	PublishedAt *time.Time ` + "`json:\"published_at\"`" + `
	CreatedAt   time.Time   ` + "`json:\"created_at\"`" + `
	UpdatedAt   time.Time   ` + "`json:\"updated_at\"`" + `
	DeletedAt   gorm.DeletedAt ` + "`json:\"deleted_at,omitempty\"`" + `
}

type PostMeta struct {
	Title        string ` + "`json:\"title\"`" + `
	Description  string ` + "`json:\"description\"`" + `
	Keywords     string ` + "`json:\"keywords\"`" + `
}

type SEO struct {
	Title       string ` + "`json:\"title\"`" + `
	Description string ` + "`json:\"description\"`" + `
	Canonical   string ` + "`json:\"canonical\"`" + `
	NoIndex     bool   ` + "`json:\"no_index\"`" + `
	NoFollow    bool   ` + "`json:\"no_follow\"`" + `
}

type PostService struct {
	db *gorm.DB
}

func (ps *PostService) GetPost(id string) (*Post, error) {
	var post Post
	result := ps.db.Preload("Tags").Preload("Category").First(&post, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &post, nil
}

func (ps *PostService) CreatePost(post *Post) (*Post, error) {
	result := ps.db.Create(post)
	if result.Error != nil {
		return nil, result.Error
	}
	return post, nil
}

func (ps *PostService) UpdatePost(id string, updates *Post) (*Post, error) {
	var post Post
	if err := ps.db.First(&post, "id = ?", id).Error; err != nil {
		return nil, err
	}

	result := ps.db.Model(&post).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}

	return &post, nil
}

func (ps *PostService) DeletePost(id string) error {
	return ps.db.Delete(&Post{}, "id = ?", id).Error
}

func (ps *PostService) ListPosts(page, limit int, filters map[string]interface{}) ([]Post, int64, error) {
	var posts []Post
	var total int64

	offset := (page - 1) * limit

	db := ps.db.Model(&Post{}).Preload("Tags").Preload("Category").Preload("Author")

	for key, value := range filters {
		db = db.Where(key, value)
	}

	db.Count(&total)
	result := db.Offset(offset).Limit(limit).Find(&posts)

	return posts, total, result.Error
}

func (ps *PostService) PublishPost(id string) error {
	now := time.Now()
	return ps.db.Model(&Post{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "published",
		"published_at": &now,
	}).Error
}

func (ps *PostService) UnpublishPost(id string) error {
	return ps.db.Model(&Post{}).Where("id = ?", id).Update("status", "draft").Error
}

func (ps *PostService) GetPostsByAuthor(authorID string, page, limit int) ([]Post, int64, error) {
	var posts []Post
	var total int64

	offset := (page - 1) * limit

	ps.db.Model(&Post{}).Where("author_id = ?", authorID).Count(&total)
	result := ps.db.Preload("Tags").Preload("Category").Where("author_id = ?", authorID).
		Offset(offset).Limit(limit).Find(&posts)

	return posts, total, result.Error
}

func (ps *PostService) GetPostsByCategory(categoryID string, page, limit int) ([]Post, int64, error) {
	var posts []Post
	var total int64

	offset := (page - 1) * limit

	ps.db.Model(&Post{}).Where("category_id = ?", categoryID).Count(&total)
	result := ps.db.Preload("Tags").Preload("Author").Where("category_id = ?", categoryID).
		Offset(offset).Limit(limit).Find(&posts)

	return posts, total, result.Error
}

func (ps *PostService) SearchPosts(query string, page, limit int) ([]Post, int64, error) {
	var posts []Post
	var total int64

	offset := (page - 1) * limit

	searchQuery := "%" + query + "%"
	ps.db.Model(&Post{}).Where("title LIKE ? OR content LIKE ? OR excerpt LIKE ?",
		searchQuery, searchQuery, searchQuery).Count(&total)

	result := ps.db.Preload("Tags").Preload("Author").Preload("Category").
		Where("title LIKE ? OR content LIKE ? OR excerpt LIKE ?",
		searchQuery, searchQuery, searchQuery).
		Offset(offset).Limit(limit).Find(&posts)

	return posts, total, result.Error
}

func (ps *PostService) GetFeaturedPosts(limit int) ([]Post, error) {
	var posts []Post
	result := ps.db.Preload("Tags").Preload("Author").Preload("Category").
		Where("featured = ? AND status = ?", true, "published").
		Limit(limit).Order("published_at DESC").Find(&posts)

	return posts, result.Error
}
`,
	}

	// Write test files
	for filePath, content := range testFiles {
		fullPath := filepath.Join(suite.testDataDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			suite.T().Skipf("Cannot create directory: %v", err)
			continue
		}

		err := os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(suite.T(), err)
	}

	// Scan and analyze
	err := suite.generator.ScanDirectory(suite.testDataDir)
	require.NoError(suite.T(), err)

	// Collect performance metrics
	metrics := suite.collectPerformanceMetrics()
	suite.assertPerformanceMetrics(metrics)

	// Generate routes and test each framework
	routes := suite.generator.GenerateAPIRoutes()
	assert.Greater(suite.T(), len(routes), 20, "Should generate substantial number of routes")

	// Test route quality
	smartRoutes := 0
	crudRoutes := 0
	for _, route := range routes {
		if route.Metadata["smart_mapping"] == true {
			smartRoutes++
		}
		if route.Metadata["auto_generated"] == true {
			crudRoutes++
		}
	}

	suite.T().Logf("Generated %d routes (%d smart-mapped, %d auto-generated)",
		len(routes), smartRoutes, crudRoutes)

	// Test framework generation for real-world scenario
	registry := GetFrameworkRegistry()

	for _, frameworkType := range []FrameworkType{FrameworkGin, FrameworkEcho, FrameworkChi, FrameworkFiber} {
		_, err := registry.GetGenerator(frameworkType)
		require.NoError(suite.T(), err)

		config := createTestFrameworkConfig(frameworkType)
		config.Auth = &AuthConfig{Required: true, Type: "jwt"}
		config.Validation = &ValidationConfig{StrictMode: true}
		config.Docs = &DocumentationConfig{
			Enabled: true,
			Path:    "/api/docs",
			Format:  "swagger",
			Title:   "Real-World Test API",
			Version: "1.0.0",
		}

		// Generate complete API
		start := time.Now()
		err = registry.GenerateForFramework(frameworkType, routes, suite.generator.pkgs, config)
		generationTime := time.Since(start)

		assert.NoError(suite.T(), err,
			fmt.Sprintf("Should generate complete %s API successfully", frameworkType))
		assert.Less(suite.T(), generationTime, 5*time.Second,
			fmt.Sprintf("%s generation should complete in reasonable time", frameworkType))

		// Verify output quality
		outputDir := fmt.Sprintf("./generated-%s-api", frameworkType)
		mainFile := filepath.Join(outputDir, "main.go")

		mainContent, err := os.ReadFile(mainFile)
		assert.NoError(suite.T(), err)

		// Validate generated code quality
		assert.NoError(suite.T(), validateCodeStructure(string(mainContent)),
			fmt.Sprintf("Generated %s code should have valid structure", frameworkType))

		// Check for proper imports
		imports := extractImports(string(mainContent))
		assert.NotEmpty(suite.T(), imports, "Should have imports")

		// Check for essential functions
		functions := extractFunctionNames(string(mainContent))
		assert.Contains(suite.T(), functions, "main", "Should have main function")

		suite.T().Logf("%s API generated in %v with %d lines",
			frameworkType, generationTime, len(strings.Split(string(mainContent), "\n")))
	}

	// Test validation system with complex rules
	engine := GetValidationEngine()

	// Test user validation
	userValidationRules := []string{"required", "string", "email", "min_length:3", "max_length:50"}

	validUser := map[string]interface{}{
		"email":    "test@example.com",
		"username": "testuser",
		"password": "SecurePass123!",
	}

	for field, value := range validUser {
		result := engine.ValidateField(field, value, userValidationRules)
		if field == "email" {
			result = engine.ValidateField(field, value, []string{"required", "email"})
		}
		assert.True(suite.T(), result.Valid,
			fmt.Sprintf("Valid user field %s should pass validation", field))
	}

	// Test plugin system with complex scenario
	manager := NewPluginManager(&PluginManagerConfig{
		AutoLoad:     false,
		SecurityMode: true,
		SandboxMode:  true,
	})

	// Register mock plugins
	for i := 0; i < 5; i++ {
		plugin := NewMockPlugin(fmt.Sprintf("test-plugin-%d", i))
		manager.RegisterPlugin(plugin)
		manager.ConfigurePlugin(plugin.GetName(), map[string]interface{}{
			"enabled": true,
			"priority": i,
		})
	}

	// Initialize and test plugins
	err = manager.InitializePlugins()
	assert.NoError(suite.T(), err)

	// Test plugin execution pipeline
	ctx := &PluginContext{
		EventType: EventAfterScan,
		Config:    map[string]interface{}{},
		Data:      make(map[string]interface{}),
		Metadata:  map[string]interface{}{
			"routes_generated": len(routes),
			"scan_duration":   100 * time.Millisecond,
			"memory_usage":    1024 * 1024,
		},
	}

	err = manager.ExecutePlugins(EventAfterScan, ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, len(manager.ListPlugins()), "All plugins should be registered")

	// Verify plugins were executed
	for i := 0; i < 5; i++ {
		pluginName := fmt.Sprintf("test-plugin-%d", i)
		plugin, exists := manager.GetPlugin(pluginName)
		assert.True(suite.T(), exists, "Mock plugin should exist")

		if mockPlugin, ok := plugin.(*MockPlugin); ok {
			assert.True(suite.T(), mockPlugin.initialized, "Plugin should be initialized")
			assert.True(suite.T(), mockPlugin.executed, "Plugin should have been executed")
		}
	}

	suite.T().Logf("E2E test completed successfully with:")
	suite.T().Logf("  - %d packages scanned", len(suite.generator.pkgs))
	suite.T().Logf("  - %d structs analyzed", metrics.StructCount)
	suite.T().Logf("  - %d methods processed", metrics.MethodCount)
	suite.T().Logf("  - %d routes generated", len(routes))
	suite.T().Logf("  - %d frameworks tested", 4)
	suite.T().Logf("  - %d plugins validated", 5)
}

// Additional test utilities
func init() {
	// Set up test environment
	os.Setenv("TEST_ENV", "true")
	os.Setenv("LOG_LEVEL", "debug")
}

// Cleanup function for test environment
func cleanupTest() {
	if testDir := "./test-data"; os.Getenv("TEST_ENV") == "true" {
		os.RemoveAll(testDir)
	}

	for _, framework := range []string{"gin", "echo", "chi", "fiber"} {
		if outputDir := fmt.Sprintf("./generated-%s-api", framework); os.Getenv("TEST_ENV") == "true" {
			os.RemoveAll(outputDir)
		}
	}
}
