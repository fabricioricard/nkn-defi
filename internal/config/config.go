package config

import "os"

type Config struct {
	DatabaseURL string
	RedisURL    string
	NKNRPCURL   string
	LogLevel    string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/nkndefi?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		NKNRPCURL:   getEnv("NKN_RPC_URL", "https://mainnet-rpc.nkn.org"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}