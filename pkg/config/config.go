package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type LogConfig struct {
	Level   string        `yaml:"level"`
	GELF    GELFConfig    `yaml:"gelf"`
	LogFile LogFileConfig `yaml:"file"`
}

type GELFConfig struct {
	Address string `yaml:"address"`
}

type LogFileConfig struct {
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxsize"`
	MaxBackups int    `yaml:"maxbackups"`
	MaxAge     int    `yaml:"maxage"`
}

type Config struct {
	Log LogConfig `yaml:"log"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("could not open config.yaml file for reading: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var cfg Config
	decoder := yaml.NewDecoder(file)

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("could not decode config.yaml: %w", err)
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Log.Level = logLevel
	}

	if logFilename := os.Getenv("LOG_FILENAME"); logFilename != "" {
		cfg.Log.LogFile.Filename = logFilename
	}

	if logMaxSize := os.Getenv("LOG_MAXSIZE"); logMaxSize != "" {
		cfg.Log.LogFile.MaxSize = atoi(logMaxSize)
	}

	if logMaxBackups := os.Getenv("LOG_MAXBACKUPS"); logMaxBackups != "" {
		cfg.Log.LogFile.MaxBackups = atoi(logMaxBackups)
	}

	if logMaxAge := os.Getenv("LOG_MAXAGE"); logMaxAge != "" {
		cfg.Log.LogFile.MaxAge = atoi(logMaxAge)
	}

	return &cfg, nil
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
