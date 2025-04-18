package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPServer  string        `env:"HTTPServer" env-required:"true"`
	StoragePath string        `env:"STORAGE_PATH" env-required:"true"`
	CookieTTL   time.Duration `env:"CookieTTL" env-required:"true"`
	RefreshTTL  time.Duration `env:"RefreshTTL" env-required:"true"`
	AccessTTL   time.Duration `env:"AccessTTL" env-required:"true"`
	HMACSecret  []byte        `env:"HMACSecret" env-required:"true"`
}

func MustLoad() *Config {
	path := getConfigPath()

	if path == "" {
		fmt.Println("CONFIG_PATH is not set, using default .env")
		path = ".env"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists" + path)
	}

	err := godotenv.Load(path)
	if err != nil {
		panic(fmt.Sprintf("No .env file found at %s, relying on environment variables", path))
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Sprintf("Failed to load environment variables: %v", err))
	}

	return &cfg
}

func getConfigPath() string {
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		return envPath
	}

	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	return res
}
