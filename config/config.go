package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	// Server
	ServerPort string
	IsProd     bool

	// Database
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string

	// JWT
	JWTSecret       string
	AccessTokenExp  int
	RefreshTokenExp int
	Issuer          string

	// Midtrans
	MidtransServerKey string
	MidtransClientKey string
	MidtransEnv       string
}

func LoadConfig() *Config {
	// Load .env file jika ada
	godotenv.Load()

	return &Config{
		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),
		IsProd:     getEnvAsBool("IS_PROD", false),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "toyrentals"),
		DBPort:     getEnv("DB_PORT", "5432"),

		// JWT
		JWTSecret:       getEnv("JWT_SECRET", "secret"),
		AccessTokenExp:  getEnvAsInt("ACCESS_TOKEN_EXP", 1),
		RefreshTokenExp: getEnvAsInt("REFRESH_TOKEN_EXP", 7),
		Issuer:          getEnv("ISSUER", "toyrentals"),

		// Midtrans
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", "HEHEHE"),
		MidtransClientKey: getEnv("MIDTRANS_CLIENT_KEY", "HEHEHE"),
		MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
	}

}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
