package config

type IndexerConfig struct {
	URL              string `yaml:"url"`
	Listen           string `yaml:"listen"`
	WorkDir          string `yaml:"work_dir"`
	Workers          int    `yaml:"workers"`
	DescFileName     string `yaml:"desc_filename"`
	TemplateFileName string `yaml:"template_filename"`
}

type Config struct {
	IndexerConfig IndexerConfig `yaml:"indexer"`
}
