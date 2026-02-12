package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Api struct {
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ApiAddr            string
	DebugAddr          string
	CORSAllowedOrigins []string
}

type DB struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
}

type Auth struct {
	Key string
}

type Config struct {
	App  Api
	DB   DB
	Auth Auth
}

func NewConfig() *Config {
	_ = godotenv.Load() // Ignore error, env vars might be set directly

	return &Config{
		App: Api{
			ReadTimeout:        getEnvAsDuration("APP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:       getEnvAsDuration("APP_WRITE_TIMEOUT", 5*time.Second),
			IdleTimeout:        getEnvAsDuration("APP_IDLE_TIMEOUT", 60*time.Second),
			ApiAddr:            getEnv("APP_ADDR", ":8080"),
			DebugAddr:          getEnv("DEBUG_ADDR", ":3000"),
			CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		},
		DB: DB{
			User:         getEnv("DB_USER", "sqlite"),
			Password:     getEnv("DB_PASSWORD", ""),
			Host:         getEnv("DB_HOST", "file:app.db?_fk=1"),
			Name:         getEnv("DB_NAME", "multitenant"),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
		},

		Auth: Auth{
			Key: getEnv("AUTH_KEY", "dev-secret-123"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
