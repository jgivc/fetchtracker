package config

type IndexerConfig struct {
	WorkDir string `yaml:"work_dir"`
	Workers int    `yaml:"workers"`
}

type Config struct {
	IndexerConfig IndexerConfig `yaml:"indexer"`
}
