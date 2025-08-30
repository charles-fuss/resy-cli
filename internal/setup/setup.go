package setup

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

func SurveyConfig() error {
	cfgPath := filepath.Join(".", "secret.yml")

	viper.SetConfigFile(cfgPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read %s: %w", cfgPath, err)
	}

	apiKey := viper.GetString("resy_api_key")
	authToken := viper.GetString("resy_auth_token")

	if apiKey == "" || authToken == "" {
		return fmt.Errorf("%s is missing required keys: resy_api_key and/or resy_auth_token", cfgPath)
	}

	viper.Set("resy_api_key", apiKey)
	viper.Set("resy_auth_token", authToken)

	fmt.Printf("Loaded Resy credentials from %s\n", cfgPath)
	return nil
}
