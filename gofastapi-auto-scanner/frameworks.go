package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FrameworkType represents supported web frameworks
type FrameworkType string

const (
	FrameworkGin   FrameworkType = "gin"
	FrameworkEcho  FrameworkType = "echo"
	FrameworkChi   FrameworkType = "chi"
	FrameworkFiber FrameworkType = "fiber"
)

// FrameworkConfig contains framework-specific configuration
type FrameworkConfig struct {
	Type        FrameworkType           `json:"type"`
	Version     string                  `json:"version"`
	Features    []string                `json:"features"`
	Middleware  []string                `json:"middleware"`
	Validation  *ValidationConfig       `json:"validation"`
	Auth        *AuthConfig             `json:"auth"`
	CORS        *CORSConfig              `json:"cors"`
	Database    *DatabaseConfig         `json:"database"`
	Docs        *DocumentationConfig    `json:"docs"`
	Testing     *TestingConfig          `json:"testing"`
	Deployment  *DeploymentConfig       `json:"deployment"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled"`
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSL      bool   `json:"ssl"`
}

// DocumentationConfig contains API documentation configuration
type DocumentationConfig struct {
	Enabled   bool   `json:"enabled"`
	Path      string `json:"path"`
	Format    string `json:"format"` // "swagger", "openapi", "redoc"
	Title     string `json:"title"`
	Version   string `json:"version"`
	Host      string `json:"host"`
	BasePath  string `json:"base_path"`
}

// TestingConfig contains testing configuration
type TestingConfig struct {
	Enabled    bool     `json:"enabled"`
	Framework  string   `json:"framework"` // "testify", "ginkgo", "gomega"
	Coverage   bool     `json:"coverage"`
	Benchmark  bool     `json:"benchmark"`
	Integration bool    `json:"integration"`
	E2E        bool     `json:"e2e"`
	Tools      []string `json:"tools"`
}

// DeploymentConfig contains deployment configuration
type DeploymentConfig struct {
	Type      string            `json:"type"` // "docker", "kubernetes", "serverless"
	Platform  string            `json:"platform"`
	Config    map[string]interface{} `json:"config"`
}

// FrameworkGenerator interface for framework-specific code generation
type FrameworkGenerator interface {
	GetName() string
	GetType() FrameworkType
	GetDefaultConfig() *FrameworkConfig
	GenerateMainFile(routes []APIRoute, config *FrameworkConfig) (string, error)
	GenerateMiddleware(config *FrameworkConfig) (string, error)
	GenerateHandlers(routes []APIRoute, config *FrameworkConfig) (string, error)
	GenerateRoutes(routes []APIRoute, config *FrameworkConfig) (string, error)
	GenerateModels(structs []StructInfo, config *FrameworkConfig) (string, error)
	GenerateTests(routes []APIRoute, config *FrameworkConfig) (string, error)
	GenerateDocs(routes []APIRoute, config *FrameworkConfig) (string, error)
	GenerateDockerfile(config *FrameworkConfig) (string, error)
	GenerateK8sManifests(config *FrameworkConfig) (map[string]string, error)
}

// FrameworkRegistry manages framework generators
type FrameworkRegistry struct {
	generators map[FrameworkType]FrameworkGenerator
}

// NewFrameworkRegistry creates a new framework registry
func NewFrameworkRegistry() *FrameworkRegistry {
	registry := &FrameworkRegistry{
		generators: make(map[FrameworkType]FrameworkGenerator),
	}

	// Register built-in framework generators
	registry.RegisterGenerator(NewGinGenerator())
	registry.RegisterGenerator(NewEchoGenerator())
	registry.RegisterGenerator(NewChiGenerator())
	registry.RegisterGenerator(NewFiberGenerator())

	return registry
}

// RegisterGenerator registers a framework generator
func (fr *FrameworkRegistry) RegisterGenerator(generator FrameworkGenerator) {
	fr.generators[generator.GetType()] = generator
}

// GetGenerator returns a generator for the specified framework
func (fr *FrameworkRegistry) GetGenerator(frameworkType FrameworkType) (FrameworkGenerator, error) {
	generator, exists := fr.generators[frameworkType]
	if !exists {
		return nil, fmt.Errorf("unsupported framework: %s", frameworkType)
	}
	return generator, nil
}

// ListFrameworks returns all supported frameworks
func (fr *FrameworkRegistry) ListFrameworks() []FrameworkType {
	frameworks := make([]FrameworkType, 0, len(fr.generators))
	for frameworkType := range fr.generators {
		frameworks = append(frameworks, frameworkType)
	}
	return frameworks
}

// GenerateForFramework generates API code for a specific framework
func (fr *FrameworkRegistry) GenerateForFramework(
	frameworkType FrameworkType,
	routes []APIRoute,
	packages map[string]*PackageInfo,
	config *FrameworkConfig,
) error {
	generator, err := fr.GetGenerator(frameworkType)
	if err != nil {
		return err
	}

	// Use default config if none provided
	if config == nil {
		config = generator.GetDefaultConfig()
		config.Type = frameworkType
	}

	// Generate main file
	mainContent, err := generator.GenerateMainFile(routes, config)
	if err != nil {
		return fmt.Errorf("failed to generate main file: %v", err)
	}

	// Generate middleware
	middlewareContent, err := generator.GenerateMiddleware(config)
	if err != nil {
		return fmt.Errorf("failed to generate middleware: %v", err)
	}

	// Generate handlers
	handlersContent, err := generator.GenerateHandlers(routes, config)
	if err != nil {
		return fmt.Errorf("failed to generate handlers: %v", err)
	}

	// Generate routes
	routesContent, err := generator.GenerateRoutes(routes, config)
	if err != nil {
		return fmt.Errorf("failed to generate routes: %v", err)
	}

	// Generate models
	var structs []StructInfo
	for _, pkg := range packages {
		structs = append(structs, pkg.Structs...)
	}
	modelsContent, err := generator.GenerateModels(structs, config)
	if err != nil {
		return fmt.Errorf("failed to generate models: %v", err)
	}

	// Write generated files
	outputDir := fmt.Sprintf("./generated-%s-api", frameworkType)
	if err := writeGeneratedFiles(outputDir, mainContent, middlewareContent, handlersContent, routesContent, modelsContent, config); err != nil {
		return fmt.Errorf("failed to write generated files: %v", err)
	}

	// Generate tests if enabled
	if config.Testing != nil && config.Testing.Enabled {
		testsContent, err := generator.GenerateTests(routes, config)
		if err != nil {
			return fmt.Errorf("failed to generate tests: %v", err)
		}
		if err := writeTestFiles(outputDir, testsContent, config); err != nil {
			return fmt.Errorf("failed to write test files: %v", err)
		}
	}

	// Generate documentation if enabled
	if config.Docs != nil && config.Docs.Enabled {
		docsContent, err := generator.GenerateDocs(routes, config)
		if err != nil {
			return fmt.Errorf("failed to generate docs: %v", err)
		}
		if err := writeDocFiles(outputDir, docsContent, config); err != nil {
			return fmt.Errorf("failed to write doc files: %v", err)
		}
	}

	// Generate deployment files if enabled
	if config.Deployment != nil {
		if config.Deployment.Type == "docker" {
			dockerfileContent, err := generator.GenerateDockerfile(config)
			if err != nil {
				return fmt.Errorf("failed to generate Dockerfile: %v", err)
			}
			if err := writeDockerfile(outputDir, dockerfileContent); err != nil {
				return fmt.Errorf("failed to write Dockerfile: %v", err)
			}
		} else if config.Deployment.Type == "kubernetes" {
			manifests, err := generator.GenerateK8sManifests(config)
			if err != nil {
				return fmt.Errorf("failed to generate K8s manifests: %v", err)
			}
			if err := writeK8sManifests(outputDir, manifests); err != nil {
				return fmt.Errorf("failed to write K8s manifests: %v", err)
			}
		}
	}

	return nil
}

