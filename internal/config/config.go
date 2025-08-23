package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project  ProjectConfig  `yaml:"project"`
	SSL      SSLConfig      `yaml:"ssl"`
	Nginx    NginxConfig    `yaml:"nginx"`
	Security SecurityConfig `yaml:"security"`
	Docker   DockerConfig   `yaml:"docker"`
}

type ProjectConfig struct {
	Name      string `yaml:"name"`
	Domain    string `yaml:"domain"`
	OutputDir string `yaml:"output_dir"`
}

type SSLConfig struct {
	KeySize      int    `yaml:"key_size"`
	ValidityDays int    `yaml:"validity_days"`
	Country      string `yaml:"country"`
	State        string `yaml:"state"`
	City         string `yaml:"city"`
	Organization string `yaml:"organization"`
	Unit         string `yaml:"unit"`
	Email        string `yaml:"email"`
}

type NginxConfig struct {
	HTTPSPort     int               `yaml:"https_port"`
	HTTPPort      int               `yaml:"http_port"`
	ServerName    string            `yaml:"server_name"`
	Services      []ServiceConfig   `yaml:"services"`
	CustomHeaders map[string]string `yaml:"custom_headers"`
}

type ServiceConfig struct {
	Name      string `yaml:"name"`
	Port      int    `yaml:"port"`
	Path      string `yaml:"path"`
	ProxyPass string `yaml:"proxy_pass"`
}

type SecurityConfig struct {
	Enabled       bool              `yaml:"enabled"`
	Level         string            `yaml:"level"` // strict, balanced, permissive
	EnableHSTS    bool              `yaml:"enable_hsts"`
	HSTSMaxAge    int               `yaml:"hsts_max_age"`
	EnableCSP     bool              `yaml:"enable_csp"`
	CSPPolicy     string            `yaml:"csp_policy"`
	CustomHeaders map[string]string `yaml:"custom_headers"`
	RateLimit     RateLimitConfig   `yaml:"rate_limit"`
}

type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	BurstSize         int  `yaml:"burst_size"`
}

type DockerConfig struct {
	ComposeVersion string `yaml:"compose_version"`
	NetworkName    string `yaml:"network_name"`
	NginxImage     string `yaml:"nginx_image"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Project: ProjectConfig{
			Name:      "keynginx-project",
			Domain:    "localhost",
			OutputDir: "./keynginx-output",
		},
		SSL: SSLConfig{
			KeySize:      2048,
			ValidityDays: 365,
			Country:      "US",
			State:        "CA",
			City:         "San Francisco",
			Organization: "KeyNginx Generated",
			Unit:         "IT Department",
			Email:        "",
		},
		Nginx: NginxConfig{
			HTTPSPort:  8443,
			HTTPPort:   8080,
			ServerName: "localhost",
			Services:   []ServiceConfig{},
			CustomHeaders: map[string]string{
				"X-Server-Created-By": "keynginx",
			},
		},
		Security: SecurityConfig{
			Enabled:       true,
			Level:         "balanced",
			EnableHSTS:    true,
			HSTSMaxAge:    31536000, // 1 year
			EnableCSP:     true,
			CSPPolicy:     "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:;",
			CustomHeaders: map[string]string{},
			RateLimit: RateLimitConfig{
				Enabled:           false,
				RequestsPerMinute: 100,
				BurstSize:         20,
			},
		},
		Docker: DockerConfig{
			ComposeVersion: "3.8",
			NetworkName:    "keynginx-network",
			NginxImage:     "nginx:alpine",
		},
	}
}

func (c *Config) Validate() error {
	if c.Project.Domain == "" {
		return fmt.Errorf("domain is required")
	}

	if c.SSL.KeySize < 2048 {
		return fmt.Errorf("SSL key size must be at least 2048 bits")
	}

	if c.SSL.ValidityDays <= 0 {
		return fmt.Errorf("SSL validity days must be positive")
	}

	if c.Nginx.HTTPSPort <= 0 || c.Nginx.HTTPSPort > 65535 {
		return fmt.Errorf("invalid HTTPS port: %d", c.Nginx.HTTPSPort)
	}

	if c.Nginx.HTTPPort <= 0 || c.Nginx.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.Nginx.HTTPPort)
	}

	return nil
}

func (c *Config) AddService(name string, port int, path string) {
	service := ServiceConfig{
		Name:      name,
		Port:      port,
		Path:      path,
		ProxyPass: fmt.Sprintf("http://%s:%d", name, port),
	}
	c.Nginx.Services = append(c.Nginx.Services, service)
}

func (c *Config) SetSecurityLevel(level string) error {
	validLevels := map[string]bool{
		"strict":     true,
		"balanced":   true,
		"permissive": true,
	}

	if !validLevels[level] {
		return fmt.Errorf("invalid security level: %s (must be strict, balanced, or permissive)", level)
	}

	c.Security.Level = level

	switch level {
	case "strict":
		c.Security.EnableHSTS = true
		c.Security.HSTSMaxAge = 63072000 // 2 years
		c.Security.EnableCSP = true
		c.Security.CSPPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self';"
	case "balanced":
		c.Security.EnableHSTS = true
		c.Security.HSTSMaxAge = 31536000 // 1 year
		c.Security.EnableCSP = true
		c.Security.CSPPolicy = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:;"
	case "permissive":
		c.Security.EnableHSTS = false
		c.Security.EnableCSP = false
	}

	return nil
}

func (c *Config) Save(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}
