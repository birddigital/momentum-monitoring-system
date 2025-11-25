package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// PackageInfo represents analyzed Go package information
type PackageInfo struct {
	Name         string        `json:"name"`
	ImportPath   string        `json:"import_path"`
	Structs      []StructInfo  `json:"structs"`
	Interfaces   []MethodInfo  `json:"interfaces"`
	Functions    []MethodInfo  `json:"functions"`
	Imports      []string      `json:"imports"`
}

// StructInfo represents analyzed struct information
type StructInfo struct {
	Name        string       `json:"name"`
	Fields      []FieldInfo  `json:"fields"`
	Methods     []MethodInfo `json:"methods"`
	Annotations []Annotation `json:"annotations"`
	Doc         string       `json:"doc"`
}

// FieldInfo represents struct field information
type FieldInfo struct {
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Tags        []TagInfo    `json:"tags"`
	Annotations []Annotation `json:"annotations"`
}

// MethodInfo represents method/function information
type MethodInfo struct {
	Name        string       `json:"name"`
	Receiver    string       `json:"receiver,omitempty"`
	Parameters  []Parameter  `json:"parameters"`
	Returns     []Parameter  `json:"returns,omitempty"`
	Annotations []Annotation `json:"annotations"`
	Doc         string       `json:"doc"`
}

// Parameter represents function parameter or return value
type Parameter struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
}

