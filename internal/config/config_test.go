package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantErr  bool
		expected *Config
	}{
		{
			name: "default config when no env vars set",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_NAME":     "todo_app",
				"DB_USER":     "postgres",
				"DB_PASSWORD": "postgres",
				"DB_SSLMODE":  "disable",
				"LOG_LEVEL":   "info",
			},
			wantErr: false,
			expected: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Name:     "todo_app",
					User:     "postgres",
					Password: "postgres",
					SSLMode:  "disable",
				},
				LogLevel: "info",
			},
		},
		{
			name: "custom config from env vars",
			envVars: map[string]string{
				"DB_HOST":     "customhost",
				"DB_PORT":     "5433",
				"DB_NAME":     "custom_db",
				"DB_USER":     "custom_user",
				"DB_PASSWORD": "custom_pass",
				"DB_SSLMODE":  "require",
				"LOG_LEVEL":   "debug",
			},
			wantErr: false,
			expected: &Config{
				Database: DatabaseConfig{
					Host:     "customhost",
					Port:     5433,
					Name:     "custom_db",
					User:     "custom_user",
					Password: "custom_pass",
					SSLMode:  "require",
				},
				LogLevel: "debug",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			cfg, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg.Database.Host != tt.expected.Database.Host {
					t.Errorf("Database.Host = %v, want %v", cfg.Database.Host, tt.expected.Database.Host)
				}
				if cfg.Database.Port != tt.expected.Database.Port {
					t.Errorf("Database.Port = %v, want %v", cfg.Database.Port, tt.expected.Database.Port)
				}
				if cfg.Database.Name != tt.expected.Database.Name {
					t.Errorf("Database.Name = %v, want %v", cfg.Database.Name, tt.expected.Database.Name)
				}
				if cfg.LogLevel != tt.expected.LogLevel {
					t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, tt.expected.LogLevel)
				}
			}
		})
	}
}