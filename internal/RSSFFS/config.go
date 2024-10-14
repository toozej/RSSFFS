package RSSFFS

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Get environment variables
func getEnvVars() error {
	if _, err := os.Stat(".env"); err == nil {
		// Initialize Viper from .env file
		viper.SetConfigFile(".env") // Specify the name of your .env file

		// Read the .env file
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	// Enable reading environment variables
	viper.AutomaticEnv()

	// get RSS Reader API endpoint and API key from Viper
	apiEndpoint = viper.GetString("RSS_READER_ENDPOINT")
	apiKey = viper.GetString("RSS_READER_API_KEY")
	if apiEndpoint == "" {
		return fmt.Errorf("RSS reader API endpoint must be provided")
	}

	if apiKey == "" {
		return fmt.Errorf("RSS reader API key must be provided")
	}

	return nil
}
