package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Mode       string           `yaml:"mode"`
	Mock       APIConfig        `yaml:"mock"`
	Production ProductionConfig `yaml:"production"`
	Common     CommonConfig     `yaml:"common"`
}

type APIConfig struct {
	JobAPIURL           string `yaml:"job_api_url"`
	DatadogAPIURL       string `yaml:"datadog_api_url"`
	SandboxLogsAPIURL   string `yaml:"sandbox_logs_api_url"`
	SandboxSmartLogsURL string `yaml:"sandbox_smart_logs_api_url"`
}

type ProductionConfig struct {
	APIConfig `yaml:",inline"`
	Auth      AuthConfig `yaml:"auth"`
}

type AuthConfig struct {
	Datadog  AuthMethod `yaml:"datadog"`
}

type AuthMethod struct {
	Type   string `yaml:"type"`
	Token  string `yaml:"token,omitempty"`
	APIKey string `yaml:"api_key,omitempty"`
	AppKey string `yaml:"app_key,omitempty"`
}

type CommonConfig struct {
	Timeout    time.Duration `yaml:"timeout"`
	RetryCount int           `yaml:"retry_count"`
	RetryDelay time.Duration `yaml:"retry_delay"`
}

var (
	globalConfig *Config
	configPath   = "config/config.yaml"
)

// LoadConfig loads configuration from file
func LoadConfig() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	// Check for config file path override
	if envPath := os.Getenv("GENESISGPT_CONFIG"); envPath != "" {
		configPath = envPath
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		// If config file doesn't exist, use default mock configuration
		if os.IsNotExist(err) {
			return getDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Replace environment variables
	config.replaceEnvVars()

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, err
	}

	globalConfig = &config
	return globalConfig, nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	if globalConfig == nil {
		config, _ := LoadConfig()
		return config
	}
	return globalConfig
}

// IsMockMode returns true if running in mock mode
func IsMockMode() bool {
	config := GetConfig()
	return config.Mode == "mock"
}

// GetAPIConfig returns the appropriate API configuration based on mode
func GetAPIConfig() APIConfig {
	config := GetConfig()
	if config.Mode == "production" {
		return config.Production.APIConfig
	}
	return config.Mock
}

// GetAuthConfig returns authentication configuration (only for production)
func GetAuthConfig() *AuthConfig {
	config := GetConfig()
	if config.Mode == "production" {
		return &config.Production.Auth
	}
	return nil
}

// replaceEnvVars replaces ${VAR} with environment variable values
func (c *Config) replaceEnvVars() {
	if c.Mode == "production" {
		// Replace auth tokens
		if c.Production.Auth.Datadog.APIKey != "" {
			c.Production.Auth.Datadog.APIKey = expandEnv(c.Production.Auth.Datadog.APIKey)
		}
		if c.Production.Auth.Datadog.AppKey != "" {
			c.Production.Auth.Datadog.AppKey = expandEnv(c.Production.Auth.Datadog.AppKey)
		}
	}
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		envVar := s[2 : len(s)-1]
		return os.Getenv(envVar)
	}
	return s
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	// Validate mode
	if c.Mode != "mock" && c.Mode != "production" {
		return fmt.Errorf("invalid mode: %s (must be 'mock' or 'production')", c.Mode)
	}

	// In production mode, validate required auth settings
	if c.Mode == "production" {
		// Check if Datadog credentials are set (only warn, don't fail)
		if c.Production.Auth.Datadog.APIKey == "" || c.Production.Auth.Datadog.AppKey == "" {
			fmt.Println("Warning: Datadog API credentials not set. Datadog API calls will fail.")
			fmt.Println("Set DD_API_KEY and DD_APPLICATION_KEY environment variables for Datadog integration.")
		}
	}

	return nil
}

// PrintConfig prints the current configuration (for debugging)
func PrintConfig() {
	cfg := GetConfig()
	fmt.Println("=== GenesisGpt Configuration ===")
	fmt.Printf("Mode: %s\n", cfg.Mode)
	
	apiConfig := GetAPIConfig()
	fmt.Println("\nAPI Endpoints:")
	fmt.Printf("  Job API: %s\n", apiConfig.JobAPIURL)
	fmt.Printf("  Datadog API: %s\n", apiConfig.DatadogAPIURL)
	fmt.Printf("  Sandbox Logs API: %s\n", apiConfig.SandboxLogsAPIURL)
	
	if cfg.Mode == "production" {
		fmt.Println("\nAuthentication:")
		fmt.Printf("  Datadog API Key: %s\n", maskSecret(cfg.Production.Auth.Datadog.APIKey))
		fmt.Printf("  Datadog App Key: %s\n", maskSecret(cfg.Production.Auth.Datadog.AppKey))
	}
	
	fmt.Println("\nCommon Settings:")
	fmt.Printf("  Timeout: %v\n", cfg.Common.Timeout)
	fmt.Printf("  Retry Count: %d\n", cfg.Common.RetryCount)
	fmt.Printf("  Retry Delay: %v\n", cfg.Common.RetryDelay)
	fmt.Println("================================")
}

// maskSecret masks sensitive information for display
func maskSecret(secret string) string {
	if secret == "" {
		return "<not set>"
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

func getDefaultConfig() *Config {
	return &Config{
		Mode: "mock",
		Mock: APIConfig{
			JobAPIURL:           "http://localhost:8080",
			DatadogAPIURL:       "http://localhost:8080/api/datadog",
			SandboxLogsAPIURL:   "http://localhost:8080/api/sandbox",
			SandboxSmartLogsURL: "http://localhost:8080/api/sandbox",
		},
		Production: ProductionConfig{
			APIConfig: APIConfig{
				JobAPIURL:           "https://genesis.company.com/api/v1",
				DatadogAPIURL:       "https://api.datadoghq.com/api/v2/traces",
				SandboxLogsAPIURL:   "https://sandboxlogs.company.com/api/logs",
				SandboxSmartLogsURL: "https://sandboxlogs.company.com/api/logs",
			},
			Auth: AuthConfig{
				Datadog: AuthMethod{
					Type:   "api-key",
					APIKey: "",  // Must be set via environment variables
					AppKey: "",  // Must be set via environment variables
				},
			},
		},
		Common: CommonConfig{
			Timeout:    30 * time.Second,
			RetryCount: 3,
			RetryDelay: 2 * time.Second,
		},
	}
}