// TagInfo represents struct field tag information
type TagInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Annotation represents API generation annotation
type Annotation struct {
	Type   string                 `json:"type"`
	Key    string                 `json:"key"`
	Value  string                 `json:"value"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// APIGenerator represents the main scanner and generator
type APIGenerator struct {
	fset    *token.FileSet
	pkgs    map[string]*PackageInfo
	config  *GeneratorConfig
}

// GeneratorConfig contains configuration for API generation
type GeneratorConfig struct {
	IncludePatterns []string `json:"include_patterns"`
	ExcludePatterns []string `json:"exclude_patterns"`
	ScanAnnotations bool     `json:"scan_annotations"`
	AutoCRUD        bool     `json:"auto_crud"`
	SmartMapping    bool     `json:"smart_mapping"`
	OutputDir       string   `json:"output_dir"`
	PackageName     string   `json:"package_name"`
}

// NewAPIGenerator creates a new API generator instance
func NewAPIGenerator(config *GeneratorConfig) *APIGenerator {
	return &APIGenerator{
		fset:   token.NewFileSet(),
		pkgs:   make(map[string]*PackageInfo),
		config: config,
	}
}

// ScanDirectory scans a directory for Go packages
func (ag *APIGenerator) ScanDirectory(root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that should be excluded
		for _, pattern := range ag.config.ExcludePatterns {
			if matched, _ := filepath.Match(pattern, path); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Check if file matches include patterns
		included := len(ag.config.IncludePatterns) == 0
		for _, pattern := range ag.config.IncludePatterns {
			if matched, _ := filepath.Match(pattern, path); matched {
				included = true
				break
			}
		}

		if !included {
			return nil
		}

		// Parse the file
		return ag.scanFile(path)
	})

	// Post-processing: associate all methods with their structs
	if err == nil {
		ag.associateMethodsWithStructs()
	}

	return err
}

// associateMethodsWithStructs ensures all methods are properly associated with their structs
func (ag *APIGenerator) associateMethodsWithStructs() {
	for _, pkg := range ag.pkgs {
		// Process all functions and associate methods with structs
		for _, funcInfo := range pkg.Functions {
			if funcInfo.Receiver != "" {
				receiverName := strings.TrimPrefix(funcInfo.Receiver, "*")

				// Find the struct and add the method if not already present
				for i := range pkg.Structs {
					if pkg.Structs[i].Name == receiverName {
						// Check if method already exists
						found := false
						for _, existingMethod := range pkg.Structs[i].Methods {
							if existingMethod.Name == funcInfo.Name {
								found = true
								break
							}
						}
						if !found {
							pkg.Structs[i].Methods = append(pkg.Structs[i].Methods, funcInfo)
						}
						break
					}
				}
			}
		}
	}
}

// scanFile scans a single Go file
func (ag *APIGenerator) scanFile(filePath string) error {
	node, err := parser.ParseFile(ag.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Error parsing file %s: %v", filePath, err)
		return nil
	}

	pkgInfo := &PackageInfo{
		Name: node.Name.Name,
	}

	// Scan imports
	if node.Imports != nil {
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			pkgInfo.Imports = append(pkgInfo.Imports, importPath)
		}
	}

	// Scan declarations
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				ag.scanTypeDeclaration(x, pkgInfo)
			}
		case *ast.FuncDecl:
			ag.scanFunction(x, pkgInfo)
		}
		return true
	})

	// Store package info
	dir := filepath.Dir(filePath)
	if len(dir) > 4 && dir[:4] == "src/" {
		pkgInfo.ImportPath = filepath.ToSlash(dir[4:])
	} else {
		pkgInfo.ImportPath = filepath.ToSlash(dir)
	}
	if ag.pkgs[dir] == nil {
		ag.pkgs[dir] = pkgInfo
	} else {
		// Merge with existing package info, preserving method associations
		existingPkg := ag.pkgs[dir]

		// Add new structs
		for _, newStruct := range pkgInfo.Structs {
			// Check if struct already exists
			found := false
			for i := range existingPkg.Structs {
				if existingPkg.Structs[i].Name == newStruct.Name {
					// Merge methods into existing struct
					existingPkg.Structs[i].Methods = append(existingPkg.Structs[i].Methods, newStruct.Methods...)
					found = true
					break
				}
			}
			if !found {
				existingPkg.Structs = append(existingPkg.Structs, newStruct)
			}
		}

		// Re-process all functions to ensure methods are properly associated
		for _, funcInfo := range pkgInfo.Functions {
			existingPkg.Functions = append(existingPkg.Functions, funcInfo)

			// If this is a method, try to associate it with a struct
			if funcInfo.Receiver != "" {
				receiverName := strings.TrimPrefix(funcInfo.Receiver, "*")
				for i := range existingPkg.Structs {
					if existingPkg.Structs[i].Name == receiverName {
						existingPkg.Structs[i].Methods = append(existingPkg.Structs[i].Methods, funcInfo)
						break
					}
				}
			}
		}

		existingPkg.Imports = append(existingPkg.Imports, pkgInfo.Imports...)
	}

	return nil
}

// scanTypeDeclaration scans type declarations for structs and interfaces
func (ag *APIGenerator) scanTypeDeclaration(decl *ast.GenDecl, pkgInfo *PackageInfo) {
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		switch t := typeSpec.Type.(type) {
		case *ast.StructType:
			structInfo := ag.scanStruct(typeSpec.Name.Name, t, decl.Doc)
			pkgInfo.Structs = append(pkgInfo.Structs, structInfo)
		case *ast.InterfaceType:
			// Handle interface types
			ifaceInfo := MethodInfo{
				Name: typeSpec.Name.Name,
				Doc:  ag.getCommentText(decl.Doc),
			}
			pkgInfo.Interfaces = append(pkgInfo.Interfaces, ifaceInfo)
		}
	}
}

// scanStruct analyzes a struct definition
func (ag *APIGenerator) scanStruct(name string, structType *ast.StructType, doc *ast.CommentGroup) StructInfo {
	structInfo := StructInfo{
		Name:        name,
		Doc:         ag.getCommentText(doc),
		Annotations: ag.parseAnnotations(doc),
		Fields:      make([]FieldInfo, 0),
		Methods:     make([]MethodInfo, 0),
	}

	if structType.Fields != nil {
		for _, field := range structType.Fields.List {
			for _, fieldName := range field.Names {
				fieldInfo := FieldInfo{
					Name:        fieldName.Name,
					Type:        ag.getTypeString(field.Type),
					Tags:        ag.parseFieldTags(field.Tag),
					Annotations: ag.parseAnnotations(field.Doc),
				}
				structInfo.Fields = append(structInfo.Fields, fieldInfo)
			}
		}
	}

	return structInfo
}

// scanFunction analyzes a function declaration
func (ag *APIGenerator) scanFunction(decl *ast.FuncDecl, pkgInfo *PackageInfo) {
	methodInfo := MethodInfo{
		Name:        decl.Name.Name,
		Doc:         ag.getCommentText(decl.Doc),
		Annotations: ag.parseAnnotations(decl.Doc),
	}

	// Check if this is a method (has receiver)
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		receiver := decl.Recv.List[0]
		methodInfo.Receiver = ag.getTypeString(receiver.Type)
	}

	// Parse parameters
	if decl.Type.Params != nil {
		for _, param := range decl.Type.Params.List {
			paramType := ag.getTypeString(param.Type)
			if len(param.Names) > 0 {
				for _, name := range param.Names {
					methodInfo.Parameters = append(methodInfo.Parameters, Parameter{
						Name: name.Name,
						Type: paramType,
					})
				}
			} else {
				methodInfo.Parameters = append(methodInfo.Parameters, Parameter{
					Type: paramType,
				})
			}
		}
	}

	// Parse return values
	if decl.Type.Results != nil {
		for _, result := range decl.Type.Results.List {
			resultType := ag.getTypeString(result.Type)
			if len(result.Names) > 0 {
				for _, name := range result.Names {
					methodInfo.Returns = append(methodInfo.Returns, Parameter{
						Name: name.Name,
						Type: resultType,
					})
				}
			} else {
				methodInfo.Returns = append(methodInfo.Returns, Parameter{
					Type: resultType,
				})
			}
		}
	}

	// Add to package functions list
	pkgInfo.Functions = append(pkgInfo.Functions, methodInfo)

	// If this is a method with a receiver, add it to the struct's methods
	if methodInfo.Receiver != "" {
		// Extract struct name from receiver (e.g., "*UserService" -> "UserService")
		receiverName := strings.TrimPrefix(methodInfo.Receiver, "*")

		// Find the struct and add the method
		for i := range pkgInfo.Structs {
			if pkgInfo.Structs[i].Name == receiverName {
				pkgInfo.Structs[i].Methods = append(pkgInfo.Structs[i].Methods, methodInfo)
				break
			}
		}
	}
}

// parseAnnotations extracts API generation annotations from comments
func (ag *APIGenerator) parseComments(commentGroup *ast.CommentGroup) []Annotation {
	var annotations []Annotation
	if commentGroup == nil {
		return annotations
	}

	for _, comment := range commentGroup.List {
		text := strings.TrimSpace(comment.Text)
		// Look for @api annotations
		if strings.HasPrefix(text, "// @api.") || strings.HasPrefix(text, "/* @api.") {
			annotation := ag.parseAnnotationLine(text)
			if annotation != nil {
				annotations = append(annotations, *annotation)
			}
		}
	}

	return annotations
}

func (ag *APIGenerator) parseAnnotations(commentGroup *ast.CommentGroup) []Annotation {
	return ag.parseComments(commentGroup)
}

// parseAnnotationLine parses a single annotation line
func (ag *APIGenerator) parseAnnotationLine(line string) *Annotation {
	// Remove comment markers
	line = strings.TrimPrefix(line, "//")
	line = strings.TrimPrefix(line, "/*")
	line = strings.TrimSuffix(line, "*/")
	line = strings.TrimSpace(line)

	// Check if it's an API annotation
	if !strings.HasPrefix(line, "@api.") {
		return nil
	}

	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 2 {
		return nil
	}

	annotation := &Annotation{
		Type: "api",
		Key:  strings.TrimPrefix(parts[0], "@api."),
	}

	if len(parts) >= 2 {
		annotation.Value = parts[1]
	}

	// Parse configuration if available
	if len(parts) >= 3 {
		config := make(map[string]interface{})
		// Simple key=value parsing
		configStr := strings.Join(parts[2:], " ")
		kvPairs := strings.FieldsFunc(configStr, func(r rune) bool {
			return r == ',' || r == ' '
		})
		for _, kv := range kvPairs {
			if equalIndex := strings.Index(kv, "="); equalIndex > 0 {
				key := kv[:equalIndex]
				value := kv[equalIndex+1:]
				config[key] = value
			}
		}
		annotation.Config = config
	}

	return annotation
}

// parseFieldTags parses struct field tags
func (ag *APIGenerator) parseFieldTags(tag *ast.BasicLit) []TagInfo {
	var tags []TagInfo
	if tag == nil {
		return tags
	}

	tagStr := strings.Trim(tag.Value, "`")
	if tagStr == "" {
		return tags
	}

	// Simple tag parsing - split by space and then by :
	parts := strings.Split(tagStr, " ")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if colonIndex := strings.Index(part, ":"); colonIndex > 0 {
			tags = append(tags, TagInfo{
				Key:   part[:colonIndex],
				Value: strings.Trim(part[colonIndex+1:], `"`),
			})
		}
	}

	return tags
}

