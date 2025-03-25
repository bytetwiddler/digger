package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		envVars        map[string]string
		expectedConfig *Config
		expectError    bool
	}{
		{
			name: "Valid config file",
			configContent: `
log:
  level: "debug"
  gelf:
    address: "127.0.0.1:12201"
  file:
    filename: "env_test.log"
    maxsize: 10
    maxbackups: 3
    maxage: 20
`,
			expectedConfig: &Config{
				Log: struct {
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
				}{
					Level: "debug",
					Gelf: struct {
						Address string `yaml:"address"`
					}{
						Address: "127.0.0.1:12201",
					},
					File: struct {
						Filename   string `yaml:"filename"`
						MaxSize    int    `yaml:"maxsize"`
						MaxBackups int    `yaml:"maxbackups"`
						MaxAge     int    `yaml:"maxage"`
					}{
						Filename:   "env_test.log",
						MaxSize:    10,
						MaxBackups: 3,
						MaxAge:     20,
					},
				},
			},
			expectError: false,
		},
		{
			name:           "Missing config file",
			configContent:  "",
			expectedConfig: nil,
			expectError:    true,
		},
		{
			name: "Invalid YAML content",
			configContent: `
log:
  level: "debug"
  gelf:
    address: "127.0.0.1:12201"
  file:
    filename: "env_test.log"
    maxsize: "invalid"
    maxbackups: 3
    maxage: 20 
`,
			expectedConfig: nil,
			expectError:    true,
		},
		{
			name: "Environment variable overrides",
			configContent: `
log:
  level: "info"
  gelf:
    address: "127.0.0.1:12201"
  file:
    filename: "env_test.log"
    maxsize: 10
    maxbackups: 3
    maxage: 20
`,
			envVars: map[string]string{
				"LOG_LEVEL":      "debug",
				"LOG_FILENAME":   "env_test.log",
				"LOG_MAXSIZE":    "10",
				"LOG_MAXBACKUPS": "3",
				"LOG_MAXAGE":     "20",
			},
			expectedConfig: &Config{
				Log: struct {
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
				}{
					Level: "info",
					Gelf: struct {
						Address string `yaml:"address"`
					}{
						Address: "127.0.0.1:12201",
					},
					File: struct {
						Filename   string `yaml:"filename"`
						MaxSize    int    `yaml:"maxsize"`
						MaxBackups int    `yaml:"maxbackups"`
						MaxAge     int    `yaml:"maxage"`
					}{
						Filename:   "env_test.log",
						MaxSize:    10,
						MaxBackups: 3,
						MaxAge:     20,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configContent != "" {
				err := ioutil.WriteFile("config.yaml", []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatalf("failed to create temporary config file: %v", err)
				}
				defer os.Remove("config.yaml")
			}

			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			cfg, err := LoadConfig("config.yaml")
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if cfg.Log.Level != tt.expectedConfig.Log.Level {
					t.Errorf("expected log level to be '%s', got '%s'", tt.expectedConfig.Log.Level, cfg.Log.Level)
				}
				if cfg.Log.Gelf.Address != tt.expectedConfig.Log.Gelf.Address {
					t.Errorf("expected GELF address to be '%s', got '%s'", tt.expectedConfig.Log.Gelf.Address, cfg.Log.Gelf.Address)
				}
				if cfg.Log.File.Filename != tt.expectedConfig.Log.File.Filename {
					t.Errorf("expected log filename to be '%s', got '%s'", tt.expectedConfig.Log.File.Filename, cfg.Log.File.Filename)
				}
				if cfg.Log.File.MaxSize != tt.expectedConfig.Log.File.MaxSize {
					t.Errorf("expected log max size to be %d, got %d", tt.expectedConfig.Log.File.MaxSize, cfg.Log.File.MaxSize)
				}
				if cfg.Log.File.MaxBackups != tt.expectedConfig.Log.File.MaxBackups {
					t.Errorf("expected log max backups to be %d, got %d", tt.expectedConfig.Log.File.MaxBackups, cfg.Log.File.MaxBackups)
				}
				if cfg.Log.File.MaxAge != tt.expectedConfig.Log.File.MaxAge {
					t.Errorf("expected log max age to be %d, got %d", tt.expectedConfig.Log.File.MaxAge, cfg.Log.File.MaxAge)
				}
			}
		})
	}
}
