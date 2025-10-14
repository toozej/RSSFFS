package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnvVars(t *testing.T) {
	tests := []struct {
		name              string
		mockEnv           map[string]string
		mockEnvFile       string
		expectError       bool
		expectRSSEndpoint string
		expectRSSAPIKey   string
		expectExitCall    bool
	}{
		{
			name: "Valid environment variables",
			mockEnv: map[string]string{
				"RSS_READER_ENDPOINT": "https://miniflux.example.com",
				"RSS_READER_API_KEY":  "test-api-key",
			},
			expectError:       false,
			expectRSSEndpoint: "https://miniflux.example.com",
			expectRSSAPIKey:   "test-api-key",
		},
		{
			name:              "Valid .env file",
			mockEnvFile:       "RSS_READER_ENDPOINT=https://miniflux.example.com\nRSS_READER_API_KEY=test-env-file-key\n",
			expectError:       false,
			expectRSSEndpoint: "https://miniflux.example.com",
			expectRSSAPIKey:   "test-env-file-key",
		},
		{
			name: "Environment variable overrides .env file",
			mockEnv: map[string]string{
				"RSS_READER_ENDPOINT": "https://env.example.com",
				"RSS_READER_API_KEY":  "env-api-key",
			},
			mockEnvFile:       "RSS_READER_ENDPOINT=https://file.example.com\nRSS_READER_API_KEY=file-api-key\n",
			expectError:       false,
			expectRSSEndpoint: "https://env.example.com",
			expectRSSAPIKey:   "env-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original directory and change to temp directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}

			// Save original environment variables
			originalEndpoint := os.Getenv("RSS_READER_ENDPOINT")
			originalAPIKey := os.Getenv("RSS_READER_API_KEY")
			defer func() {
				if originalEndpoint != "" {
					os.Setenv("RSS_READER_ENDPOINT", originalEndpoint)
				} else {
					os.Unsetenv("RSS_READER_ENDPOINT")
				}
				if originalAPIKey != "" {
					os.Setenv("RSS_READER_API_KEY", originalAPIKey)
				} else {
					os.Unsetenv("RSS_READER_API_KEY")
				}
			}()

			tmpDir := t.TempDir()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}
			defer func() {
				if err := os.Chdir(originalDir); err != nil {
					t.Errorf("Failed to restore original directory: %v", err)
				}
			}()

			// Clear environment variables first
			os.Unsetenv("RSS_READER_ENDPOINT")
			os.Unsetenv("RSS_READER_API_KEY")

			// Create .env file if applicable
			if tt.mockEnvFile != "" {
				envPath := filepath.Join(tmpDir, ".env")
				if err := os.WriteFile(envPath, []byte(tt.mockEnvFile), 0644); err != nil {
					t.Fatalf("Failed to write mock .env file: %v", err)
				}
			}

			// Set mock environment variables (these should override .env file)
			for key, value := range tt.mockEnv {
				os.Setenv(key, value)
			}

			// Call function - only test cases that shouldn't exit
			if !tt.expectExitCall {
				conf := GetEnvVars()

				// Verify output
				if conf.RSSReaderEndpoint != tt.expectRSSEndpoint {
					t.Errorf("expected RSS endpoint %q, got %q", tt.expectRSSEndpoint, conf.RSSReaderEndpoint)
				}
				if conf.RSSReaderAPIKey != tt.expectRSSAPIKey {
					t.Errorf("expected RSS API key %q, got %q", tt.expectRSSAPIKey, conf.RSSReaderAPIKey)
				}
			}
		})
	}
}