// Helper function to write generated files
func writeGeneratedFiles(outputDir, mainContent, middlewareContent, handlersContent, routesContent, modelsContent string, config *FrameworkConfig) error {
	// Create output directory
	if err := createDirectory(outputDir); err != nil {
		return err
	}

	// Write main.go
	if err := writeFile(filepath.Join(outputDir, "main.go"), mainContent); err != nil {
		return err
	}

	// Write middleware.go
	if middlewareContent != "" {
		if err := writeFile(filepath.Join(outputDir, "middleware.go"), middlewareContent); err != nil {
			return err
		}
	}

	// Write handlers.go
	if handlersContent != "" {
		if err := writeFile(filepath.Join(outputDir, "handlers.go"), handlersContent); err != nil {
			return err
		}
	}

	// Write routes.go
	if routesContent != "" {
		if err := writeFile(filepath.Join(outputDir, "routes.go"), routesContent); err != nil {
			return err
		}
	}

	// Write models.go
	if modelsContent != "" {
		if err := writeFile(filepath.Join(outputDir, "models.go"), modelsContent); err != nil {
			return err
		}
	}

	// Write go.mod
	goModContent := fmt.Sprintf(`module generated-%s-api

go 1.21

require (
`, string(config.Type))

	// Add framework-specific dependencies
	switch config.Type {
	case FrameworkGin:
		goModContent += `	github.com/gin-gonic/gin v1.10.0
	github.com/golang-jwt/jwt/v4 v4.5.2
`
	case FrameworkEcho:
		goModContent += `	github.com/labstack/echo/v4 v4.11.4
	github.com/golang-jwt/jwt/v4 v4.5.2
`
	case FrameworkChi:
		goModContent += `	github.com/go-chi/chi/v5 v5.0.12
	github.com/golang-jwt/jwt/v4 v4.5.2
`
	case FrameworkFiber:
		goModContent += `	github.com/gofiber/fiber/v2 v2.52.4
	github.com/golang-jwt/jwt/v4 v4.5.2
`
	}

	// Add common dependencies
	goModContent += `	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.26.0
)`

	if err := writeFile(filepath.Join(outputDir, "go.mod"), goModContent); err != nil {
		return err
	}

	// Write .env.example
	envExample := `PORT=8080
JWT_SECRET=your-secret-key-here
DATABASE_URL=postgresql://user:password@localhost:5432/dbname?sslmode=disable
LOG_LEVEL=info
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
`
	if err := writeFile(filepath.Join(outputDir, ".env.example"), envExample); err != nil {
		return err
	}

	return nil
}

// Framework-specific generators

// Gin Framework Generator
type GinGenerator struct{}

func NewGinGenerator() FrameworkGenerator {
	return &GinGenerator{}
}

func (g *GinGenerator) GetName() string { return "Gin" }
func (g *GinGenerator) GetType() FrameworkType { return FrameworkGin }
func (g *GinGenerator) GetDefaultConfig() *FrameworkConfig {
	return &FrameworkConfig{
		Type:     FrameworkGin,
		Version:  "v1.10.0",
		Features: []string{"middleware", "validation", "cors", "jwt"},
		Middleware: []string{"logger", "recovery", "cors", "auth"},
		Validation: &ValidationConfig{
			StrictMode: true,
		},
		CORS: &CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		},
		Docs: &DocumentationConfig{
			Enabled: true,
			Path:    "/swagger",
			Format:  "swagger",
		},
		Testing: &TestingConfig{
			Enabled:   true,
			Framework: "testify",
			Coverage:  true,
		},
	}
}

func (g *GinGenerator) GenerateMainFile(routes []APIRoute, config *FrameworkConfig) (string, error) {
	return fmt.Sprintf(`package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title Generated %s API
// @version 1.0
// @description Auto-generated API using GoFastAPI
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Gin
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create server
	server := NewServer()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting %s server on port %%s", port)
	if err := server.router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %%v", err)
	}
}`, strings.Title(string(config.Type))), nil
}

func (g *GinGenerator) GenerateMiddleware(config *FrameworkConfig) (string, error) {
	return fmt.Sprintf(`package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// SetupMiddleware configures all middleware for the Gin server
func (s *Server) setupMiddleware() {
	// CORS middleware
	if %t {
		s.router.Use(cors.New(cors.Config{
			AllowOrigins:     %s,
			AllowMethods:     %s,
			AllowHeaders:     %s,
			ExposeHeaders:    %s,
			AllowCredentials: %t,
			MaxAge:           %d * time.Hour,
		}))
	}

	// Request ID middleware
	s.router.Use(requestIDMiddleware())

	// Security headers middleware
	s.router.Use(securityHeadersMiddleware())
}

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
		}

		c.Next()
	}
}

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateUUID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})
}

// securityHeadersMiddleware adds security headers
func securityHeadersMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	})
}
`,
		config.CORS.Enabled,
		formatStringSlice(config.CORS.AllowOrigins),
		formatStringSlice(config.CORS.AllowMethods),
		formatStringSlice(config.CORS.AllowHeaders),
		formatStringSlice(config.CORS.ExposeHeaders),
		config.CORS.AllowCredentials,
		config.CORS.MaxAge/3600,
	), nil
}

func (g *GinGenerator) GenerateHandlers(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var handlers strings.Builder

	handlers.WriteString("package main\n\n")
	handlers.WriteString("import (\n")
	handlers.WriteString(`	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
` + "}\n\n")

	for _, route := range routes {
		handlerName := toCamelCase(route.Function) + "Handler"
		handlers.WriteString(fmt.Sprintf(`// %s handles %s %s
func (s *Server) %s(c *gin.Context) {
	// TODO: Implement business logic for %s

	// Extract path parameters
`, handlerName, strings.ToUpper(route.Method), route.Path, handlerName, route.Function))

		// Generate parameter extraction
		for _, param := range route.Parameter {
			if param.Name == "id" {
				handlers.WriteString(fmt.Sprintf("	id := c.Param(\"id\")\n"))
			} else if param.Name == "q" {
				handlers.WriteString(fmt.Sprintf("	q := c.Query(\"q\")\n"))
			} else if param.Name == "limit" {
				handlers.WriteString(fmt.Sprintf(`	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
`))
			} else if param.Name == "offset" {
				handlers.WriteString(fmt.Sprintf(`	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
`))
			}
		}

		handlers.WriteString("\n")
		handlers.WriteString(fmt.Sprintf("	// Response\n"))
		handlers.WriteString(fmt.Sprintf("	c.JSON(http.StatusOK, gin.H{\n"))
		handlers.WriteString(fmt.Sprintf("		\"message\": \"%s endpoint\",\n", route.Function))
		handlers.WriteString(fmt.Sprintf("		\"method\": \"%s\",\n", route.Method))
		handlers.WriteString(fmt.Sprintf("		\"path\": \"%s\",\n", route.Path))
		handlers.WriteString(fmt.Sprintf("		\"timestamp\": time.Now().UTC(),\n"))
		handlers.WriteString(fmt.Sprintf("		\"auto_generated\": true,\n"))
		handlers.WriteString(fmt.Sprintf("	}))\n"))
		handlers.WriteString("}\n\n")
	}

	return handlers.String(), nil
}

