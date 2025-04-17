package config

import (
	"io/ioutil"

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
	DB struct {
		Path string `yaml:"path"`
	} `yaml:"db"`
	SMTP struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		From     string `yaml:"from"`
		To       string `yaml:"to"`
	} `yaml:"smtp"`
	DiggerPath string `yaml:"digger_path"`
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
