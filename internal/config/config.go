package config

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelDebug = "debug"

	defaultListen            = ":10011"
	defaultURL               = "http://127.0.0.1"
	defaultLogLevel          = LogLevelInfo
	defaultWorkDir           = "/tmp/testdata"
	defaultWorkers           = 2
	defaultIndexPageFileName = "index.html"
	defaultTemplateFileName  = "template.html"
	defaultDescFileName      = "description.md"
	defaultRedisURL          = "http://127.0.0.1/0"
	defaultRedirectHeader    = "X-Accel-Redirect"
	defaultRealIPHeader      = "X-Real-IP"
	defaultDumpFilename      = "/tmp/fetchtracker_counters.json"

	envHandlerURLname = "FT_URL"
)

type IndexerConfig struct {
	WorkDir              string   `yaml:"work_dir"`
	Workers              int      `yaml:"workers"`
	IndexPageFileName    string   `yaml:"index_filename"` // If it is present in the shared folder, the page is generated only based on it. Template and markdown files are ignored.
	DescFileName         string   `yaml:"desc_filename"`
	TemplateFileName     string   `yaml:"template_filename"`
	DefaultIndexTemplate string   `yaml:"index_template"`
	DefaultMDTemplate    string   `yaml:"md_template"`
	SkipFiles            []string `yaml:"skip_files"`
	DumpFileName         string   `yaml:"dump_filename"`
}

type FSAdapterConfig struct {
	WorkDir           string
	URL               string
	IndexPageFileName string
	DescFileName      string
	TemplateFileName  string
	SkipFiles         []string
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

func (c *Config) SetDefaults() error {
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

	if c.IndexerConfig.TemplateFileName == "" {
		c.IndexerConfig.TemplateFileName = defaultTemplateFileName
	}

	if c.IndexerConfig.IndexPageFileName == "" {
		c.IndexerConfig.IndexPageFileName = defaultIndexPageFileName
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

	if c.IndexerConfig.DumpFileName == "" {
		c.IndexerConfig.DumpFileName = defaultDumpFilename
	}

	// HandlerConfig
	// Fix handler URL
	var (
		strURL string
		noPort bool
	)

	if c.HandlerConfig.URL == "" {
		strURL = defaultURL
	}
	if ftURL := os.Getenv(envHandlerURLname); ftURL != "" {
		strURL = ftURL
	}

	u, err := url.Parse(strURL)
	if err != nil {
		return fmt.Errorf("cannot parse handler url: %w", err)
	}

	if port := u.Port(); u.Scheme == "http" && port == "80" || u.Scheme == "https" && port == "443" {
		noPort = true
	}

	if noPort {
		strURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Hostname(), u.RequestURI())
	} else {
		strURL = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.RequestURI())
	}

	// To prevent double slash
	c.HandlerConfig.URL = strings.TrimSuffix(strURL, "/")

	if c.HandlerConfig.RedirectHeader == "" {
		c.HandlerConfig.RedirectHeader = defaultRedirectHeader
	}

	if c.HandlerConfig.RealIPHeader == "" {
		c.HandlerConfig.RealIPHeader = defaultRealIPHeader
	}

	return nil
}

func (c *Config) FSAdapterConfig() *FSAdapterConfig {
	return &FSAdapterConfig{
		WorkDir:           c.IndexerConfig.WorkDir,
		URL:               c.HandlerConfig.URL,
		IndexPageFileName: c.IndexerConfig.IndexPageFileName,
		DescFileName:      c.IndexerConfig.DescFileName,
		TemplateFileName:  c.IndexerConfig.TemplateFileName,
		SkipFiles:         c.IndexerConfig.SkipFiles,
	}
}

func (c *Config) Validate() error {
	panic("not implemented")
}
