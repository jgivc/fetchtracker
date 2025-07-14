package config

const (
	RedirectHeader = "X-Accel-Redirect"
	RealIPHeader   = "X-Real-IP"
)

type IndexerConfig struct {
	WorkDir          string `yaml:"work_dir"`
	Workers          int    `yaml:"workers"`
	DescFileName     string `yaml:"desc_filename"`
	TemplateFileName string `yaml:"template_filename"`
}

type Config struct {
	URL            string        `yaml:"url"`
	Listen         string        `yaml:"listen"`
	RedirectHeader string        `yaml:"header_redirect"`
	RealIPHeader   string        `yaml:"header_realip"`
	IndexerConfig  IndexerConfig `yaml:"indexer"`
}
