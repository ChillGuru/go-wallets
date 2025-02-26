package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `yaml:"env" env-required:"true"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {

	//load local.env
	err := godotenv.Load("local.env")
	if err != nil {
		log.Fatal("Can't read local.env: ", err)
	}

	//get config path
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("Can't load config (CONFIG_PATH is not set)")
	}

	//check config file existing
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file is not exist. Path: %s", configPath)
	}

	//read config
	var config Config
	if err = cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("Can't read config: %s", err)
	}

	return &config
}