func (g *GinGenerator) GenerateRoutes(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var routesBuilder strings.Builder

	routesBuilder.WriteString("package main\n\n")
	routesBuilder.WriteString("import (\n")
	routesBuilder.WriteString(`	"github.com/gin-gonic/gin"
` + "}\n\n")

	routesBuilder.WriteString("// setupRoutes configures all API routes\n")
	routesBuilder.WriteString("func (s *Server) setupRoutes() {\n")
	routesBuilder.WriteString("	// Health check\n")
	routesBuilder.WriteString("	s.router.GET(\"/health\", s.healthCheck)\n\n")

	routesBuilder.WriteString("	// API v1 routes\n")
	routesBuilder.WriteString("	v1 := s.router.Group(\"/api/v1\")\n")

	// Check if auth is enabled
	authEnabled := false
	if config.Auth != nil && config.Auth.Required {
		authEnabled = true
		routesBuilder.WriteString("	// Authenticated routes\n")
		routesBuilder.WriteString("	auth := v1.Group(\"/\", AuthMiddleware(s.config.JWTSecret))\n")
		routesBuilder.WriteString("	{\n")
	}

	// Generate route definitions
	for _, route := range routes {
		handlerName := toCamelCase(route.Function) + "Handler"
		routePath := route.Path

		// Convert path parameters to Gin format
		routePath = strings.ReplaceAll(routePath, "{id}", ":id")
		routePath = strings.ReplaceAll(routePath, "{field}", ":field")

		routeDef := fmt.Sprintf("		%s.%s(\"%s\", s.%s)",
			getRouteGroup(authEnabled, route.Auth.Required),
			strings.ToUpper(route.Method),
			routePath,
			handlerName)

		if authEnabled && route.Auth.Required {
			routesBuilder.WriteString(routeDef + "\n")
		} else if !authEnabled {
			routesBuilder.WriteString("	v1." + strings.ToUpper(route.Method) + "(\"" + routePath + "\", s." + handlerName + ")\n")
		}
	}

	if authEnabled {
		routesBuilder.WriteString("	}\n")
	}

	routesBuilder.WriteString("}\n\n")

	// Health check handler
	routesBuilder.WriteString("// healthCheck returns the health status of the server\n")
	routesBuilder.WriteString("func (s *Server) healthCheck(c *gin.Context) {\n")
	routesBuilder.WriteString("	c.JSON(http.StatusOK, gin.H{\n")
	routesBuilder.WriteString("		\"status\": \"healthy\",\n")
	routesBuilder.WriteString("		\"timestamp\": time.Now().UTC(),\n")
	routesBuilder.WriteString("		\"version\": \"1.0.0\",\n")
	routesBuilder.WriteString("		\"framework\": \"gin\",\n")
	routesBuilder.WriteString("	})\n")
	routesBuilder.WriteString("}\n")

	return routesBuilder.String(), nil
}

func (g *GinGenerator) GenerateModels(structs []StructInfo, config *FrameworkConfig) (string, error) {
	var models strings.Builder

	models.WriteString("package main\n\n")
	models.WriteString("import (\n")
	models.WriteString(`	"time"
` + "}\n\n")

	for _, structInfo := range structs {
		models.WriteString(fmt.Sprintf("// %s represents the %s entity\n", structInfo.Name, strings.ToLower(structInfo.Name)))
		models.WriteString(fmt.Sprintf("type %s struct {\n", structInfo.Name))

		// Add standard fields
		models.WriteString("	ID        string    `json:\"id\" gorm:\"primaryKey\"`\n")
		models.WriteString("	CreatedAt time.Time `json:\"created_at\"`\n")
		models.WriteString("	UpdatedAt time.Time `json:\"updated_at\"`\n")

		// Add struct fields
		for _, field := range structInfo.Fields {
			if field.Name != "ID" && field.Name != "CreatedAt" && field.Name != "UpdatedAt" {
				jsonTag := strings.ToLower(field.Name)
				models.WriteString(fmt.Sprintf("	%s    %s    `json:\"%s\"`\n", field.Name, field.Type, jsonTag))
			}
		}

		models.WriteString("}\n\n")
	}

	return models.String(), nil
}

func (g *GinGenerator) GenerateTests(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var tests strings.Builder

	tests.WriteString("package main\n\n")
	tests.WriteString("import (\n")
	tests.WriteString(`	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
` + "}\n\n")

	tests.WriteString("func setupTestRouter() *gin.Engine {\n")
	tests.WriteString("	gin.SetMode(gin.TestMode)\n")
	tests.WriteString("	router := gin.New()\n")
	tests.WriteString("	server := NewServer()\n")
	tests.WriteString("	server.router = router\n")
	tests.WriteString("	server.setupRoutes()\n")
	tests.WriteString("	return router\n")
	tests.WriteString("}\n\n")

	// Generate health check test
	tests.WriteString("func TestHealthCheck(t *testing.T) {\n")
	tests.WriteString("	router := setupTestRouter()\n")
	tests.WriteString("	w := httptest.NewRecorder()\n")
	tests.WriteString("	req, _ := http.NewRequest(\"GET\", \"/health\", nil)\n")
	tests.WriteString("	router.ServeHTTP(w, req)\n")
	tests.WriteString("	assert.Equal(t, http.StatusOK, w.Code)\n")
	tests.WriteString("	var response map[string]interface{}\n")
	tests.WriteString("	err := json.Unmarshal(w.Body.Bytes(), &response)\n")
	tests.WriteString("	assert.NoError(t, err)\n")
	tests.WriteString("	assert.Equal(t, \"healthy\", response[\"status\"])\n")
	tests.WriteString("}\n\n")

	// Generate tests for each route
	for _, route := range routes {
		testName := fmt.Sprintf("Test%s", toCamelCase(route.Function))
		tests.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
		tests.WriteString("	router := setupTestRouter()\n")
		tests.WriteString("	w := httptest.NewRecorder()\n")

		// Generate request based on method
		switch route.Method {
		case "GET":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString(fmt.Sprintf("	req, _ := http.NewRequest(\"GET\", \"%s\", nil)\n", path))
		case "POST":
			tests.WriteString(fmt.Sprintf("	body := bytes.NewBuffer([]byte(\"{}\"))\n"))
			tests.WriteString(fmt.Sprintf("	req, _ := http.NewRequest(\"POST\", \"%s\", body)\n", route.Path))
			tests.WriteString("	req.Header.Set(\"Content-Type\", \"application/json\")\n")
		case "PUT":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString(fmt.Sprintf("	body := bytes.NewBuffer([]byte(\"{}\"))\n"))
			tests.WriteString(fmt.Sprintf("	req, _ := http.NewRequest(\"PUT\", \"%s\", body)\n", path))
			tests.WriteString("	req.Header.Set(\"Content-Type\", \"application/json\")\n")
		case "DELETE":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString(fmt.Sprintf("	req, _ := http.NewRequest(\"DELETE\", \"%s\", nil)\n", path))
		}

		tests.WriteString("	router.ServeHTTP(w, req)\n")
		tests.WriteString("	assert.Equal(t, http.StatusOK, w.Code)\n")
		tests.WriteString("	var response map[string]interface{}\n")
		tests.WriteString("	err := json.Unmarshal(w.Body.Bytes(), &response)\n")
		tests.WriteString("	assert.NoError(t, err)\n")
		tests.WriteString("	assert.Equal(t, true, response[\"auto_generated\"])\n")
		tests.WriteString("}\n\n")
	}

	return tests.String(), nil
}

