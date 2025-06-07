package config

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Version of the shared-libs config package
const ConfigVersion = "2.0.0"

// Configuration modes
type ConfigMode string

const (
	ModeBasic         ConfigMode = "basic"          // Original behavior (env vars only)
	ModeSecretManager ConfigMode = "secret_manager" // New behavior with Secret Manager caching
	ModeAuto          ConfigMode = "auto"           // Detect based on environment
)

// Configuration struct to hold all cached secrets and settings
type AppConfig struct {
	Mode           ConfigMode
	AppEnv         string
	ProjectID      string
	MongoURI       string
	DBName         string
	JWTSecret      string
	NATSURL        string
	AllowedOrigins string
	Port           string
	Version        string
	LoadTime       time.Time
}

// Global variables
var (
	Config    *AppConfig
	configMux sync.RWMutex
	once      sync.Once
)

// ConfigOptions allows applications to configure how config is loaded
type ConfigOptions struct {
	Mode                 ConfigMode
	EnableSecretManager  bool
	SecretManagerProject string
	RequiredSecrets      []string
	OptionalSecrets      []string
	FallbackToEnv        bool
}

// LoadEnv loads configuration with default options (backward compatible)
func LoadEnv() {
	LoadEnvWithOptions(ConfigOptions{
		Mode:            ModeAuto,
		FallbackToEnv:   true,
		RequiredSecrets: []string{}, // No required secrets for backward compatibility
	})
}

// LoadEnvWithSecretManager loads configuration with Secret Manager enabled
func LoadEnvWithSecretManager(projectID string, requiredSecrets []string) {
	LoadEnvWithOptions(ConfigOptions{
		Mode:                 ModeSecretManager,
		EnableSecretManager:  true,
		SecretManagerProject: projectID,
		RequiredSecrets:      requiredSecrets,
		FallbackToEnv:        true,
	})
}

// LoadEnvWithOptions provides full control over configuration loading
func LoadEnvWithOptions(options ConfigOptions) {
	once.Do(func() {
		log.Printf("🔧 Loading configuration (shared-libs v%s)...", ConfigVersion)

		config := &AppConfig{
			Mode:     options.Mode,
			AppEnv:   GetEnv("APP_ENV", "development"),
			Port:     GetEnv("PORT", "8080"),
			Version:  ConfigVersion,
			LoadTime: time.Now(),
		}

		// Auto-detect mode if specified
		if config.Mode == ModeAuto {
			config.Mode = detectConfigMode()
		}

		// Load .env file for development
		if config.AppEnv == "development" {
			if err := godotenv.Load(); err != nil {
				log.Println("⚠️  Warning: .env file not found, using system environment variables")
			} else {
				log.Println("✅ Loaded .env file for development")
			}
		}

		// Load configuration based on mode
		var err error
		switch config.Mode {
		case ModeSecretManager:
			config.ProjectID = options.SecretManagerProject
			if config.ProjectID == "" {
				config.ProjectID = GetEnv("GOOGLE_CLOUD_PROJECT", "")
			}
			err = loadSecretsFromManager(config, options)
		case ModeBasic:
			err = loadBasicConfig(config)
		default:
			err = fmt.Errorf("unsupported config mode: %s", config.Mode)
		}

		if err != nil {
			log.Fatalf("❌ Failed to load configuration: %v", err)
		}

		// Thread-safe assignment
		configMux.Lock()
		Config = config
		configMux.Unlock()

		log.Printf("✅ Configuration loaded successfully (mode: %s, env: %s)", config.Mode, config.AppEnv)
	})
}

// detectConfigMode automatically detects the best configuration mode
func detectConfigMode() ConfigMode {
	// Check if running in Google Cloud environment
	if projectID := GetEnv("GOOGLE_CLOUD_PROJECT", ""); projectID != "" {
		log.Println("🔍 Google Cloud environment detected, using Secret Manager mode")
		return ModeSecretManager
	}

	// Check if Secret Manager is explicitly requested
	if GetEnv("USE_SECRET_MANAGER", "") == "true" {
		log.Println("🔍 Secret Manager explicitly enabled")
		return ModeSecretManager
	}

	log.Println("🔍 Standard environment detected, using basic mode")
	return ModeBasic
}

// loadBasicConfig loads configuration using only environment variables (original behavior)
func loadBasicConfig(config *AppConfig) error {
	log.Println("📝 Loading basic configuration from environment variables...")

	config.MongoURI = GetEnv("MONGO_URI", "")
	config.DBName = GetEnv("DB_NAME", "mrexperiences_service")
	config.JWTSecret = GetEnv("JWT_SECRET", "")
	config.NATSURL = GetEnv("NATS_URL", "")
	config.AllowedOrigins = getDefaultAllowedOrigins(config.AppEnv)

	if envOrigins := GetEnv("ALLOWED_ORIGINS", ""); envOrigins != "" {
		config.AllowedOrigins = envOrigins
	}

	log.Println("✅ Basic configuration loaded from environment variables")
	return nil
}

