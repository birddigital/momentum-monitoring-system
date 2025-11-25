package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidationRule represents a validation rule that can be applied to data
type ValidationRule struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	Config      map[string]interface{} `json:"config"`
	Priority    int                    `json:"priority"`
	Required    bool                   `json:"required"`
	Middleware bool                   `json:"middleware"`
}

// ValidationResult represents the result of applying validation rules
type ValidationResult struct {
	Valid   bool                    `json:"valid"`
	Errors  []ValidationError       `json:"errors"`
	Fields  map[string]interface{}   `json:"fields"`
	Rules   []string                `json:"applied_rules"`
	Context map[string]interface{}   `json:"context"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Validator interface for custom validation implementations
type Validator interface {
	Validate(value interface{}, config map[string]interface{}) ValidationResult
	GetName() string
	GetType() string
}

// ValidationEngine manages and applies validation rules
type ValidationEngine struct {
	rules      map[string]*ValidationRule
	validators map[string]Validator
	config     *ValidationConfig
}

// ValidationConfig contains configuration for the validation engine
type ValidationConfig struct {
	StopOnFirstError bool     `json:"stop_on_first_error"`
	StrictMode       bool     `json:"strict_mode"`
	DefaultRules     []string `json:"default_rules"`
	CustomRulesPath  string   `json:"custom_rules_path"`
}

// NewValidationEngine creates a new validation engine instance
func NewValidationEngine(config *ValidationConfig) *ValidationEngine {
	engine := &ValidationEngine{
		rules:      make(map[string]*ValidationRule),
		validators: make(map[string]Validator),
		config:     config,
	}

	// Register built-in validators
	engine.registerBuiltinValidators()

	// Load default rules
	engine.loadDefaultRules()

	return engine
}

// registerBuiltinValidators registers all built-in validation validators
func (ve *ValidationEngine) registerBuiltinValidators() {
	ve.RegisterValidator(&RequiredValidator{})
	ve.RegisterValidator(&StringValidator{})
	ve.RegisterValidator(&NumericValidator{})
	ve.RegisterValidator(&EmailValidator{})
	ve.RegisterValidator(&URLValidator{})
	ve.RegisterValidator(&RegexValidator{})
	ve.RegisterValidator(&LengthValidator{})
	ve.RegisterValidator(&RangeValidator{})
	ve.RegisterValidator(&EnumValidator{})
	ve.RegisterValidator(&DateValidator{})
	ve.RegisterValidator(&UUIDValidator{})
	ve.RegisterValidator(&PhoneValidator{})
	ve.RegisterValidator(&PasswordValidator{})
	ve.RegisterValidator(&BusinessIdentifierValidator{})
	ve.RegisterValidator(&GeoLocationValidator{})
}

// loadDefaultRules loads default validation rules
func (ve *ValidationEngine) loadDefaultRules() {
	defaultRules := []*ValidationRule{
		{
			Name:     "required",
			Type:     "field",
			Message:  "This field is required",
			Required: true,
			Priority: 100,
		},
		{
			Name:     "email",
			Type:     "field",
			Message:  "Must be a valid email address",
			Priority: 90,
			Config: map[string]interface{}{
				"allow_display_name": true,
			},
		},
		{
			Name:     "password_strength",
			Type:     "field",
			Message:  "Password must meet security requirements",
			Priority: 95,
			Config: map[string]interface{}{
				"min_length":     8,
				"require_upper":  true,
				"require_lower":  true,
				"require_number": true,
				"require_symbol": true,
			},
		},
		{
			Name:     "api_key_format",
			Type:     "field",
			Message:  "API key must follow required format",
			Priority: 85,
			Config: map[string]interface{}{
				"pattern": `^[a-zA-Z0-9]{32,}$`,
			},
		},
		{
			Name:     "pagination_limit",
			Type:     "field",
			Message:  "Pagination limit exceeds maximum allowed",
			Priority: 80,
			Config: map[string]interface{}{
				"max": 100,
			},
		},
		{
			Name:     "rate_limit_check",
			Type:     "middleware",
			Message:  "Rate limit exceeded",
			Priority: 70,
			Config: map[string]interface{}{
				"requests_per_minute": 60,
				"burst_size":          10,
			},
		},
		{
			Name:     "cors_validation",
			Type:     "middleware",
			Message:  "CORS policy violation",
			Priority: 60,
		},
		{
			Name:     "jwt_token_validation",
			Type:     "middleware",
			Message:  "Invalid or expired JWT token",
			Priority: 90,
		},
		{
			Name:     "sql_injection_check",
			Type:     "middleware",
			Message:  "Potential SQL injection detected",
			Priority: 95,
		},
		{
			Name:     "xss_prevention",
			Type:     "middleware",
			Message:  "Potential XSS attack detected",
			Priority: 90,
		},
	}

	for _, rule := range defaultRules {
		ve.AddRule(rule)
	}
}

// RegisterValidator adds a custom validator to the engine
func (ve *ValidationEngine) RegisterValidator(validator Validator) {
	ve.validators[validator.GetName()] = validator
}

// AddRule adds a validation rule to the engine
func (ve *ValidationEngine) AddRule(rule *ValidationRule) {
	ve.rules[rule.Name] = rule
}

// ValidateField validates a single field against specified rules
func (ve *ValidationEngine) ValidateField(fieldName string, value interface{}, rules []string) ValidationResult {
	result := ValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Fields:  make(map[string]interface{}),
		Rules:   []string{},
		Context: make(map[string]interface{}),
	}

	result.Fields[fieldName] = value

	for _, ruleName := range rules {
		rule, exists := ve.rules[ruleName]
		if !exists {
			continue
		}

		validator, validatorExists := ve.validators[ruleName]
		if !validatorExists {
			continue
		}

		ruleResult := validator.Validate(value, rule.Config)
		result.Rules = append(result.Rules, ruleName)

		if !ruleResult.Valid {
			result.Valid = false
			for _, err := range ruleResult.Errors {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fieldName,
					Rule:    ruleName,
					Value:   fmt.Sprintf("%v", value),
					Message: rule.Message,
					Code:    err.Code,
				})
			}

			if ve.config.StopOnFirstError {
				break
			}
		}
	}

	return result
}

// ValidateMiddleware applies middleware validation rules
func (ve *ValidationEngine) ValidateMiddleware(context map[string]interface{}) ValidationResult {
	result := ValidationResult{
		Valid:   true,
		Errors:  []ValidationError{},
		Fields:  make(map[string]interface{}),
		Rules:   []string{},
		Context: context,
	}

	for _, rule := range ve.rules {
		if !rule.Middleware {
			continue
		}

		validator, exists := ve.validators[rule.Name]
		if !exists {
			continue
		}

		ruleResult := validator.Validate(context, rule.Config)
		result.Rules = append(result.Rules, rule.Name)

		if !ruleResult.Valid {
			result.Valid = false
			for _, err := range ruleResult.Errors {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "middleware",
					Rule:    rule.Name,
					Value:   "context",
					Message: rule.Message,
					Code:    err.Code,
				})
			}

			if ve.config.StopOnFirstError {
				break
			}
		}
	}

	return result
}

// GetRulesForFramework returns validation rules optimized for specific frameworks
func (ve *ValidationEngine) GetRulesForFramework(framework string) map[string][]string {
	frameworkRules := map[string]map[string][]string{
		"gin": {
			"user_registration": {"required", "email", "password_strength"},
			"api_key_auth":      {"required", "api_key_format"},
			"pagination":        {"required", "numeric", "range"},
		},
		"echo": {
			"user_registration": {"required", "email", "password_strength"},
			"api_key_auth":      {"required", "api_key_format"},
			"pagination":        {"required", "numeric", "range"},
		},
		"chi": {
			"user_registration": {"required", "email", "password_strength"},
			"api_key_auth":      {"required", "api_key_format"},
			"pagination":        {"required", "numeric", "range"},
		},
		"fiber": {
			"user_registration": {"required", "email", "password_strength"},
			"api_key_auth":      {"required", "api_key_format"},
			"pagination":        {"required", "numeric", "range"},
		},
	}

	return frameworkRules[framework]
}

// GenerateValidationCode generates validation code for different frameworks
func (ve *ValidationEngine) GenerateValidationCode(framework string, validationType string, rules []string) string {
	templates := map[string]map[string]string{
		"gin": {
			"middleware": `
// Gin middleware for validation
func ValidationMiddleware(validator *ValidationEngine, endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rules := validator.GetRulesForFramework("gin")[endpoint]
		context := map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"headers": c.Request.Header,
		}

		result := validator.ValidateMiddleware(context)
		if !result.Valid {
			c.JSON(400, gin.H{
				"error": "Validation failed",
				"errors": result.Errors,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}`,
			"field": `
// Field validation helper for Gin
func ValidateField(c *gin.Context, fieldName string, rules []string) interface{} {
	value := c.PostForm(fieldName)
	if value == "" {
		value = c.Query(fieldName)
	}

	validator := GetValidationEngine()
	result := validator.ValidateField(fieldName, value, rules)

	if !result.Valid {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("Field %s validation failed", fieldName),
			"errors": result.Errors,
		})
		return nil
	}

	return value
}`,
		},
		"echo": {
			"middleware": `
// Echo middleware for validation
func ValidationMiddleware(validator *ValidationEngine, endpoint string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rules := validator.GetRulesForFramework("echo")[endpoint]
			context := map[string]interface{}{
				"method": c.Request().Method,
				"path":   c.Request().URL.Path,
				"headers": c.Request().Header,
			}

			result := validator.ValidateMiddleware(context)
			if !result.Valid {
				return c.JSON(400, map[string]interface{}{
					"error": "Validation failed",
					"errors": result.Errors,
				})
			}

			return next(c)
		}
	}
}`,
		},
	}

	if frameworkTemplates, exists := templates[framework]; exists {
		if template, exists := frameworkTemplates[validationType]; exists {
			return template
		}
	}

	return fmt.Sprintf("// No template available for %s %s\n", framework, validationType)
}

// Built-in Validators

type RequiredValidator struct{}

func (v *RequiredValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	if value == nil || value == "" {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "REQUIRED_MISSING", Message: "Field is required"},
		}
	}

	return result
}

func (v *RequiredValidator) GetName() string { return "required" }
func (v *RequiredValidator) GetType() string { return "field" }

type StringValidator struct{}

func (v *StringValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	str, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	if minLen, ok := config["min_length"].(int); ok && len(str) < minLen {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    "MIN_LENGTH",
			Message: fmt.Sprintf("Must be at least %d characters", minLen),
		})
	}

	if maxLen, ok := config["max_length"].(int); ok && len(str) > maxLen {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    "MAX_LENGTH",
			Message: fmt.Sprintf("Must be at most %d characters", maxLen),
		})
	}

	return result
}

func (v *StringValidator) GetName() string { return "string" }
func (v *StringValidator) GetType() string { return "field" }

type EmailValidator struct{}

func (v *EmailValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	email, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_EMAIL", Message: "Must be a valid email address"},
		}
	}

	return result
}

func (v *EmailValidator) GetName() string { return "email" }
func (v *EmailValidator) GetType() string { return "field" }

type NumericValidator struct{}

func (v *NumericValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	var num float64
	var err error

	switch val := value.(type) {
	case float64:
		num = val
	case int:
		num = float64(val)
	case string:
		num, err = strconv.ParseFloat(val, 64)
		if err != nil {
			result.Valid = false
			result.Errors = []ValidationError{
				{Code: "INVALID_NUMBER", Message: "Must be a valid number"},
			}
			return result
		}
	default:
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be numeric"},
		}
		return result
	}

	if min, ok := config["min"].(float64); ok && num < min {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    "MIN_VALUE",
			Message: fmt.Sprintf("Must be at least %.2f", min),
		})
	}

	if max, ok := config["max"].(float64); ok && num > max {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    "MAX_VALUE",
			Message: fmt.Sprintf("Must be at most %.2f", max),
		})
	}

	return result
}

func (v *NumericValidator) GetName() string { return "numeric" }
func (v *NumericValidator) GetType() string { return "field" }

type RangeValidator struct{}

func (v *RangeValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	validator := &NumericValidator{}
	return validator.Validate(value, config)
}

func (v *RangeValidator) GetName() string { return "range" }
func (v *RangeValidator) GetType() string { return "field" }

type LengthValidator struct{}

func (v *LengthValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	validator := &StringValidator{}
	return validator.Validate(value, config)
}

func (v *LengthValidator) GetName() string { return "length" }
func (v *LengthValidator) GetType() string { return "field" }

type RegexValidator struct{}

func (v *RegexValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	str, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	pattern, ok := config["pattern"].(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_CONFIG", Message: "Pattern configuration required"},
		}
		return result
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_PATTERN", Message: "Invalid regex pattern"},
		}
		return result
	}

	if !regex.MatchString(str) {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "PATTERN_MISMATCH", Message: "Does not match required pattern"},
		}
	}

	return result
}

func (v *RegexValidator) GetName() string { return "regex" }
func (v *RegexValidator) GetType() string { return "field" }

type EnumValidator struct{}

func (v *EnumValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	allowedValues, ok := config["values"].([]interface{})
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_CONFIG", Message: "Values configuration required"},
		}
		return result
	}

	for _, allowed := range allowedValues {
		if fmt.Sprintf("%v", allowed) == fmt.Sprintf("%v", value) {
			return result
		}
	}

	result.Valid = false
	result.Errors = []ValidationError{
		{Code: "INVALID_ENUM", Message: "Value not in allowed list"},
	}

	return result
}

func (v *EnumValidator) GetName() string { return "enum" }
func (v *EnumValidator) GetType() string { return "field" }

type DateValidator struct{}

func (v *DateValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	_, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	return result
}

func (v *DateValidator) GetName() string { return "date" }
func (v *DateValidator) GetType() string { return "field" }

type UUIDValidator struct{}

func (v *UUIDValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	uuidStr, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(uuidStr) {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_UUID", Message: "Must be a valid UUID"},
		}
	}

	return result
}

func (v *UUIDValidator) GetName() string { return "uuid" }
func (v *UUIDValidator) GetType() string { return "field" }

type URLValidator struct{}

func (v *URLValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	urlStr, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_URL", Message: "Must be a valid URL"},
		}
	}

	return result
}

func (v *URLValidator) GetName() string { return "url" }
func (v *URLValidator) GetType() string { return "field" }

type PhoneValidator struct{}

func (v *PhoneValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	phoneStr, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	phoneRegex := regexp.MustCompile(`^\+?[\d\s\-\(\)]{10,}$`)
	if !phoneRegex.MatchString(phoneStr) {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_PHONE", Message: "Must be a valid phone number"},
		}
	}

	return result
}

func (v *PhoneValidator) GetName() string { return "phone" }
func (v *PhoneValidator) GetType() string { return "field" }

type PasswordValidator struct{}

func (v *PasswordValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	password, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	if minLength, ok := config["min_length"].(int); ok && len(password) < minLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    "PASSWORD_TOO_SHORT",
			Message: fmt.Sprintf("Password must be at least %d characters", minLength),
		})
	}

	if requireUpper, ok := config["require_upper"].(bool); ok && requireUpper {
		if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    "PASSWORD_MISSING_UPPER",
				Message: "Password must contain uppercase letter",
			})
		}
	}

	if requireLower, ok := config["require_lower"].(bool); ok && requireLower {
		if !regexp.MustCompile(`[a-z]`).MatchString(password) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    "PASSWORD_MISSING_LOWER",
				Message: "Password must contain lowercase letter",
			})
		}
	}

	if requireNumber, ok := config["require_number"].(bool); ok && requireNumber {
		if !regexp.MustCompile(`\d`).MatchString(password) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    "PASSWORD_MISSING_NUMBER",
				Message: "Password must contain number",
			})
		}
	}

	if requireSymbol, ok := config["require_symbol"].(bool); ok && requireSymbol {
		if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    "PASSWORD_MISSING_SYMBOL",
				Message: "Password must contain special character",
			})
		}
	}

	return result
}

func (v *PasswordValidator) GetName() string { return "password_strength" }
func (v *PasswordValidator) GetType() string { return "field" }

type BusinessIdentifierValidator struct{}

func (v *BusinessIdentifierValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	str, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	if pattern, ok := config["pattern"].(string); ok {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			result.Valid = false
			result.Errors = []ValidationError{
				{Code: "INVALID_PATTERN", Message: "Invalid regex pattern"},
			}
			return result
		}

		if !regex.MatchString(str) {
			result.Valid = false
			result.Errors = []ValidationError{
				{Code: "PATTERN_MISMATCH", Message: "Does not match required format"},
			}
		}
	}

	return result
}

func (v *BusinessIdentifierValidator) GetName() string { return "api_key_format" }
func (v *BusinessIdentifierValidator) GetType() string { return "field" }

type GeoLocationValidator struct{}

func (v *GeoLocationValidator) Validate(value interface{}, config map[string]interface{}) ValidationResult {
	result := ValidationResult{Valid: true}

	coordsStr, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_TYPE", Message: "Must be a string"},
		}
		return result
	}

	parts := strings.Split(coordsStr, ",")
	if len(parts) != 2 {
		result.Valid = false
		result.Errors = []ValidationError{
			{Code: "INVALID_COORDINATES", Message: "Must be in format 'latitude,longitude'"},
		}
		return result
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil || lat < -90 || lat > 90 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    "INVALID_LATITUDE",
			Message: "Invalid latitude value",
		})
	}

	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil || lon < -180 || lon > 180 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    "INVALID_LONGITUDE",
			Message: "Invalid longitude value",
		})
	}

	return result
}

func (v *GeoLocationValidator) GetName() string { return "geo_location" }
func (v *GeoLocationValidator) GetType() string { return "field" }

// Global validation engine instance
var globalValidationEngine *ValidationEngine

// GetValidationEngine returns the global validation engine instance
func GetValidationEngine() *ValidationEngine {
	if globalValidationEngine == nil {
		config := &ValidationConfig{
			StopOnFirstError: false,
			StrictMode:       true,
			DefaultRules:     []string{"required", "string", "email"},
		}
		globalValidationEngine = NewValidationEngine(config)
	}
	return globalValidationEngine
}