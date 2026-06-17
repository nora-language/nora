package format

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	IndentSize      int  `yaml:"indent_size"`
	UseTabs         bool `yaml:"use_tabs"`
	MaxLineWidth    int  `yaml:"max_line_width"`
	OrganizeImports bool `yaml:"organize_imports"`
}

func DefaultConfig() *Config {
	return &Config{
		IndentSize:      4,
		UseTabs:         false,
		MaxLineWidth:    100,
		OrganizeImports: true,
	}
}

func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}