// getTypeString converts AST type expression to string
func (ag *APIGenerator) getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + ag.getTypeString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + ag.getTypeString(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", ag.getTypeString(t.Len), ag.getTypeString(t.Elt))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", ag.getTypeString(t.X), t.Sel.Name)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", ag.getTypeString(t.Key), ag.getTypeString(t.Value))
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return "func()"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// getCommentText extracts text from comment group
func (ag *APIGenerator) getCommentText(commentGroup *ast.CommentGroup) string {
	if commentGroup == nil {
		return ""
	}
	var texts []string
	for _, comment := range commentGroup.List {
		texts = append(texts, strings.TrimSpace(strings.TrimPrefix(comment.Text, "//")))
	}
	return strings.Join(texts, "\n")
}

// generateSmartRoutes generates routes using intelligent method mapping
func (ag *APIGenerator) generateSmartRoutes(pkg *PackageInfo, structInfo StructInfo) []APIRoute {
	var routes []APIRoute

	// Scan all methods in the struct and generate smart routes
	for _, method := range structInfo.Methods {
		mapping, found := ag.SmartMethodMapping(method.Name, structInfo.Name)

		if found && mapping.AutoGenerate {
			// Build parameters based on method signature and operation type
			parameters := ag.buildParametersForOperation(method, mapping.Operation)

			// Build response based on method signature and operation type
			responses := ag.buildResponsesForOperation(method, mapping.Operation)

			route := APIRoute{
				Path:      mapping.Path,
				Method:    mapping.Method,
				Struct:    structInfo.Name,
				Function:  method.Name,
				Package:   pkg.Name,
				Parameter: parameters,
				Response:  responses,
				Metadata: map[string]interface{}{
					"auto_generated":   true,
					"smart_mapping":     true,
					"operation":        mapping.Operation,
					"method_patterns":   mapping.Patterns,
					"intelligent_route": true,
				},
			}
			routes = append(routes, route)
		}
	}

	return routes
}

