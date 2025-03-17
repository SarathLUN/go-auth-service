package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// Config holds all the configuration settings for the application.
type Config struct {
	DBHost          string `envconfig:"DB_HOST" default:"localhost"`
	DBPort          string `envconfig:"DB_PORT" default:"5432"`
	DBUser          string `envconfig:"DB_USER" default:"postgres"`
	DBPassword      string `envconfig:"DB_PASSWORD" default:"postgres"`
	DBName          string `envconfig:"DB_NAME" default:"postgres"`
	DBSSLMode       string `envconfig:"DB_SSL_MODE" default:"disable"`
	JWTSecret       string `envconfig:"JWT_SECRET" default:"secret" required:"true"`
	SMTPHost        string `envconfig:"SMTP_HOST" default:"smtp.gmail.com"`
	SMTPPort        string `envconfig:"SMTP_PORT" default:"587"`
	SMTPUsername    string `enconfig:"SMTP_USERNAME"`
	SMTPPassword    string `enconfig:"SMTP_PASSWORD"`
	SMTPFromEmail   string `enconfig:"SMTP_FROM_EMAIL"`
	AppPort         string `envconfig:"APP_PORT" default:"8080"`
	ActivateBaseURL string `envconfig:"ACTIVATE_BASE_URL" default:"http://localhost:8080/activate"`
}

var (
	once   sync.Once
	config *Config
)

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	once.Do(func() {
		// load environment variables from .env file (if it exists).
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not found. Using default values.")
		}
	})
	// Retrieve environment variables, providing defaults if not set.
	dbHost := getEnv("DB_HOST", "localhost")
	dbPortStr := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "postgres")
	dbSslMode := getEnv("DB_SSL_MODE", "disable")
	jwtSecret := getEnv("JWT_SECRET", "secret")
	smtpHost := getEnv("SMTP_HOST", "smtp.example.com")               // Example - use your SMTP server
	smtpPortStr := getEnv("SMTP_PORT", "587")                         // Common SMTP ports: 587 (TLS), 465 (SSL)
	smtpUsername := getEnv("SMTP_USERNAME", "")                       // Your SMTP username (if required)
	smtpPassword := getEnv("SMTP_PASSWORD", "")                       // Your SMTP password
	smtpFromEmail := getEnv("SMTP_FROM_EMAIL", "noreply@example.com") // Sender email
	appPort := getEnv("APP_PORT", "8080")                             // Default to port 8080
	activateBaseURL := getEnv("ACTIVATE_BASE_URL", "http://localhost:3000")

	// Create the Config instance.
	config = &Config{
		DBHost:          dbHost,
		DBPort:          dbPortStr,
		DBUser:          dbUser,
		DBPassword:      dbPassword,
		DBName:          dbName,
		DBSSLMode:       dbSslMode,
		JWTSecret:       jwtSecret,
		SMTPHost:        smtpHost,
		SMTPPort:        smtpPortStr,
		SMTPUsername:    smtpUsername,
		SMTPPassword:    smtpPassword,
		SMTPFromEmail:   smtpFromEmail,
		AppPort:         appPort,
		ActivateBaseURL: activateBaseURL,
	}
	return config
}

// getEnv retrieves an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetDBConnectionString builds the database connection string.
func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}