func (g *GinGenerator) GenerateDocs(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var docs strings.Builder

	docs.WriteString("# API Documentation\n\n")
	docs.WriteString(fmt.Sprintf("Generated %s API Documentation\n\n", strings.Title(string(config.Type))))

	docs.WriteString("## Base URL\n")
	docs.WriteString("```\nhttp://localhost:8080/api/v1\n```\n\n")

	docs.WriteString("## Authentication\n")
	docs.WriteString("Add JWT token to Authorization header:\n")
	docs.WriteString("```\nAuthorization: Bearer <token>\n```\n\n")

	docs.WriteString("## Endpoints\n\n")
	docs.WriteString("### Health Check\n")
	docs.WriteString("```\nGET /health\n```\n\n")

	for _, route := range routes {
		docs.WriteString(fmt.Sprintf("### %s %s\n", strings.ToUpper(route.Method), route.Path))
		docs.WriteString(fmt.Sprintf("**Description**: %s endpoint\n\n", route.Function))

		if len(route.Parameter) > 0 {
			docs.WriteString("**Parameters**:\n")
			for _, param := range route.Parameter {
				docs.WriteString(fmt.Sprintf("- `%s` (%s): %s\n", param.Name, param.Type, "parameter description"))
			}
			docs.WriteString("\n")
		}

		if len(route.Response) > 0 {
			docs.WriteString("**Response**:\n")
			for _, resp := range route.Response {
				docs.WriteString(fmt.Sprintf("- `%s`: %s\n", resp.Type, "response data"))
			}
			docs.WriteString("\n")
		}

		docs.WriteString("```bash\n")
		switch route.Method {
		case "GET":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			docs.WriteString(fmt.Sprintf("curl -X GET http://localhost:8080/api/v1%s\n", path))
		case "POST":
			docs.WriteString(fmt.Sprintf("curl -X POST http://localhost:8080/api/v1%s \\\n", route.Path))
			docs.WriteString("  -H \"Content-Type: application/json\" \\\n")
			docs.WriteString("  -d '{}'\n")
		case "PUT":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			docs.WriteString(fmt.Sprintf("curl -X PUT http://localhost:8080/api/v1%s \\\n", path))
			docs.WriteString("  -H \"Content-Type: application/json\" \\\n")
			docs.WriteString("  -d '{}'\n")
		case "DELETE":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			docs.WriteString(fmt.Sprintf("curl -X DELETE http://localhost:8080/api/v1%s\n", path))
		}
		docs.WriteString("```\n\n")
	}

	return docs.String(), nil
}

func (g *GinGenerator) GenerateDockerfile(config *FrameworkConfig) (string, error) {
	return `# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
`, nil
}

func (g *GinGenerator) GenerateK8sManifests(config *FrameworkConfig) (map[string]string, error) {
	manifests := make(map[string]string)

	// Deployment
	deployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: generated-%s-api
  labels:
    app: generated-%s-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: generated-%s-api
  template:
    metadata:
      labels:
        app: generated-%s-api
    spec:
      containers:
      - name: api
        image: generated-%s-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: GIN_MODE
          value: "release"
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
`, string(config.Type), string(config.Type), string(config.Type), string(config.Type), string(config.Type))

	// Service
	service := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: generated-%s-api-service
spec:
  selector:
    app: generated-%s-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
`, string(config.Type), string(config.Type))

	// Ingress
	ingress := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: generated-%s-api-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: generated-%s-api-service
            port:
              number: 80
`, string(config.Type), string(config.Type))

	manifests["deployment.yaml"] = deployment
	manifests["service.yaml"] = service
	manifests["ingress.yaml"] = ingress

	return manifests, nil
}

// Echo Framework Generator
type EchoGenerator struct{}

func NewEchoGenerator() FrameworkGenerator {
	return &EchoGenerator{}
}

func (e *EchoGenerator) GetName() string { return "Echo" }
func (e *EchoGenerator) GetType() FrameworkType { return FrameworkEcho }
func (e *EchoGenerator) GetDefaultConfig() *FrameworkConfig {
	return &FrameworkConfig{
		Type:     FrameworkEcho,
		Version:  "v4.11.4",
		Features: []string{"middleware", "validation", "cors", "jwt"},
		Middleware: []string{"logger", "recover", "cors", "auth"},
		CORS: &CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		},
		Docs: &DocumentationConfig{
			Enabled: true,
			Path:    "/swagger",
			Format:  "openapi",
		},
		Testing: &TestingConfig{
			Enabled:   true,
			Framework: "testify",
			Coverage:  true,
		},
	}
}

func (e *EchoGenerator) GenerateMainFile(routes []APIRoute, config *FrameworkConfig) (string, error) {
	return fmt.Sprintf(`package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Create Echo instance
	e := echo.New()

	// Hide Echo banner
	e.HideBanner = true

	// Create server
	server := NewServer(e)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting %s server on port %%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %%v", err)
	}
}`, strings.Title(string(config.Type))), nil
}

func (e *EchoGenerator) GenerateMiddleware(config *FrameworkConfig) (string, error) {
	return fmt.Sprintf(`package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/golang-jwt/jwt/v4"
)

// setupMiddleware configures all middleware for the Echo server
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.e.Use(middleware.Recover())

	// Logger middleware
	s.e.Use(middleware.Logger())

	// CORS middleware
	if %t {
		s.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     %s,
			AllowMethods:     %s,
			AllowHeaders:     %s,
			ExposeHeaders:    %s,
			AllowCredentials: %t,
			MaxAge:           %d,
		}))
	}

	// Request ID middleware
	s.e.Use(middleware.RequestID())

	// Security headers middleware
	s.e.Use(securityHeadersMiddleware())
}

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Authorization header required",
				})
			}

			tokenString := authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error": "Invalid token",
				})
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user_id", claims["user_id"])
				c.Set("username", claims["username"])
			}

			return next(c)
		}
	}
}

// securityHeadersMiddleware adds security headers
func securityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			return next(c)
		}
	}
}
`,
		config.CORS.Enabled,
		formatStringSlice(config.CORS.AllowOrigins),
		formatStringSlice(config.CORS.AllowMethods),
		formatStringSlice(config.CORS.AllowHeaders),
		formatStringSlice(config.CORS.ExposeHeaders),
		config.CORS.AllowCredentials,
		config.CORS.MaxAge,
	), nil
}

func (e *EchoGenerator) GenerateHandlers(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var handlers strings.Builder

	handlers.WriteString("package main\n\n")
	handlers.WriteString("import (\n")
	handlers.WriteString(`	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
` + "}\n\n")

	for _, route := range routes {
		handlerName := toCamelCase(route.Function) + "Handler"
		handlers.WriteString(fmt.Sprintf("// %s handles %s %s\n", handlerName, strings.ToUpper(route.Method), route.Path))
		handlers.WriteString(fmt.Sprintf("func (s *Server) %s(c echo.Context) error {\n", handlerName))
		handlers.WriteString(fmt.Sprintf("	// TODO: Implement business logic for %s\n\n", route.Function))

		// Generate parameter extraction
		for _, param := range route.Parameter {
			if param.Name == "id" {
				handlers.WriteString("	id := c.Param(\"id\")\n")
			} else if param.Name == "q" {
				handlers.WriteString("	q := c.QueryParam(\"q\")\n")
			} else if param.Name == "limit" {
				handlers.WriteString("	limit, _ := strconv.Atoi(c.QueryParam(\"limit\"))\n")
			} else if param.Name == "offset" {
				handlers.WriteString("	offset, _ := strconv.Atoi(c.QueryParam(\"offset\"))\n")
			}
		}

		handlers.WriteString("\n")
		handlers.WriteString("	return c.JSON(http.StatusOK, map[string]interface{}{\n")
		handlers.WriteString(fmt.Sprintf("		\"message\": \"%s endpoint\",\n", route.Function))
		handlers.WriteString(fmt.Sprintf("		\"method\": \"%s\",\n", route.Method))
		handlers.WriteString(fmt.Sprintf("		\"path\": \"%s\",\n", route.Path))
		handlers.WriteString("		\"timestamp\": time.Now().UTC(),\n")
		handlers.WriteString("		\"auto_generated\": true,\n")
		handlers.WriteString("	})\n")
		handlers.WriteString("}\n\n")
	}

	return handlers.String(), nil
}

func (e *EchoGenerator) GenerateRoutes(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var routesBuilder strings.Builder

	routesBuilder.WriteString("package main\n\n")
	routesBuilder.WriteString("import (\n")
	routesBuilder.WriteString(`	"net/http"

	"github.com/labstack/echo/v4"
