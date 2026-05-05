package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Config struct {
	AppPort    string
	AppEnv     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
	JWTExpiry  string

	// Pool Postgres (sesuaikan dengan PgBouncer / max_connections)
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// Timeout HTTP server (upgrade WebSocket selesai cepat; body besar tetap dibatasi ReadTimeout)
	HTTPReadHeaderTimeout time.Duration
	HTTPReadTimeout       time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	HTTPShutdownTimeout   time.Duration

	// Midtrans
	MidtransServerKey string
	MidtransBaseURL   string

	// MongoDB (untuk notifikasi + outbox)
	MongoURI             string
	MongoDBName          string
	MongoMaxPoolSize     uint64
	MongoMinPoolSize     uint64
	MongoMaxConnIdleTime time.Duration

	// Redis (opsional — antrian webhook Midtrans / scaling)
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

var Cfg *Config

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	maxOpen := parseIntEnv("DB_MAX_OPEN_CONNS", 25)
	maxIdle := parseIntEnv("DB_MAX_IDLE_CONNS", 5)
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	Cfg = &Config{
		AppPort:    getEnv("APP_PORT", "8080"),
		AppEnv:     getEnv("APP_ENV", "development"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "pos_koperasi"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "secret"),
		JWTExpiry:  getEnv("JWT_EXPIRY_HOURS", "24"),

		DBMaxOpenConns:        maxOpen,
		DBMaxIdleConns:        maxIdle,
		DBConnMaxLifetime:     parseDurationEnv("DB_CONN_MAX_LIFETIME", time.Hour),
		HTTPReadHeaderTimeout: parseDurationEnv("HTTP_READ_HEADER_TIMEOUT", 10*time.Second),
		HTTPReadTimeout:       parseDurationEnv("HTTP_READ_TIMEOUT", 60*time.Second),
		HTTPWriteTimeout:      parseDurationEnv("HTTP_WRITE_TIMEOUT", 60*time.Second),
		HTTPIdleTimeout:       parseDurationEnv("HTTP_IDLE_TIMEOUT", 120*time.Second),
		HTTPShutdownTimeout:   parseDurationEnv("HTTP_SHUTDOWN_TIMEOUT", 25*time.Second),

		MidtransServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransBaseURL:      getEnv("MIDTRANS_BASE_URL", "https://api.sandbox.midtrans.com"),
		MongoURI:             getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:          getEnv("MONGO_DB_NAME", "pos_koperasi"),
		MongoMaxPoolSize:     parseUint64Env("MONGO_MAX_POOL_SIZE", 100),
		MongoMinPoolSize:     parseUint64Env("MONGO_MIN_POOL_SIZE", 0),
		MongoMaxConnIdleTime: parseDurationEnv("MONGO_MAX_CONN_IDLE_TIME", 10*time.Minute),

		RedisAddr:     getEnv("REDIS_ADDR", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       parseIntEnv("REDIS_DB", 0),
	}
	return Cfg
}

func parseIntEnv(key string, def int) int {
	s := getEnv(key, "")
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseUint64Env(key string, def uint64) uint64 {
	s := getEnv(key, "")
	if s == "" {
		return def
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}

// parseDurationEnv memahami nilai seperti "60s", "2m", "0" (tanpa timeout untuk HTTP jika 0).
func parseDurationEnv(key string, def time.Duration) time.Duration {
	s := getEnv(key, "")
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
