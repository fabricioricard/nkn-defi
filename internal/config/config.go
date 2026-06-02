package config

import "os"

type Config struct {
	DatabaseURL         string
	RedisURL            string
	NKNRPCURL           string
	LogLevel            string
	BaseRPCURL          string // URL do RPC da Base (ex: https://mainnet.base.org)
	WNKNContractAddress string // endereço do contrato wNKN na Base
}

func Load() *Config {
	return &Config{
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/nkndefi?sslmode=disable"),
		RedisURL:            getEnv("REDIS_URL", "localhost:6379"),
		NKNRPCURL:           getEnv("NKN_RPC_URL", "https://mainnet-rpc.nkn.org"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		BaseRPCURL:          getEnv("BASE_RPC_URL", "https://mainnet.base.org"),
		WNKNContractAddress: getEnv("WNKN_CONTRACT_ADDRESS", ""),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func (c *Config) Port() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}