` + "}\n\n")

	routesBuilder.WriteString("// setupRoutes configures all API routes\n")
	routesBuilder.WriteString("func (s *Server) setupRoutes() {\n")
	routesBuilder.WriteString("	// Health check\n")
	routesBuilder.WriteString("	s.e.GET(\"/health\", s.healthCheck)\n\n")

	// Check if auth is enabled
	authEnabled := false
	if config.Auth != nil && config.Auth.Required {
		authEnabled = true
	}

	// Generate route definitions
	for _, route := range routes {
		handlerName := toCamelCase(route.Function) + "Handler"
		routePath := route.Path

		// Echo already uses :param format, so just replace {id} with :id
		routePath = strings.ReplaceAll(routePath, "{id}", ":id")
		routePath = strings.ReplaceAll(routePath, "{field}", ":field")

		if authEnabled && route.Auth.Required {
			routesBuilder.WriteString(fmt.Sprintf("	s.e.%s(\"%s\", AuthMiddleware(s.config.JWTSecret)(s.%s))\n",
				strings.ToLower(route.Method),
				routePath,
				handlerName))
		} else {
			routesBuilder.WriteString(fmt.Sprintf("	s.e.%s(\"%s\", s.%s)\n",
				strings.ToLower(route.Method),
				routePath,
				handlerName))
		}
	}

	routesBuilder.WriteString("}\n\n")

	// Health check handler
	routesBuilder.WriteString("// healthCheck returns the health status of the server\n")
	routesBuilder.WriteString("func (s *Server) healthCheck(c echo.Context) error {\n")
	routesBuilder.WriteString("	return c.JSON(http.StatusOK, map[string]interface{}{\n")
	routesBuilder.WriteString("		\"status\": \"healthy\",\n")
	routesBuilder.WriteString("		\"timestamp\": time.Now().UTC(),\n")
	routesBuilder.WriteString("		\"version\": \"1.0.0\",\n")
	routesBuilder.WriteString("		\"framework\": \"echo\",\n")
	routesBuilder.WriteString("	})\n")
	routesBuilder.WriteString("}\n")

	return routesBuilder.String(), nil
}

func (e *EchoGenerator) GenerateModels(structs []StructInfo, config *FrameworkConfig) (string, error) {
	// Echo uses the same model generation as Gin
	return (&GinGenerator{}).GenerateModels(structs, config)
}

func (e *EchoGenerator) GenerateTests(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var tests strings.Builder

	tests.WriteString("package main\n\n")
	tests.WriteString("import (\n")
	tests.WriteString(`	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
` + "}\n\n")

	tests.WriteString("func setupTestEcho() *echo.Echo {\n")
	tests.WriteString("	e := echo.New()\n")
	tests.WriteString("	server := NewServer(e)\n")
	tests.WriteString("	server.setupRoutes()\n")
	tests.WriteString("	return e\n")
	tests.WriteString("}\n\n")

	// Generate health check test
	tests.WriteString("func TestHealthCheck(t *testing.T) {\n")
	tests.WriteString("	e := setupTestEcho()\n")
	tests.WriteString("	req := httptest.NewRequest(http.MethodGet, \"/health\", nil)\n")
	tests.WriteString("	rec := httptest.NewRecorder()\n")
	tests.WriteString("	e.ServeHTTP(rec, req)\n")
	tests.WriteString("	assert.Equal(t, http.StatusOK, rec.Code)\n")
	tests.WriteString("	var response map[string]interface{}\n")
	tests.WriteString("	err := json.Unmarshal(rec.Body.Bytes(), &response)\n")
	tests.WriteString("	assert.NoError(t, err)\n")
	tests.WriteString("	assert.Equal(t, \"healthy\", response[\"status\"])\n")
	tests.WriteString("}\n\n")

	// Generate tests for each route
	for _, route := range routes {
		testName := fmt.Sprintf("Test%s", toCamelCase(route.Function))
		tests.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
		tests.WriteString("	e := setupTestEcho()\n")

		// Generate request based on method
		switch route.Method {
		case "GET":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			path = strings.ReplaceAll(path, "{field}", "test")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodGet, \"%s\", nil)\n", path))
		case "POST":
			tests.WriteString("	body := bytes.NewBuffer([]byte(\"{}\"))\n")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodPost, \"%s\", body)\n", route.Path))
			tests.WriteString("	req.Header.Set(\"Content-Type\", \"application/json\")\n")
		case "PUT":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString("	body := bytes.NewBuffer([]byte(\"{}\"))\n")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodPut, \"%s\", body)\n", path))
			tests.WriteString("	req.Header.Set(\"Content-Type\", \"application/json\")\n")
		case "DELETE":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodDelete, \"%s\", nil)\n", path))
		}

		tests.WriteString("	rec := httptest.NewRecorder()\n")
		tests.WriteString("	e.ServeHTTP(rec, req)\n")
		tests.WriteString("	assert.Equal(t, http.StatusOK, rec.Code)\n")
		tests.WriteString("	var response map[string]interface{}\n")
		tests.WriteString("	err := json.Unmarshal(rec.Body.Bytes(), &response)\n")
		tests.WriteString("	assert.NoError(t, err)\n")
		tests.WriteString("	assert.Equal(t, true, response[\"auto_generated\"])\n")
		tests.WriteString("}\n\n")
	}

	return tests.String(), nil
}

func (e *EchoGenerator) GenerateDocs(routes []APIRoute, config *FrameworkConfig) (string, error) {
	// Echo uses the same documentation generation as Gin
	return (&GinGenerator{}).GenerateDocs(routes, config)
}

func (e *EchoGenerator) GenerateDockerfile(config *FrameworkConfig) (string, error) {
	return (&GinGenerator{}).GenerateDockerfile(config)
}

func (e *EchoGenerator) GenerateK8sManifests(config *FrameworkConfig) (map[string]string, error) {
	return (&GinGenerator{}).GenerateK8sManifests(config)
}

// Chi Framework Generator
type ChiGenerator struct{}

func NewChiGenerator() FrameworkGenerator {
	return &ChiGenerator{}
}

func (c *ChiGenerator) GetName() string { return "Chi" }
func (c *ChiGenerator) GetType() FrameworkType { return FrameworkChi }
func (c *ChiGenerator) GetDefaultConfig() *FrameworkConfig {
	return &FrameworkConfig{
		Type:     FrameworkChi,
		Version:  "v5.0.12",
		Features: []string{"middleware", "router", "cors"},
		Middleware: []string{"logger", "recover", "cors", "auth"},
		CORS: &CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		},
		Docs: &DocumentationConfig{
			Enabled: true,
			Path:    "/docs",
			Format:  "openapi",
		},
		Testing: &TestingConfig{
			Enabled:   true,
			Framework: "testify",
			Coverage:  true,
		},
	}
}

func (c *ChiGenerator) GenerateMainFile(routes []APIRoute, config *FrameworkConfig) (string, error) {
	return fmt.Sprintf(`package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Create Chi router
	r := chi.NewRouter()

	// Create server
	server := NewServer(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	go func() {
		if err := http.ListenAndServe(":"+port, server.router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %%v", err)
		}
	}()

	log.Printf("Starting %s server on port %%s", port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.router.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %%v", err)
	}

	log.Println("Server exited")
}`, strings.Title(string(config.Type))), nil
}

func (c *ChiGenerator) GenerateMiddleware(config *FrameworkConfig) (string, error) {
	return fmt.Sprintf(`package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v4"
)

