package config

import (
	"flag"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort    string
	GQLPort     string
	GRPCPort    string
	GRPCHost    string
	EnabledSrvs map[string]bool
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// We don't return this error since missing .env file is not critical
		// as we can still use environment variables
	}

	cfg := &Config{
		HTTPPort:    getEnvWithDefault("HTTP_PORT", "3000"),
		GQLPort:     getEnvWithDefault("GQL_PORT", "8080"),
		GRPCPort:    getEnvWithDefault("GRPC_PORT", "50051"),
		GRPCHost:    getEnvWithDefault("GRPC_HOST", "localhost"),
		EnabledSrvs: make(map[string]bool),
	}

	servers := flag.String("servers", "all", "Comma-separated list of servers to run (http,gql,grpc,all)")
	flag.Parse()

	if *servers == "all" {
		cfg.EnabledSrvs["http"] = true
		cfg.EnabledSrvs["gql"] = true
		cfg.EnabledSrvs["grpc"] = true
	} else {
		for _, srv := range strings.Split(*servers, ",") {
			cfg.EnabledSrvs[strings.TrimSpace(srv)] = true
		}
	}

	return cfg, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) IsServerEnabled(server string) bool {
	return c.EnabledSrvs[server]
}
