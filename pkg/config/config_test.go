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
    filename: "test.log"
    maxsize: 5
    maxbackups: 2
    maxage: 15
`,
			expectedConfig: &Config{
				Log: LogConfig{
					Level: "debug",
					GELF: GELFConfig{
						Address: "127.0.0.1:12201",
					},
					LogFile: LogFileConfig{
						Filename:   "test.log",
						MaxSize:    5,
						MaxBackups: 2,
						MaxAge:     15,
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
    filename: "test.log"
    maxsize: "invalid"
    maxbackups: 2
    maxage: 15
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
    filename: "test.log"
    maxsize: 5
    maxbackups: 2
    maxage: 15
`,
			envVars: map[string]string{
				"LOG_LEVEL":      "debug",
				"LOG_FILENAME":   "env_test.log",
				"LOG_MAXSIZE":    "10",
				"LOG_MAXBACKUPS": "3",
				"LOG_MAXAGE":     "20",
			},
			expectedConfig: &Config{
				Log: LogConfig{
					Level: "debug",
					GELF: GELFConfig{
						Address: "127.0.0.1:12201",
					},
					LogFile: LogFileConfig{
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

			cfg, err := LoadConfig()
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
				if cfg.Log.GELF.Address != tt.expectedConfig.Log.GELF.Address {
					t.Errorf("expected GELF address to be '%s', got '%s'", tt.expectedConfig.Log.GELF.Address, cfg.Log.GELF.Address)
				}
				if cfg.Log.LogFile.Filename != tt.expectedConfig.Log.LogFile.Filename {
					t.Errorf("expected log filename to be '%s', got '%s'", tt.expectedConfig.Log.LogFile.Filename, cfg.Log.LogFile.Filename)
				}
				if cfg.Log.LogFile.MaxSize != tt.expectedConfig.Log.LogFile.MaxSize {
					t.Errorf("expected log max size to be %d, got %d", tt.expectedConfig.Log.LogFile.MaxSize, cfg.Log.LogFile.MaxSize)
				}
				if cfg.Log.LogFile.MaxBackups != tt.expectedConfig.Log.LogFile.MaxBackups {
					t.Errorf("expected log max backups to be %d, got %d", tt.expectedConfig.Log.LogFile.MaxBackups, cfg.Log.LogFile.MaxBackups)
				}
				if cfg.Log.LogFile.MaxAge != tt.expectedConfig.Log.LogFile.MaxAge {
					t.Errorf("expected log max age to be %d, got %d", tt.expectedConfig.Log.LogFile.MaxAge, cfg.Log.LogFile.MaxAge)
				}
			}
		})
	}
}
