// Package config provides configuration management using Viper.
// Precedence: config.yml < .env < environment variables
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// LoadWithViper loads configuration using Viper with proper precedence.
// Precedence order: config.yml < .env < environment variables
func LoadWithViper() error {
	// Set config file paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Look for config in multiple locations
	viper.AddConfigPath("./configs")      // Docker deployment location
	viper.AddConfigPath(".")              // Current directory
	viper.AddConfigPath("/app/configs")   // Container location

	// Read from config.yml (lowest priority)
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional, continue with env vars only
		fmt.Printf("Warning: No config file found: %v\n", err)
	} else {
		fmt.Printf("Loaded config from: %s\n", viper.ConfigFileUsed())
	}

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Replace dots with underscores in env var names
	// e.g., database.host becomes DATABASE_HOST
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return nil
}

// Helper functions to get config values with proper precedence

func GetString(key string, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

func GetInt(key string, defaultValue int) int {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt(key)
}

func GetBool(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}

func GetDuration(key string, defaultValue string) time.Duration {
	viper.SetDefault(key, defaultValue)
	duration, err := time.ParseDuration(viper.GetString(key))
	if err != nil {
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

func GetStringSlice(key string, defaultValue []string) []string {
	viper.SetDefault(key, defaultValue)
	return viper.GetStringSlice(key)
}

func GetStringMap(key string) map[string]string {
	return viper.GetStringMapString(key)
}
