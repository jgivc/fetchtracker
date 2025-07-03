package config

type IndexerConfig struct {
	WorkDir      string `yaml:"work_dir"`
	Workers      int    `yaml:"workers"`
	DescFileName string `yaml:"desc_filename"`
}

type Config struct {
	IndexerConfig IndexerConfig `yaml:"indexer"`
}
