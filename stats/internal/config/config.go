package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-required:"true"`
	DBServer   `yaml:"db_server" env-required:"true"`
	HTTPServer `yaml:"http_server"`
	Kafka      `yaml:"kafka"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Kafka struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

type DBServer struct {
	Host string `yaml:"host" env-required:"true"`
	Port int    `yaml:"port" env-required:"true"`
}

const configPath = "internal/config/local.yaml"

func MustLoad() *Config {
	//check config file existing
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file is not exist. Path: %s", configPath)
	}

	//read config
	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("Can't read config: %s", err)
	}

	return &config
}