// buildParametersForOperation creates parameters based on operation type and method signature
func (ag *APIGenerator) buildParametersForOperation(method MethodInfo, operation string) []Parameter {
	var params []Parameter

	switch operation {
	case "get", "get_by", "exists":
		params = append(params, Parameter{Name: "id", Type: "string"})
	case "create":
		// Add request body parameter if method has parameters
		if len(method.Parameters) > 0 {
			params = append(params, method.Parameters[0])
		}
	case "update", "bulk_update", "activate", "deactivate", "archive", "restore":
		params = append(params, Parameter{Name: "id", Type: "string"})
		if len(method.Parameters) > 1 {
			params = append(params, method.Parameters[1])
		}
	case "delete", "bulk_delete", "unassign":
		params = append(params, Parameter{Name: "id", Type: "string"})
	case "search", "count", "list":
		// Query parameters for search/filter operations
		params = append(params, Parameter{Name: "q", Type: "string"})
		params = append(params, Parameter{Name: "limit", Type: "int"})
		params = append(params, Parameter{Name: "offset", Type: "int"})
	case "assign":
		params = append(params, Parameter{Name: "id", Type: "string"})
		if len(method.Parameters) > 0 {
			params = append(params, method.Parameters[0])
		}
	}

	return params
}

// buildResponsesForOperation creates responses based on operation type and method signature
func (ag *APIGenerator) buildResponsesForOperation(method MethodInfo, operation string) []Parameter {
	var responses []Parameter

	switch operation {
	case "get", "get_by", "find":
		if len(method.Returns) > 0 {
			responses = append(responses, method.Returns[0])
		}
	case "list", "search", "find_all":
		// Return array of the struct type
		if len(method.Returns) > 0 {
			arrayParam := method.Returns[0]
			arrayParam.Type = "[]" + arrayParam.Type
			responses = append(responses, arrayParam)
		}
	case "create", "update", "bulk_update":
		if len(method.Returns) > 0 {
			responses = append(responses, method.Returns[0])
		}
	case "delete", "exists":
		responses = append(responses, Parameter{Type: "bool"})
	case "count":
		responses = append(responses, Parameter{Type: "int"})
	case "activate", "deactivate", "archive", "restore", "assign", "unassign":
		if len(method.Returns) > 0 {
			responses = append(responses, method.Returns[0])
		} else {
			responses = append(responses, Parameter{Type: "bool"})
		}
	}

	return responses
}

// GenerateAPIRoutes generates API routes from scanned packages
func (ag *APIGenerator) GenerateAPIRoutes() []APIRoute {
	var routes []APIRoute

	for _, pkg := range ag.pkgs {
		for _, structInfo := range pkg.Structs {
			// Check for API annotations on the struct
			for _, annotation := range structInfo.Annotations {
				if annotation.Key == "route" {
					route := APIRoute{
						Path:     annotation.Value,
						Struct:   structInfo.Name,
						Package:  pkg.Name,
						Methods:  ag.extractMethodsFromConfig(annotation.Config),
						Auth:     ag.extractAuthConfig(annotation.Config),
						Metadata: annotation.Config,
					}
					routes = append(routes, route)
				}
			}

			// Auto-generate smart routes if enabled
			if ag.config.SmartMapping {
				smartRoutes := ag.generateSmartRoutes(pkg, structInfo)
				routes = append(routes, smartRoutes...)
			}

			// Auto-generate basic CRUD routes if enabled
			if ag.config.AutoCRUD {
				crudRoutes := ag.generateCRUDRoutes(pkg, structInfo)
				routes = append(routes, crudRoutes...)
			}
		}

		// Check for function-level annotations
		for _, funcInfo := range pkg.Functions {
			for _, annotation := range funcInfo.Annotations {
				if annotation.Key == "endpoint" {
					route := APIRoute{
						Path:      annotation.Value,
						Function:  funcInfo.Name,
						Package:   pkg.Name,
						Method:    ag.extractMethodFromConfig(annotation.Config),
						Auth:      ag.extractAuthConfig(annotation.Config),
						Parameter: ag.extractParameterInfo(funcInfo),
						Response:  ag.extractResponseInfo(funcInfo),
						Metadata:  annotation.Config,
					}
					routes = append(routes, route)
				}
			}
		}
	}

	return routes
}

