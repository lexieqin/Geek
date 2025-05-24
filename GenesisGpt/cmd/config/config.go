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
	JobAPI   AuthMethod `yaml:"job_api"`
	Datadog  AuthMethod `yaml:"datadog"`
	Sandbox  AuthMethod `yaml:"sandbox"`
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
		if c.Production.Auth.JobAPI.Token != "" {
			c.Production.Auth.JobAPI.Token = expandEnv(c.Production.Auth.JobAPI.Token)
		}
		if c.Production.Auth.Datadog.APIKey != "" {
			c.Production.Auth.Datadog.APIKey = expandEnv(c.Production.Auth.Datadog.APIKey)
		}
		if c.Production.Auth.Datadog.AppKey != "" {
			c.Production.Auth.Datadog.AppKey = expandEnv(c.Production.Auth.Datadog.AppKey)
		}
		if c.Production.Auth.Sandbox.Token != "" {
			c.Production.Auth.Sandbox.Token = expandEnv(c.Production.Auth.Sandbox.Token)
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

func getDefaultConfig() *Config {
	return &Config{
		Mode: "mock",
		Mock: APIConfig{
			JobAPIURL:           "http://localhost:8080/tenant/{tenant}/jobs",
			DatadogAPIURL:       "http://localhost:8080/api/datadog/trace/{traceID}",
			SandboxLogsAPIURL:   "http://localhost:8080/api/sandbox/logs",
			SandboxSmartLogsURL: "http://localhost:8080/api/sandbox/logs/smart",
		},
		Common: CommonConfig{
			Timeout:    30 * time.Second,
			RetryCount: 3,
			RetryDelay: 2 * time.Second,
		},
	}
}