// setupMiddleware configures all middleware for the Chi router
func (s *Server) setupMiddleware() {
	// Chi built-in middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	if %t {
		s.router.Use(corsMiddleware(%s, %s, %s, %t, %d))
	}

	// Security headers middleware
	s.router.Use(securityHeadersMiddleware())
}

// corsMiddleware creates CORS middleware
func corsMiddleware(origins, methods, headers []string, credentials bool, maxAge int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", strings.Join(origins, ", "))
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%%d", maxAge))
			w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprintf("%%t", credentials))

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
				ctx = context.WithValue(ctx, "username", claims["username"])
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// securityHeadersMiddleware adds security headers
func securityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			next.ServeHTTP(w, r)
		})
	}
}
`,
		config.CORS.Enabled,
		formatStringSlice(config.CORS.AllowOrigins),
		formatStringSlice(config.CORS.AllowMethods),
		formatStringSlice(config.CORS.AllowHeaders),
		config.CORS.AllowCredentials,
		config.CORS.MaxAge,
	), nil
}

func (c *ChiGenerator) GenerateHandlers(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var handlers strings.Builder

	handlers.WriteString("package main\n\n")
	handlers.WriteString("import (\n")
	handlers.WriteString(`	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
` + "}\n\n")

	for _, route := range routes {
		handlerName := toCamelCase(route.Function) + "Handler"
		handlers.WriteString(fmt.Sprintf("// %s handles %s %s\n", handlerName, strings.ToUpper(route.Method), route.Path))
		handlers.WriteString(fmt.Sprintf("func (s *Server) %s(w http.ResponseWriter, r *http.Request) {\n", handlerName))
		handlers.WriteString(fmt.Sprintf("	// TODO: Implement business logic for %s\n\n", route.Function))

		// Generate parameter extraction
		for _, param := range route.Parameter {
			if param.Name == "id" {
				handlers.WriteString("	id := chi.URLParam(r, \"id\")\n")
			} else if param.Name == "q" {
				handlers.WriteString("	q := r.URL.Query().Get(\"q\")\n")
			} else if param.Name == "limit" {
				handlers.WriteString("	limit, _ := strconv.Atoi(r.URL.Query().Get(\"limit\"))\n")
			} else if param.Name == "offset" {
				handlers.WriteString("	offset, _ := strconv.Atoi(r.URL.Query().Get(\"offset\"))\n")
			}
		}

		handlers.WriteString("\n")
		handlers.WriteString("	response := map[string]interface{}{\n")
		handlers.WriteString(fmt.Sprintf("		\"message\": \"%s endpoint\",\n", route.Function))
		handlers.WriteString(fmt.Sprintf("		\"method\": \"%s\",\n", route.Method))
		handlers.WriteString(fmt.Sprintf("		\"path\": \"%s\",\n", route.Path))
		handlers.WriteString("		\"timestamp\": time.Now().UTC(),\n")
		handlers.WriteString("		\"auto_generated\": true,\n")
		handlers.WriteString("	}\n\n")

		handlers.WriteString("	w.Header().Set(\"Content-Type\", \"application/json\")\n")
		handlers.WriteString("	json.NewEncoder(w).Encode(response)\n")
		handlers.WriteString("}\n\n")
	}

	return handlers.String(), nil
}

func (c *ChiGenerator) GenerateRoutes(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var routesBuilder strings.Builder

	routesBuilder.WriteString("package main\n\n")
	routesBuilder.WriteString("import (\n")
	routesBuilder.WriteString(`	"net/http"

	"github.com/go-chi/chi/v5"
` + "}\n\n")

	routesBuilder.WriteString("// setupRoutes configures all API routes\n")
	routesBuilder.WriteString("func (s *Server) setupRoutes() {\n")
	routesBuilder.WriteString("	// Health check\n")
	routesBuilder.WriteString("	s.router.Get(\"/health\", s.healthCheckHandler)\n\n")

	// Check if auth is enabled
	authEnabled := false
	if config.Auth != nil && config.Auth.Required {
		authEnabled = true
	}

	// Generate route definitions
	for _, route := range routes {
		handlerName := toCamelCase(route.Function) + "Handler"
		routePath := route.Path

		// Chi uses {param} format, so no conversion needed
		if authEnabled && route.Auth.Required {
			routesBuilder.WriteString(fmt.Sprintf("	s.router.With(AuthMiddleware(s.config.JWTSecret)).%s(\"%s\", s.%s)\n",
				strings.ToLower(route.Method),
				routePath,
				handlerName))
		} else {
			routesBuilder.WriteString(fmt.Sprintf("	s.router.%s(\"%s\", s.%s)\n",
				strings.ToLower(route.Method),
				routePath,
				handlerName))
		}
	}

	routesBuilder.WriteString("}\n\n")

	// Health check handler
	routesBuilder.WriteString("// healthCheckHandler returns the health status of the server\n")
	routesBuilder.WriteString("func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {\n")
	routesBuilder.WriteString("	response := map[string]interface{}{\n")
	routesBuilder.WriteString("		\"status\": \"healthy\",\n")
	routesBuilder.WriteString("		\"timestamp\": time.Now().UTC(),\n")
	routesBuilder.WriteString("		\"version\": \"1.0.0\",\n")
	routesBuilder.WriteString("		\"framework\": \"chi\",\n")
	routesBuilder.WriteString("	}\n")
	routesBuilder.WriteString("	w.Header().Set(\"Content-Type\", \"application/json\")\n")
	routesBuilder.WriteString("	json.NewEncoder(w).Encode(response)\n")
	routesBuilder.WriteString("}\n")

	return routesBuilder.String(), nil
}

func (c *ChiGenerator) GenerateModels(structs []StructInfo, config *FrameworkConfig) (string, error) {
	// Chi uses the same model generation as Gin
	return (&GinGenerator{}).GenerateModels(structs, config)
}

func (c *ChiGenerator) GenerateTests(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var tests strings.Builder

	tests.WriteString("package main\n\n")
	tests.WriteString("import (\n")
	tests.WriteString(`	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
