package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Log struct {
		Level string `yaml:"level"`
		Gelf  struct {
			Address string `yaml:"address"`
		} `yaml:"gelf"`
		File struct {
			Filename   string `yaml:"filename"`
			MaxSize    int    `yaml:"maxsize"`
			MaxBackups int    `yaml:"maxbackups"`
			MaxAge     int    `yaml:"maxage"`
		} `yaml:"file"`
	} `yaml:"log"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
