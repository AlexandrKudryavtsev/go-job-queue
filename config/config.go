package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlexandrKudryavtsev/go-job-queue/pkg/logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig  `yaml:"server"`
	Queue  QueueConfig   `yaml:"queue"`
	Logger logger.Config `yaml:"logger"`
}

type ServerConfig struct {
	Port            int      `yaml:"port"`
	ReadTimeout     Duration `yaml:"read_timeout"`
	WriteTimeout    Duration `yaml:"write_timeout"`
	ShutdownTimeout Duration `yaml:"shutdown_timeout"`
}

type QueueConfig struct {
	VisibilityTimeout Duration `yaml:"visibility_timeout"`
	RetryBaseDelay    Duration `yaml:"retry_base_delay"`
	SweepInterval     Duration `yaml:"sweep_interval"`
	MaxPayloadSize    int      `yaml:"max_payload_size"` // bytes
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed read config: %w", err)
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal config: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	// server
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port")
	}
	if c.Server.ReadTimeout.Duration <= 0 {
		return fmt.Errorf("invalid read timeout")
	}
	if c.Server.WriteTimeout.Duration <= 0 {
		return fmt.Errorf("invalid write timeout")
	}
	if c.Server.ShutdownTimeout.Duration <= 0 {
		return fmt.Errorf("invalid shutdown timeout")
	}

	// queue
	if c.Queue.MaxPayloadSize <= 0 || c.Queue.MaxPayloadSize > 1048576 {
		return fmt.Errorf("invalid queue max payload size")
	}
	if c.Queue.RetryBaseDelay.Duration <= 0 {
		return fmt.Errorf("invalid retry base delay")
	}
	if c.Queue.VisibilityTimeout.Duration <= 0 {
		return fmt.Errorf("invalid visibility timeout")
	}

	if c.Queue.SweepInterval.Duration <= 0 {
		return fmt.Errorf("invalid sweep timeout")
	}

	// logger
	format := strings.TrimSpace(strings.ToLower(string(c.Logger.Format)))
	level := strings.TrimSpace(strings.ToLower(c.Logger.Level))

	if format != "json" && format != "text" {
		return fmt.Errorf("invalid logger format")
	}

	switch level {
	case "debug", "info", "warn", "warning", "error":
	default:
		return fmt.Errorf("invalid logger level")
	}

	return nil
}
