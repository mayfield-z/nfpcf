package factory

import (
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Info        *Info        `yaml:"info"`
	Server      *Server      `yaml:"server"`
	NRF         *NRF         `yaml:"nrf"`
	Cache       *Cache       `yaml:"cache"`
	Logger      *Logger      `yaml:"logger"`
}

type Info struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type Server struct {
	BindAddr string `yaml:"bindAddr"`
}

type NRF struct {
	URL string `yaml:"url"`
}

type Cache struct {
	TTL time.Duration `yaml:"ttl"`
}

type Logger struct {
	Level string `yaml:"level"`
}

var NfpcfConfig *Config

func ReadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if config.Cache == nil {
		config.Cache = &Cache{TTL: 5 * time.Minute}
	}

	if config.Server == nil {
		config.Server = &Server{BindAddr: ":8000"}
	}

	if config.Logger == nil {
		config.Logger = &Logger{Level: "info"}
	}

	NfpcfConfig = config
	return config, nil
}
