package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	EnvLocalhost = "LOCALHOST"
	EnvHomeLab   = "HOMELAB"
)

type Config struct {
	App struct {
		Env     string `yaml:"env"`
		Port    string `yaml:"port"`
		BaseURL string `yaml:"base_url"`
	} `yaml:"app"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`

	Google struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectURL  string `yaml:"redirect_url"`
	} `yaml:"google"`

	JWT struct {
		Secret      string `yaml:"secret"`
		ExpiryHours int    `yaml:"expiry_hours"`
	} `yaml:"jwt"`
}

func Load() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = EnvLocalhost
	}

	filePath := fmt.Sprintf("configs/config.%s.yaml", env)

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}