` + "}\n\n")

	tests.WriteString("func setupTestChi() http.Handler {\n")
	tests.WriteString("	r := chi.NewRouter()\n")
	tests.WriteString("	server := NewServer(r)\n")
	tests.WriteString("	server.setupRoutes()\n")
	tests.WriteString("	return r\n")
	tests.WriteString("}\n\n")

	// Generate health check test
	tests.WriteString("func TestHealthCheck(t *testing.T) {\n")
	tests.WriteString("	handler := setupTestChi()\n")
	tests.WriteString("	req := httptest.NewRequest(http.MethodGet, \"/health\", nil)\n")
	tests.WriteString("	rec := httptest.NewRecorder()\n")
	tests.WriteString("	handler.ServeHTTP(rec, req)\n")
	tests.WriteString("	assert.Equal(t, http.StatusOK, rec.Code)\n")
	tests.WriteString("	var response map[string]interface{}\n")
	tests.WriteString("	err := json.Unmarshal(rec.Body.Bytes(), &response)\n")
	tests.WriteString("	assert.NoError(t, err)\n")
	tests.WriteString("	assert.Equal(t, \"healthy\", response[\"status\"])\n")
	tests.WriteString("}\n\n")

	// Generate tests for each route
	for _, route := range routes {
		testName := fmt.Sprintf("Test%s", toCamelCase(route.Function))
		tests.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
		tests.WriteString("	handler := setupTestChi()\n")

		// Generate request based on method
		switch route.Method {
		case "GET":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodGet, \"%s\", nil)\n", path))
		case "POST":
			tests.WriteString("	body := bytes.NewBuffer([]byte(\"{}\"))\n")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodPost, \"%s\", body)\n", route.Path))
			tests.WriteString("	req.Header.Set(\"Content-Type\", \"application/json\")\n")
		case "PUT":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString("	body := bytes.NewBuffer([]byte(\"{}\"))\n")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodPut, \"%s\", body)\n", path))
			tests.WriteString("	req.Header.Set(\"Content-Type\", \"application/json\")\n")
		case "DELETE":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodDelete, \"%s\", nil)\n", path))
		}

		tests.WriteString("	rec := httptest.NewRecorder()\n")
		tests.WriteString("	handler.ServeHTTP(rec, req)\n")
		tests.WriteString("	assert.Equal(t, http.StatusOK, rec.Code)\n")
		tests.WriteString("	var response map[string]interface{}\n")
		tests.WriteString("	err := json.Unmarshal(rec.Body.Bytes(), &response)\n")
		tests.WriteString("	assert.NoError(t, err)\n")
		tests.WriteString("	assert.Equal(t, true, response[\"auto_generated\"])\n")
		tests.WriteString("}\n\n")
	}

	return tests.String(), nil
}

func (c *ChiGenerator) GenerateDocs(routes []APIRoute, config *FrameworkConfig) (string, error) {
	// Chi uses the same documentation generation as Gin
	return (&GinGenerator{}).GenerateDocs(routes, config)
}

func (c *ChiGenerator) GenerateDockerfile(config *FrameworkConfig) (string, error) {
	return (&GinGenerator{}).GenerateDockerfile(config)
}

func (c *ChiGenerator) GenerateK8sManifests(config *FrameworkConfig) (map[string]string, error) {
	return (&GinGenerator{}).GenerateK8sManifests(config)
}

// Fiber Framework Generator
type FiberGenerator struct{}

func NewFiberGenerator() FrameworkGenerator {
	return &FiberGenerator{}
}

func (f *FiberGenerator) GetName() string { return "Fiber" }
func (f *FiberGenerator) GetType() FrameworkType { return FrameworkFiber }
func (f *FiberGenerator) GetDefaultConfig() *FrameworkConfig {
	return &FrameworkConfig{
		Type:     FrameworkFiber,
		Version:  "v2.52.4",
		Features: []string{"middleware", "validation", "cors", "jwt"},
		Middleware: []string{"logger", "recover", "cors", "auth"},
		CORS: &CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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

func (f *FiberGenerator) GenerateMainFile(routes []APIRoute, config *FrameworkConfig) (string, error) {
	return fmt.Sprintf(`package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Create Fiber instance
	app := fiber.New()

	// Create server
	server := NewServer(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting %s server on port %%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %%v", err)
	}
}`, strings.Title(string(config.Type))), nil
}

func (f *FiberGenerator) GenerateMiddleware(config *FrameworkConfig) (string, error) {
	return fmt.Sprintf(`package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v4"
)

// setupMiddleware configures all middleware for the Fiber app
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.app.Use(recover.New())

	// Logger middleware
	s.app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))

	// CORS middleware
	if %t {
		s.app.Use(cors.New(cors.Config{
			AllowOrigins:     %s,
			AllowMethods:     %s,
			AllowHeaders:     %s,
			ExposeHeaders:    %s,
			AllowCredentials: %t,
			MaxAge:           %d * time.Hour,
		}))
	}

	// Request ID middleware
	s.app.Use(requestIDMiddleware())

	// Security headers middleware
	s.app.Use(securityHeadersMiddleware())
}

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("user_id", claims["user_id"])
			c.Locals("username", claims["username"])
		}

		return c.Next()
	}
}

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateUUID()
		}
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)
		return c.Next()
	}
}

// securityHeadersMiddleware adds security headers
func securityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		return c.Next()
	}
}
`,
		config.CORS.Enabled,
		formatStringSlice(config.CORS.AllowOrigins),
		formatStringSlice(config.CORS.AllowMethods),
		formatStringSlice(config.CORS.AllowHeaders),
		formatStringSlice(config.CORS.ExposeHeaders),
		config.CORS.AllowCredentials,
		config.CORS.MaxAge/3600,
	), nil
}

func (f *FiberGenerator) GenerateHandlers(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var handlers strings.Builder

	handlers.WriteString("package main\n\n")
	handlers.WriteString("import (\n")
	handlers.WriteString(`	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
` + "}\n\n")

	for _, route := range routes {
		handlerName := toCamelCase(route.Function) + "Handler"
		handlers.WriteString(fmt.Sprintf("// %s handles %s %s\n", handlerName, strings.ToUpper(route.Method), route.Path))
		handlers.WriteString(fmt.Sprintf("func (s *Server) %s(c *fiber.Ctx) error {\n", handlerName))
		handlers.WriteString(fmt.Sprintf("	// TODO: Implement business logic for %s\n\n", route.Function))

		// Generate parameter extraction
		for _, param := range route.Parameter {
			if param.Name == "id" {
				handlers.WriteString("	id := c.Params(\"id\")\n")
			} else if param.Name == "q" {
				handlers.WriteString("	q := c.Query(\"q\")\n")
			} else if param.Name == "limit" {
				handlers.WriteString("	limit, _ := strconv.Atoi(c.Query(\"limit\", \"10\"))\n")
			} else if param.Name == "offset" {
				handlers.WriteString("	offset, _ := strconv.Atoi(c.Query(\"offset\", \"0\"))\n")
			}
		}

		handlers.WriteString("\n")
		handlers.WriteString("	return c.JSON(fiber.StatusOK, fiber.Map{\n")
		handlers.WriteString(fmt.Sprintf("		\"message\": \"%s endpoint\",\n", route.Function))
		handlers.WriteString(fmt.Sprintf("		\"method\": \"%s\",\n", route.Method))
		handlers.WriteString(fmt.Sprintf("		\"path\": \"%s\",\n", route.Path))
		handlers.WriteString("		\"timestamp\": time.Now().UTC(),\n")
		handlers.WriteString("		\"auto_generated\": true,\n")
		handlers.WriteString("	})\n")
		handlers.WriteString("}\n\n")
	}

	return handlers.String(), nil
}

func (f *FiberGenerator) GenerateRoutes(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var routesBuilder strings.Builder

	routesBuilder.WriteString("package main\n\n")
	routesBuilder.WriteString("import (\n")
	routesBuilder.WriteString(`	"time"

	"github.com/gofiber/fiber/v2"
