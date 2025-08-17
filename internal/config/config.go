package config

import (
	"flag"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"reflect"
	"time"
)

type Config struct {
	AppEnv      string // Environment type (development, production)
	Timezone    *time.Location
	Server      ServerConfig       `yaml:"server"`
	Database    DatabaseConfig     `yaml:"database"`
	Redis       RedisConfig        `yaml:"redis"`
	Logging     LoggingConfig      `yaml:"logging"`
	SMTPServers []SMTPServerConfig `yaml:"smtp_servers"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // e.g. "json", "text"
}

type SMTPServerConfig struct {
	Name      string `json:"name"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	FromEmail string `yaml:"from_email"`
}

func LoadConfig() (*Config, error) {
	// Load config from environment variables
	config := loadConfigFromEnv()

	// Parse command line flags
	var configFile string
	flag.StringVar(&configFile, "config.file", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load config file
	fileConfig, err := loadConfigFromFile(configFile)
	if err != nil {
		slog.Error("Failed loading configuration file", "config", configFile, "error", err)
		return nil, err
	}

	// Override with config file (if exists)
	if fileConfig != nil {
		config = mergeConfigs(config, fileConfig)
		slog.Info("Loaded configuration file", "config", configFile)
	} else {
		slog.Info("Config file not found or empty, using defaults", "config", configFile)
	}

	config.configLogger()

	// Config timezone
	time.Local = config.Timezone
	slog.Info("Application timezone configured", "timezone", config.Timezone.String())

	return config, nil
}

func (config *Config) configLogger() {
	var logHandler slog.Handler
	var logLevel slog.Level

	switch config.Logging.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
		slog.Warn("Invalid log level, using 'info' instead", "level", config.Logging.Level)
	}

	handlerOptions := &slog.HandlerOptions{Level: logLevel}

	if config.Logging.Format == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, handlerOptions)
	}

	logger := slog.New(logHandler)
	slog.SetDefault(logger)
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}

	tzName := getStringEnv("TZ", "UTC")
	timezone, err := time.LoadLocation(tzName)
	if err != nil {
		slog.Warn("Invalid timezone, using UTC instead", "timezone", tzName, "error", err)
		timezone = time.UTC
	}

	// Default config from environment variables
	config := &Config{
		AppEnv:   getStringEnv("APP_ENV", "development"),
		Timezone: timezone,
	}

	// Server config
	config.Server = ServerConfig{
		Host: getStringEnv("HOST", "0.0.0.0"),
		Port: getIntEnv("PORT", 8080),
	}

	// Database config
	config.Database = DatabaseConfig{
		Host:     getStringEnv("DB_HOST", "localhost"),
		Port:     getIntEnv("DB_PORT", 27017),
		Username: getStringEnv("DB_USER", "mongo"),
		Password: getStringEnv("DB_PASSWORD", "mongo"),
		Name:     getStringEnv("DB_NAME", "notiflow"),
	}

	// Redis config
	config.Redis = RedisConfig{
		Host:     getStringEnv("REDIS_HOST", "localhost"),
		Port:     getIntEnv("REDIS_PORT", 6379),
		Username: getStringEnv("REDIS_USER", ""),
		Password: getStringEnv("REDIS_PASS", ""),
		DB:       getIntEnv("REDIS_DB", 0),
	}

	// Logging config
	config.Logging = LoggingConfig{
		Level:  getStringEnv("LOG_LEVEL", "info"),
		Format: getStringEnv("LOG_FORMAT", "text"),
	}

	// SMTP config
	config.SMTPServers = []SMTPServerConfig{
		{
			Host:      getStringEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:      getIntEnv("SMTP_PORT", 587),
			Username:  getStringEnv("SMTP_USERNAME", ""),
			Password:  getStringEnv("SMTP_PASSWORD", ""),
			FromEmail: getStringEnv("FROM_EMAIL", ""),
		},
	}

	return config
}

// loadConfigFromFile loads configuration from YAML file
func loadConfigFromFile(filename string) (*Config, error) {
	if filename == "" {
		return nil, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist
		}

		slog.Error("Failed to open config file", "filename", filename, "error", err)
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&config); err != nil {
		slog.Error("Failed to decode config file", "filename", filename, "error", err)
		return nil, err
	}

	return &config, nil
}

// mergeConfigs merges two configs using reflection, with override taking precedence over base
func mergeConfigs(base, override *Config) *Config {
	if override == nil {
		return base
	}

	result := *base // Copy base config
	mergeStructs(reflect.ValueOf(&result).Elem(), reflect.ValueOf(override).Elem())
	return &result
}

// mergeStructs recursively merges struct fields using reflection
func mergeStructs(dst, src reflect.Value) {
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.Field(i)

		// Skip unexported fields
		if !srcField.CanInterface() {
			continue
		}

		switch srcField.Kind() {
		case reflect.Struct:
			// Recursively merge nested structs
			mergeStructs(dstField, srcField)
		case reflect.String:
			// Override if source string is not empty
			if srcField.String() != "" {
				dstField.SetString(srcField.String())
			}
		case reflect.Int:
			// Override if source int is not zero
			if srcField.Int() != 0 {
				dstField.SetInt(srcField.Int())
			}
		case reflect.Float64:
			// Override if source float is not zero
			if srcField.Float() != 0 {
				dstField.SetFloat(srcField.Float())
			}
		case reflect.Bool:
			dstField.SetBool(srcField.Bool())
		default:
			// For other types, try direct assignment if possible
			if dstField.CanSet() && srcField.Type() == dstField.Type() {
				dstField.Set(srcField)
			}
		}
	}
}

// getStringEnv retrieves a string environment variable with a default value
func getStringEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv retrieves an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		for _, char := range value {
			if char >= '0' && char <= '9' {
				result = result*10 + int(char-'0')
			} else {
				slog.Warn("Invalid integer value, using default",
					"key", key,
					"value", value,
					"default", defaultValue)
				return defaultValue
			}
		}
		return result
	}
	return defaultValue
}

// getFloatEnv retrieves a float environment variable with a default value
func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var result float64
		var decimalFound bool
		var decimalPlace float64 = 0.1
		for _, char := range value {
			if char >= '0' && char <= '9' {
				if !decimalFound {
					result = result*10 + float64(char-'0')
				} else {
					result += float64(char-'0') * decimalPlace
					decimalPlace *= 0.1
				}
			} else if char == '.' && !decimalFound {
				decimalFound = true
			} else {
				slog.Warn("Invalid float value, using default",
					"key", key,
					"value", value,
					"default", defaultValue)
				return defaultValue
			}
		}
		return result
	}
	return defaultValue
}

// getBoolEnv retrieves a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if value == "true" || value == "1" || value == "yes" {
			return true
		} else if value == "false" || value == "0" || value == "no" {
			return false
		} else {
			slog.Warn("Invalid boolean value, using default",
				"key", key,
				"value", value,
				"default", defaultValue)
			return defaultValue
		}
	}
	return defaultValue
}
