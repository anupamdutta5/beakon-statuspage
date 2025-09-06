package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	Server     ServerConfig `mapstructure:",squash"`
	Database   DatabaseConfig `mapstructure:",squash"`
	JWT        JWTConfig    `mapstructure:",squash"`
	Email      EmailConfig  `mapstructure:",squash"`
}

type ServerConfig struct {
	Port         string `mapstructure:"SERVER_PORT"`
	ReadTimeout  int    `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout int    `mapstructure:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout  int    `mapstructure:"SERVER_IDLE_TIMEOUT"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	DBName   string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
}

type JWTConfig struct {
	Secret    string `mapstructure:"JWT_SECRET"`
	ExpiresIn int    `mapstructure:"JWT_EXPIRES_IN"` // in hours
}

type EmailConfig struct {
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     int    `mapstructure:"SMTP_PORT"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	SMTPFrom     string `mapstructure:"SMTP_FROM"`
	SMTPUseTLS   bool   `mapstructure:"SMTP_USE_TLS"`
	UseSMTP      bool   `mapstructure:"USE_SMTP"`
}

func Load() (*Config, error) {
	// Set default values
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_READ_TIMEOUT", 10)
	viper.SetDefault("SERVER_WRITE_TIMEOUT", 10)
	viper.SetDefault("SERVER_IDLE_TIMEOUT", 120)
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "statuspage")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_SECRET", "change-this-to-a-secure-secret-key")
	viper.SetDefault("JWT_EXPIRES_IN", 24) // hours
	viper.SetDefault("SMTP_HOST", "smtp.gmail.com")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_FROM", "noreply@yourcompany.com")
	viper.SetDefault("SMTP_USE_TLS", true)
	viper.SetDefault("USE_SMTP", false)

	// Read from .env file if it exists
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Try to read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Create config directory if it doesn't exist
	configDir := "./configs"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Create example .env file if it doesn't exist
	exampleEnvPath := filepath.Join(configDir, ".env.example")
	if _, err := os.Stat(exampleEnvPath); os.IsNotExist(err) {
		err := os.WriteFile(exampleEnvPath, []byte(`# Server Configuration
ENVIRONMENT=development
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10
SERVER_WRITE_TIMEOUT=10
SERVER_IDLE_TIMEOUT=120

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=statuspage
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=change-this-to-a-secure-secret-key
JWT_EXPIRES_IN=24

# Email Configuration
USE_SMTP=false
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourcompany.com
SMTP_USE_TLS=true
`), 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to create example .env file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
