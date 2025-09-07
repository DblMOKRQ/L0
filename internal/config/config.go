package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Storage  `yaml:"storage"`
	Rest     `yaml:"rest"`
	Kafka    `yaml:"kafka"`
	Redis    `yaml:"redis"`
	LogLevel string `yaml:"log_level"`
}

type Storage struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type Rest struct {
	Addr string `yaml:"addr"`
}

type Kafka struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}
type Redis struct {
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	DB            int    `yaml:"db"`
	Cache
}
type Cache struct {
	TTL   time.Duration `yaml:"ttl"`
	Limit int           `yaml:"limit"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.yaml"
	}
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to get absolute path: %w", err))
	}

	configFile, err := os.Open(absPath)
	if err != nil {
		panic(fmt.Errorf("failed to open config file at %s: %w", absPath, err))
	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		panic(fmt.Errorf("failed to decode config: %w", err))
	}

	return &config
}
