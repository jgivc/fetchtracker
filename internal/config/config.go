package config

import (
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v2"
)

const (
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelDebug = "debug"

	defaultListen         = ":10011"
	defaultURL            = "http://127.0.0.1"
	defaultLogLevel       = LogLevelInfo
	defaultWorkDir        = "/tmp/testdata"
	defaultWorkers        = 2
	defaultDescFileName   = "description.md"
	defaultRedisURL       = "http://127.0.0.1/0"
	defaultRedirectHeader = "X-Accel-Redirect"
	defaultRealIPHeader   = "X-Real-IP"
)

type IndexerConfig struct {
	WorkDir          string   `yaml:"work_dir"`
	Workers          int      `yaml:"workers"`
	DescFileName     string   `yaml:"desc_filename"`
	TemplateFileName string   `yaml:"template_filename"`
	SkipFiles        []string `yaml:"skip_files"`
}

type HandlerConfig struct {
	URL            string `yaml:"url"`
	RedirectHeader string `yaml:"header_redirect"`
	RealIPHeader   string `yaml:"header_realip"`
}

type Config struct {
	Listen        string        `yaml:"listen"`
	RedisURL      string        `yaml:"redis"`
	LogLevel      string        `yaml:"log_level"`
	IndexerConfig IndexerConfig `yaml:"indexer"`
	HandlerConfig HandlerConfig `yaml:"handler"`
}

func LoadConfig(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	// expandedData := expandConfigEnvVars(data)

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set default values
	config.SetDefaults()

	// Validate config
	// if err := config.Validate(); err != nil {
	// 	return nil, fmt.Errorf("invalid configuration: %w", err)
	// }

	return &config, nil
}

func MustLoad(path string) *Config {
	config, err := LoadConfig(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}
	return config
}

func (c *Config) SetDefaults() {
	if c.Listen == "" {
		c.Listen = defaultListen
	}

	if c.RedisURL == "" {
		c.RedisURL = defaultRedisURL
	}

	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}

	// IndexerConfig
	if c.IndexerConfig.WorkDir == "" {
		c.IndexerConfig.WorkDir = defaultWorkDir
	}

	if c.IndexerConfig.Workers == 0 {
		c.IndexerConfig.Workers = defaultWorkers
	}

	if c.IndexerConfig.DescFileName == "" {
		c.IndexerConfig.DescFileName = defaultDescFileName
	}

	if len(c.IndexerConfig.SkipFiles) < 1 {
		c.IndexerConfig.SkipFiles = []string{c.IndexerConfig.DescFileName}
	} else {
		if !slices.Contains[[]string, string](c.IndexerConfig.SkipFiles, c.IndexerConfig.DescFileName) {
			c.IndexerConfig.SkipFiles = append(c.IndexerConfig.SkipFiles, c.IndexerConfig.DescFileName)
		}
	}

	// HandlerConfig
	if c.HandlerConfig.URL == "" {
		c.HandlerConfig.URL = defaultURL
	}

	if c.HandlerConfig.RedirectHeader == "" {
		c.HandlerConfig.RedirectHeader = defaultRedirectHeader
	}

	if c.HandlerConfig.RealIPHeader == "" {
		c.HandlerConfig.RealIPHeader = defaultRealIPHeader
	}
}

func (c *Config) Validate() error {
	panic("not implemented")
}