` + "}\n\n")

	routesBuilder.WriteString("// setupRoutes configures all API routes\n")
	routesBuilder.WriteString("func (s *Server) setupRoutes() {\n")
	routesBuilder.WriteString("	// Health check\n")
	routesBuilder.WriteString("	s.app.Get(\"/health\", s.healthCheckHandler)\n\n")

	// Check if auth is enabled
	authEnabled := false
	if config.Auth != nil && config.Auth.Required {
		authEnabled = true
	}

	// Generate route definitions
	for _, route := range routes {
		handlerName := toCamelCase(route.Function) + "Handler"
		routePath := route.Path

		// Fiber uses :param format
		routePath = strings.ReplaceAll(routePath, "{id}", ":id")
		routePath = strings.ReplaceAll(routePath, "{field}", ":field")

		if authEnabled && route.Auth.Required {
			routesBuilder.WriteString(fmt.Sprintf("	s.app.%s(\"%s\", AuthMiddleware(s.config.JWTSecret), s.%s)\n",
				strings.ToLower(route.Method),
				routePath,
				handlerName))
		} else {
			routesBuilder.WriteString(fmt.Sprintf("	s.app.%s(\"%s\", s.%s)\n",
				strings.ToLower(route.Method),
				routePath,
				handlerName))
		}
	}

	routesBuilder.WriteString("}\n\n")

	// Health check handler
	routesBuilder.WriteString("// healthCheckHandler returns the health status of the server\n")
	routesBuilder.WriteString("func (s *Server) healthCheckHandler(c *fiber.Ctx) error {\n")
	routesBuilder.WriteString("	return c.JSON(fiber.StatusOK, fiber.Map{\n")
	routesBuilder.WriteString("		\"status\": \"healthy\",\n")
	routesBuilder.WriteString("		\"timestamp\": time.Now().UTC(),\n")
	routesBuilder.WriteString("		\"version\": \"1.0.0\",\n")
	routesBuilder.WriteString("		\"framework\": \"fiber\",\n")
	routesBuilder.WriteString("	})\n")
	routesBuilder.WriteString("}\n")

	return routesBuilder.String(), nil
}

func (f *FiberGenerator) GenerateModels(structs []StructInfo, config *FrameworkConfig) (string, error) {
	// Fiber uses the same model generation as Gin
	return (&GinGenerator{}).GenerateModels(structs, config)
}

func (f *FiberGenerator) GenerateTests(routes []APIRoute, config *FrameworkConfig) (string, error) {
	var tests strings.Builder

	tests.WriteString("package main\n\n")
	tests.WriteString("import (\n")
	tests.WriteString(`	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
` + "}\n\n")

	tests.WriteString("func setupTestFiber() *fiber.App {\n")
	tests.WriteString("	app := fiber.New()\n")
	tests.WriteString("	server := NewServer(app)\n")
	tests.WriteString("	server.setupRoutes()\n")
	tests.WriteString("	return app\n")
	tests.WriteString("}\n\n")

	// Generate health check test
	tests.WriteString("func TestHealthCheck(t *testing.T) {\n")
	tests.WriteString("	app := setupTestFiber()\n")
	tests.WriteString("	req := httptest.NewRequest(http.MethodGet, \"/health\", nil)\n")
	tests.WriteString("	resp, _ := app.Test(req)\n")
	tests.WriteString("	defer resp.Body.Close()\n")
	tests.WriteString("	assert.Equal(t, 200, resp.StatusCode)\n")
	tests.WriteString("	var response map[string]interface{}\n")
	tests.WriteString("	err := json.NewDecoder(resp.Body).Decode(&response)\n")
	tests.WriteString("	assert.NoError(t, err)\n")
	tests.WriteString("	assert.Equal(t, \"healthy\", response[\"status\"])\n")
	tests.WriteString("}\n\n")

	// Generate tests for each route
	for _, route := range routes {
		testName := fmt.Sprintf("Test%s", toCamelCase(route.Function))
		tests.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
		tests.WriteString("	app := setupTestFiber()\n")

		// Generate request based on method
		switch route.Method {
		case "GET":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			path = strings.ReplaceAll(path, "{field}", "test")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodGet, \"%s\", nil)\n", path))
		case "POST":
			tests.WriteString("	body := bytes.NewBuffer([]byte(\"{}\"))\n")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodPost, \"%s\", body)\n", route.Path))
			tests.WriteString("	req.Header.Set(\"Content-Type\", \"application/json\")\n")
		case "PUT":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString("	body := bytes.NewBuffer([]byte(\"{}\"))\n")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodPut, \"%s\", body)\n", path))
			tests.WriteString("	req.Header.Set(\"Content-Type\", \"application/json\")\n")
		case "DELETE":
			path := strings.ReplaceAll(route.Path, "{id}", "123")
			tests.WriteString(fmt.Sprintf("	req := httptest.NewRequest(http.MethodDelete, \"%s\", nil)\n", path))
		}

		tests.WriteString("	resp, _ := app.Test(req)\n")
		tests.WriteString("	defer resp.Body.Close()\n")
		tests.WriteString("	assert.Equal(t, 200, resp.StatusCode)\n")
		tests.WriteString("	var response map[string]interface{}\n")
		tests.WriteString("	err := json.NewDecoder(resp.Body).Decode(&response)\n")
		tests.WriteString("	assert.NoError(t, err)\n")
		tests.WriteString("	assert.Equal(t, true, response[\"auto_generated\"])\n")
		tests.WriteString("}\n\n")
	}

	return tests.String(), nil
}

func (f *FiberGenerator) GenerateDocs(routes []APIRoute, config *FrameworkConfig) (string, error) {
	// Fiber uses the same documentation generation as Gin
	return (&GinGenerator{}).GenerateDocs(routes, config)
}

func (f *FiberGenerator) GenerateDockerfile(config *FrameworkConfig) (string, error) {
	return `# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
`, nil
}

func (f *FiberGenerator) GenerateK8sManifests(config *FrameworkConfig) (map[string]string, error) {
	return (&GinGenerator{}).GenerateK8sManifests(config)
}

// Helper functions
func toCamelCase(s string) string {
	words := strings.Split(strings.ToLower(s), "_")
	for i, word := range words {
		if i > 0 || len(word) > 0 {
			if len(word) > 0 {
				words[i] = strings.Title(word)
			}
		}
	}
	return strings.Join(words, "")
}

func formatStringSlice(slice []string) string {
	if len(slice) == 0 {
		return "[]string{}"
	}
	result := "[]string{"
	for i, s := range slice {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%q", s)
	}
	result += "}"
	return result
}

func getRouteGroup(authEnabled, routeAuthRequired bool) string {
	if authEnabled && routeAuthRequired {
		return "auth"
	}
	return "v1"
}

func writeTestFiles(outputDir, testsContent string, config *FrameworkConfig) error {
	testDir := filepath.Join(outputDir, "tests")
	if err := createDirectory(testDir); err != nil {
		return err
	}

	return writeFile(filepath.Join(testDir, "handlers_test.go"), testsContent)
}

func writeDocFiles(outputDir, docsContent string, config *FrameworkConfig) error {
	docsDir := filepath.Join(outputDir, "docs")
	if err := createDirectory(docsDir); err != nil {
		return err
	}

	return writeFile(filepath.Join(docsDir, "api.md"), docsContent)
}

func writeDockerfile(outputDir, dockerfileContent string) error {
	return writeFile(filepath.Join(outputDir, "Dockerfile"), dockerfileContent)
}

func writeK8sManifests(outputDir string, manifests map[string]string) error {
	k8sDir := filepath.Join(outputDir, "k8s")
	if err := createDirectory(k8sDir); err != nil {
		return err
	}

	for filename, content := range manifests {
		if err := writeFile(filepath.Join(k8sDir, filename), content); err != nil {
			return err
		}
	}

	return nil
}

// Global framework registry instance
var globalFrameworkRegistry *FrameworkRegistry

// GetFrameworkRegistry returns the global framework registry instance
func GetFrameworkRegistry() *FrameworkRegistry {
	if globalFrameworkRegistry == nil {
		globalFrameworkRegistry = NewFrameworkRegistry()
	}
	return globalFrameworkRegistry
}