// APIRoute represents a generated API route
type APIRoute struct {
	Path      string            `json:"path"`
	Method    string            `json:"method"`
	Struct    string            `json:"struct,omitempty"`
	Function  string            `json:"function,omitempty"`
	Package   string            `json:"package"`
	Methods   []string          `json:"methods,omitempty"`
	Auth      AuthConfig        `json:"auth"`
	Parameter []Parameter       `json:"parameter,omitempty"`
	Response  []Parameter       `json:"response,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Required bool   `json:"required"`
	Type     string `json:"type"`
	JWT      JWTConfig `json:"jwt,omitempty"`
}

type JWTConfig struct {
	Secret string `json:"secret"`
}

// Helper methods for extracting configuration from annotations
func (ag *APIGenerator) extractMethodsFromConfig(config map[string]interface{}) []string {
	if methods, ok := config["methods"]; ok {
		if methodsStr, ok := methods.(string); ok {
			return strings.Split(methodsStr, ",")
		}
	}
	return []string{"GET"} // Default to GET
}

func (ag *APIGenerator) extractAuthConfig(config map[string]interface{}) AuthConfig {
	auth := AuthConfig{Required: false}

	if required, ok := config["auth"]; ok {
		if reqStr, ok := required.(string); ok && reqStr == "required" {
			auth.Required = true
		}
	}

	if authType, ok := config["auth"]; ok {
		if typeStr, ok := authType.(string); ok {
			auth.Type = typeStr
		}
	}

	return auth
}

func (ag *APIGenerator) extractMethodFromConfig(config map[string]interface{}) string {
	if method, ok := config["method"]; ok {
		if methodStr, ok := method.(string); ok {
			return methodStr
		}
	}
	return "GET" // Default to GET
}

func (ag *APIGenerator) extractParameterInfo(funcInfo MethodInfo) []Parameter {
	return funcInfo.Parameters
}

func (ag *APIGenerator) extractResponseInfo(funcInfo MethodInfo) []Parameter {
	return funcInfo.Returns
}

// MethodMapping represents intelligent method-to-HTTP mappings
type MethodMapping struct {
	Patterns   []string `json:"patterns"`
	Method     string   `json:"method"`
	Path       string   `json:"path"`
	Operation  string   `json:"operation"`
	AutoGenerate bool   `json:"auto_generate"`
}

// SmartMethodMappings contains intelligent method mapping rules
var SmartMethodMappings = []MethodMapping{
	// CRUD operations
	{Patterns: []string{"Get*", "Find*"}, Method: "GET", Path: "/{resource}/{id}", Operation: "get", AutoGenerate: true},
	{Patterns: []string{"List*", "GetAll*", "FindAll*", "Query*"}, Method: "GET", Path: "/{resource}", Operation: "list", AutoGenerate: true},
	{Patterns: []string{"Create*", "Add*", "New*", "Insert*"}, Method: "POST", Path: "/{resource}", Operation: "create", AutoGenerate: true},
	{Patterns: []string{"Update*", "Modify*", "Edit*", "Change*"}, Method: "PUT", Path: "/{resource}/{id}", Operation: "update", AutoGenerate: true},
	{Patterns: []string{"Delete*", "Remove*", "Destroy*"}, Method: "DELETE", Path: "/{resource}/{id}", Operation: "delete", AutoGenerate: true},

	// Search and filter operations
	{Patterns: []string{"Search*", "Query*", "Filter*"}, Method: "GET", Path: "/{resource}/search", Operation: "search", AutoGenerate: true},
	{Patterns: []string{"Count*", "Total*"}, Method: "GET", Path: "/{resource}/count", Operation: "count", AutoGenerate: true},
	{Patterns: []string{"Exists*", "Check*"}, Method: "GET", Path: "/{resource}/{id}/exists", Operation: "exists", AutoGenerate: true},

	// Bulk operations
	{Patterns: []string{"BulkCreate*", "BatchCreate*", "CreateMultiple*"}, Method: "POST", Path: "/{resource}/bulk", Operation: "bulk_create", AutoGenerate: true},
	{Patterns: []string{"BulkUpdate*", "BatchUpdate*", "UpdateMultiple*"}, Method: "PUT", Path: "/{resource}/bulk", Operation: "bulk_update", AutoGenerate: true},
	{Patterns: []string{"BulkDelete*", "BatchDelete*", "DeleteMultiple*"}, Method: "DELETE", Path: "/{resource}/bulk", Operation: "bulk_delete", AutoGenerate: true},

	// Status and state operations
	{Patterns: []string{"Activate*", "Enable*"}, Method: "PUT", Path: "/{resource}/{id}/activate", Operation: "activate", AutoGenerate: true},
	{Patterns: []string{"Deactivate*", "Disable*"}, Method: "PUT", Path: "/{resource}/{id}/deactivate", Operation: "deactivate", AutoGenerate: true},
	{Patterns: []string{"Archive*"}, Method: "PUT", Path: "/{resource}/{id}/archive", Operation: "archive", AutoGenerate: true},
	{Patterns: []string{"Restore*", "Unarchive*"}, Method: "PUT", Path: "/{resource}/{id}/restore", Operation: "restore", AutoGenerate: true},

	// Relationship operations
	{Patterns: []string{"Get*By*", "Find*By*"}, Method: "GET", Path: "/{resource}/by/{field}", Operation: "get_by", AutoGenerate: true},
	{Patterns: []string{"Assign*", "Link*"}, Method: "POST", Path: "/{resource}/{id}/assign", Operation: "assign", AutoGenerate: true},
	{Patterns: []string{"Unassign*", "Unlink*"}, Method: "DELETE", Path: "/{resource}/{id}/assign", Operation: "unassign", AutoGenerate: true},
}

// SmartMethodMapping intelligently maps method names to HTTP routes
func (ag *APIGenerator) SmartMethodMapping(methodName string, structName string) (MethodMapping, bool) {
	methodLower := strings.ToLower(methodName)

	for _, mapping := range SmartMethodMappings {
		for _, pattern := range mapping.Patterns {
			patternLower := strings.ToLower(pattern)
			if ag.matchPattern(methodLower, patternLower) {
				// Customize path for this method
				customPath := ag.buildCustomPath(mapping.Path, methodName, structName)
				mapping.Path = customPath
				return mapping, true
			}
		}
	}

	// Default mapping for unrecognized methods
	return MethodMapping{
		Patterns:     []string{methodName},
		Method:       "POST",
		Path:         "/" + strings.ToLower(structName) + "/" + strings.ToLower(methodName),
		Operation:    "custom",
		AutoGenerate: false,
	}, false
}

// matchPattern checks if a method name matches a pattern
func (ag *APIGenerator) matchPattern(methodName, pattern string) bool {
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(methodName, prefix)
	}

	if strings.Contains(pattern, "*") {
		// Handle more complex patterns like "Get*By*"
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(methodName, parts[0]) && strings.HasSuffix(methodName, parts[1])
		}
	}

	return methodName == pattern
}

// buildCustomPath creates a custom path based on method name and pattern
func (ag *APIGenerator) buildCustomPath(basePath, methodName, structName string) string {
	path := basePath

	// Replace {resource} with actual resource name
	resourceName := strings.ToLower(structName) + "s"
	path = strings.ReplaceAll(path, "{resource}", resourceName)

	// Handle special patterns
	if strings.Contains(methodName, "by") && strings.Contains(path, "/{field}") {
		// Extract field name from "GetUserByEmail" -> "email"
		parts := strings.Split(strings.ToLower(methodName), "by")
		if len(parts) > 1 {
			fieldName := parts[1]
			path = strings.ReplaceAll(path, "/{field}", "/by/"+fieldName)
		}
	}

	return path
}

// generateCRUDRoutes auto-generates CRUD routes for a struct
func (ag *APIGenerator) generateCRUDRoutes(pkg *PackageInfo, structInfo StructInfo) []APIRoute {
	var routes []APIRoute
	structName := strings.ToLower(structInfo.Name)
	pluralName := structName + "s"

	// GET /{resource} - List all
	routes = append(routes, APIRoute{
		Path:     "/" + pluralName,
		Method:   "GET",
		Struct:   structInfo.Name,
		Package:  pkg.Name,
		Function: "List" + structInfo.Name,
		Metadata: map[string]interface{}{
			"auto_generated": true,
			"operation":     "list",
		},
	})

	// POST /{resource} - Create new
	routes = append(routes, APIRoute{
		Path:     "/" + pluralName,
		Method:   "POST",
		Struct:   structInfo.Name,
		Package:  pkg.Name,
		Function: "Create" + structInfo.Name,
		Parameter: []Parameter{{Type: structInfo.Name}},
		Metadata: map[string]interface{}{
			"auto_generated": true,
			"operation":     "create",
		},
	})

	// GET /{resource}/{id} - Get by ID
	routes = append(routes, APIRoute{
		Path:     "/" + pluralName + "/{id}",
		Method:   "GET",
		Struct:   structInfo.Name,
		Package:  pkg.Name,
		Function: "Get" + structInfo.Name,
		Response: []Parameter{{Type: structInfo.Name}},
		Metadata: map[string]interface{}{
			"auto_generated": true,
			"operation":     "get",
		},
	})

	// PUT /{resource}/{id} - Update
	routes = append(routes, APIRoute{
		Path:     "/" + pluralName + "/{id}",
		Method:   "PUT",
		Struct:   structInfo.Name,
		Package:  pkg.Name,
		Function: "Update" + structInfo.Name,
		Parameter: []Parameter{
			{Name: "id", Type: "string"},
			{Type: structInfo.Name},
		},
		Metadata: map[string]interface{}{
			"auto_generated": true,
			"operation":     "update",
		},
	})

	// DELETE /{resource}/{id} - Delete
	routes = append(routes, APIRoute{
		Path:     "/" + pluralName + "/{id}",
		Method:   "DELETE",
		Struct:   structInfo.Name,
		Package:  pkg.Name,
		Function: "Delete" + structInfo.Name,
		Parameter: []Parameter{{Name: "id", Type: "string"}},
		Metadata: map[string]interface{}{
			"auto_generated": true,
			"operation":     "delete",
		},
	})

	return routes
}

// GenerateAPIServer generates a complete API server from the scanned information
func (ag *APIGenerator) GenerateAPIServer() error {
	routes := ag.GenerateAPIRoutes()

	// Generate main.go
	mainTemplate := `package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Config struct {
	Port      string ` + "`json:\"port\" env:\"PORT\" default:\"8080\"`" + `
	JWTSecret string ` + "`json:\"jwt_secret\" env:\"JWT_SECRET\"`" + `
	RedisURL  string ` + "`json:\"redis_url\" env:\"REDIS_URL\"`" + `
}

type Server struct {
	config *Config
	router *gin.Engine
}

func NewServer(config *Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	server := &Server{
		config: config,
		router: gin.New(),
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())

	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		{{range .Routes}}
		v1.{{.Method|upper}}("{{.Path}}", s.{{.Function|title}})
		{{end}}
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": "auto-generated",
		"version":   "1.0.0",
	})
}

{{range .Routes}}
func (s *Server) {{.Function|title}}(c *gin.Context) {
	// Auto-generated implementation for {{.Function}}
	// TODO: Implement business logic

	{{if .Auth.Required}}
	// JWT authentication would be here
	{{end}}

	c.JSON(http.StatusOK, gin.H{
		"message": "{{.Function}} endpoint auto-generated",
		"method": "{{.Method}}",
		"path": "{{.Path}}",
		"auto_generated": true,
	})
}
{{end}}

func main() {
	config := &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		RedisURL:  getEnv("REDIS_URL", "localhost:6379"),
	}

	server := NewServer(config)
	log.Printf("Auto-generated API server starting on port %s", config.Port)

	if err := server.router.Run(":" + config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}`

	// Execute template with routes data
	// This is a simplified version - in production you'd use Go's text/template
	output := strings.Replace(mainTemplate, `{{range .Routes}}`, generateRoutesSection(routes), 1)
	output = strings.Replace(output, `{{end}}`, "", 2)

	// Write main.go
	err := os.WriteFile(filepath.Join(ag.config.OutputDir, "main.go"), []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write main.go: %v", err)
	}

	// Generate go.mod
	goModContent := `module ` + ag.config.PackageName + `

go 1.21

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/golang-jwt/jwt/v4 v4.5.2
)
`
	err = os.WriteFile(filepath.Join(ag.config.OutputDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write go.mod: %v", err)
	}

	// Generate README
	readmeContent := `# Auto-Generated API Server

## Overview
This API server was auto-generated using gofastapi scanner.

## Generated Routes
` + generateRoutesTable(routes) + `

## Usage
1. Install dependencies:
   ` + "```bash" + `
   go mod tidy
   ` + "```" + `

2. Set environment variables:
   ` + "```bash" + `
   export PORT=8080
   export JWT_SECRET=your-secret-key
   export REDIS_URL=localhost:6379
   ` + "```" + `

3. Run the server:
   ` + "```bash" + `
   go run main.go
   ` + "```" + `

## API Documentation
- Health Check: GET /health
- Generated API: GET /api/v1/...

## Notes
This is an auto-generated API. You should implement the business logic in the handler functions.
`

	err = os.WriteFile(filepath.Join(ag.config.OutputDir, "README.md"), []byte(readmeContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write README.md: %v", err)
	}

	return nil
}

// Helper functions for template generation
func generateRoutesSection(routes []APIRoute) string {
	var result string
	for _, route := range routes {
		result += fmt.Sprintf(`		v1.%s("%s", s.%s)
`, strings.ToUpper(route.Method), route.Path, route.Function)
	}
	return result
}

func generateRoutesTable(routes []APIRoute) string {
	var result string
	result += "| Method | Path | Function | Auth |\n"
	result += "|--------|------|----------|------|\n"

	for _, route := range routes {
		authStatus := "❌"
		if route.Auth.Required {
			authStatus = "✅"
		}
		result += fmt.Sprintf("| %s | %s | %s | %s |\n",
			strings.ToUpper(route.Method),
			route.Path,
			route.Function,
			authStatus)
	}

	return result
}

// PrintSummary prints a summary of scanned packages and generated routes
func (ag *APIGenerator) PrintSummary() {
	fmt.Println("\n🔍 GoFastAPI Auto-Scanner Results")
	fmt.Println("=" + strings.Repeat("=", 49))

	for pkgPath, pkg := range ag.pkgs {
		fmt.Printf("\n📦 Package: %s (%s)\n", pkg.Name, pkgPath)
		fmt.Printf("   📋 Structs: %d\n", len(pkg.Structs))
		fmt.Printf("   🔧 Functions: %d\n", len(pkg.Functions))
		fmt.Printf("   📚 Imports: %d\n", len(pkg.Imports))

		for _, structInfo := range pkg.Structs {
			fmt.Printf("      🏗️  %s (%d fields)\n", structInfo.Name, len(structInfo.Fields))
			if len(structInfo.Annotations) > 0 {
				fmt.Printf("         📝 Annotations: %d\n", len(structInfo.Annotations))
			}
		}
	}

	routes := ag.GenerateAPIRoutes()
	fmt.Printf("\n🚀 Generated API Routes: %d\n", len(routes))

	for _, route := range routes {
		auth := "Public"
		if route.Auth.Required {
			auth = "Auth Required"
		}
		fmt.Printf("   %s %s - %s (%s)\n",
			strings.ToUpper(route.Method),
			route.Path,
			route.Function,
			auth)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
}

// SaveAnalysis saves the analysis results to JSON
func (ag *APIGenerator) SaveAnalysis(filename string) error {
	analysis := map[string]interface{}{
		"packages": ag.pkgs,
		"routes":   ag.GenerateAPIRoutes(),
		"config":   ag.config,
	}

	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func main() {
	config := &GeneratorConfig{
		IncludePatterns: []string{"*.go"},
		ExcludePatterns: []string{"*_test.go", "vendor/*", ".git/*"},
		ScanAnnotations: true,
		AutoCRUD:        true,
		SmartMapping:    true,
		OutputDir:       "./generated-api",
		PackageName:     "autogenerated-api",
	}

	generator := NewAPIGenerator(config)

	// Scan current directory
	fmt.Println("🔍 Scanning Go files...")
	err := generator.ScanDirectory(".")
	if err != nil {
		log.Fatalf("Error scanning directory: %v", err)
	}

	// Print summary
	generator.PrintSummary()

	// Save analysis to JSON
	err = generator.SaveAnalysis("api-analysis.json")
	if err != nil {
		log.Printf("Warning: Failed to save analysis: %v", err)
	}

	// Create output directory
	err = os.MkdirAll(config.OutputDir, 0755)
	if err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Generate API server
	fmt.Println("\n🚀 Generating API server...")
	err = generator.GenerateAPIServer()
	if err != nil {
		log.Fatalf("Error generating API server: %v", err)
	}

	fmt.Printf("\n✅ API server generated in: %s\n", config.OutputDir)
	fmt.Println("\nNext steps:")
	fmt.Printf("   cd %s\n", config.OutputDir)
	fmt.Println("   go mod tidy")
	fmt.Println("   go run main.go")
}