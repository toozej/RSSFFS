package RSSFFS

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestGetEnvVars(t *testing.T) {
	// Define test cases with different scenarios
	tests := []struct {
		name            string
		envVars         map[string]string
		expectError     bool
		expectErrorText string
	}{
		{
			name: "Valid environment variables",
			envVars: map[string]string{
				"RSS_READER_API_KEY":  "valid-api-key",
				"RSS_READER_ENDPOINT": "https://miniflux.example.com",
			},
			expectError: false,
		},
		{
			name:            "Missing RSS_READER_API_KEY",
			envVars:         map[string]string{"RSS_READER_ENDPOINT": "https://miniflux.example.com"},
			expectError:     true,
			expectErrorText: "RSS reader API key must be provided",
		},
		{
			name:            "Missing RSS_READER_ENDPOINT",
			envVars:         map[string]string{"RSS_READER_API_KEY": "valid-api-key"},
			expectError:     true,
			expectErrorText: "RSS reader API endpoint must be provided",
		},
		{
			name:            "No environment variables",
			envVars:         map[string]string{},
			expectError:     true,
			expectErrorText: "RSS reader API endpoint must be provided",
		},
	}

	// Iterate through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables for the test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Ensure Viper reads environment variables
			viper.Reset()
			viper.AutomaticEnv()

			// Call the function and capture the error (if any)
			err := getEnvVars()

			// Check for expected error outcomes
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.expectErrorText {
					t.Errorf("Expected error message '%s' but got '%s'", tt.expectErrorText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Clean up environment variables after the test
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