// loadSecretsFromManager loads configuration using Secret Manager with caching
func loadSecretsFromManager(config *AppConfig, options ConfigOptions) error {
	log.Println("🔐 Loading configuration with Secret Manager caching...")

	if config.ProjectID == "" {
		return fmt.Errorf("Google Cloud project ID is required for Secret Manager mode")
	}

	// Load secrets based on requirements
	secretMap := map[string]string{
		"mongo-uri":  "MONGO_URI",
		"db-name":    "DB_NAME",
		"jwt-secret": "JWT_SECRET",
		"nats-url":   "NATS_URL",
	}

	// Load required secrets
	for secretKey, envKey := range secretMap {
		value, err := getSecretOrEnv(config.ProjectID, secretKey, envKey, "", options.FallbackToEnv)
		if err != nil {
			// Check if this is a required secret
			isRequired := contains(options.RequiredSecrets, secretKey)
			if isRequired {
				return fmt.Errorf("required secret %s failed to load: %v", secretKey, err)
			}
			log.Printf("⚠️  Optional secret %s not available: %v", secretKey, err)
			value = ""
		}

		// Assign to config
		switch secretKey {
		case "mongo-uri":
			config.MongoURI = value
		case "db-name":
			config.DBName = value
			if config.DBName == "" {
				config.DBName = "mrexperiences_service" // Default
			}
		case "jwt-secret":
			config.JWTSecret = value
		case "nats-url":
			config.NATSURL = value
		}
	}

	// Load CORS origins
	config.AllowedOrigins = getDefaultAllowedOrigins(config.AppEnv)
	if envOrigins := GetEnv("ALLOWED_ORIGINS", ""); envOrigins != "" {
		config.AllowedOrigins = envOrigins
	}

	log.Printf("✅ Secret Manager configuration loaded (project: %s)", config.ProjectID)
	return nil
}

// getSecretOrEnv tries Secret Manager first, then falls back to environment variables
func getSecretOrEnv(projectID, secretKey, envKey, fallback string, allowFallback bool) (string, error) {
	// Try Secret Manager first
	if value, err := fetchSecretFromManager(projectID, secretKey); err == nil {
		return value, nil
	} else if !allowFallback {
		return "", err
	} else {
		log.Printf("⚠️  Secret Manager failed for %s, falling back to env var %s", secretKey, envKey)
	}

	// Fall back to environment variable
	envValue := GetEnv(envKey, fallback)
	if envValue == "" {
		return "", fmt.Errorf("both secret %s and environment variable %s are empty", secretKey, envKey)
	}

	return envValue, nil
}

// fetchSecretFromManager retrieves a secret from Google Cloud Secret Manager
func fetchSecretFromManager(projectID, secretName string) (string, error) {
	// This function will be implemented in secret_manager.go with build tags
	return getSecretFromGoogleSecretManager(projectID, secretName)
}

// getDefaultAllowedOrigins returns default CORS origins based on environment
func getDefaultAllowedOrigins(env string) string {
	switch env {
	case "production":
		return "https://your-production-domain.com"
	case "staging":
		return "https://staging.your-domain.com"
	default:
		return "http://localhost:5173,http://127.0.0.1:5173"
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetEnv retrieves environment variables with a default fallback (unchanged for compatibility)
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// Thread-safe getter functions (unchanged interface for backward compatibility)
func GetMongoURI() string {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		log.Fatal("Configuration not loaded. Call LoadEnv() first")
	}
	return Config.MongoURI
}

func GetDBName() string {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		log.Fatal("Configuration not loaded. Call LoadEnv() first")
	}
	return Config.DBName
}

func GetJWTSecret() string {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		log.Fatal("Configuration not loaded. Call LoadEnv() first")
	}
	return Config.JWTSecret
}

func GetNATSURL() string {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		log.Fatal("Configuration not loaded. Call LoadEnv() first")
	}
	return Config.NATSURL
}

func GetAllowedOrigins() string {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		log.Fatal("Configuration not loaded. Call LoadEnv() first")
	}
	return Config.AllowedOrigins
}

func GetPort() string {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		log.Fatal("Configuration not loaded. Call LoadEnv() first")
	}
	return Config.Port
}

func GetAppEnv() string {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		log.Fatal("Configuration not loaded. Call LoadEnv() first")
	}
	return Config.AppEnv
}

// GetConfig returns the entire configuration (read-only)
func GetConfig() *AppConfig {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		log.Fatal("Configuration not loaded. Call LoadEnv() first")
	}
	configCopy := *Config
	return &configCopy
}

// GetConfigMode returns the current configuration mode
func GetConfigMode() ConfigMode {
	configMux.RLock()
	defer configMux.RUnlock()
	if Config == nil {
		return ModeBasic
	}
	return Config.Mode
}

// IsSecretManagerEnabled returns true if Secret Manager is being used
func IsSecretManagerEnabled() bool {
	return GetConfigMode() == ModeSecretManager
}
