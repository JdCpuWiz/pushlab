package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	APNs     APNsConfig     `yaml:"apns"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	APIPort         int           `yaml:"api_port"`
	WorkerCount     int           `yaml:"worker_count"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host               string        `yaml:"host"`
	Port               int           `yaml:"port"`
	User               string        `yaml:"user"`
	Password           string        `yaml:"password"`
	Database           string        `yaml:"database"`
	SSLMode            string        `yaml:"ssl_mode"`
	MaxConnections     int           `yaml:"max_connections"`
	MaxIdleConnections int           `yaml:"max_idle_connections"`
	ConnectionLifetime time.Duration `yaml:"connection_lifetime"`
}

type RabbitMQConfig struct {
	URL            string        `yaml:"url"`
	QueueName      string        `yaml:"queue_name"`
	PrefetchCount  int           `yaml:"prefetch_count"`
	ReconnectDelay time.Duration `yaml:"reconnect_delay"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Enabled  bool   `yaml:"enabled"`
}

type JWTConfig struct {
	Secret       string `yaml:"secret"`
	ExpiryHours  int    `yaml:"expiry_hours"`
	Issuer       string `yaml:"issuer"`
}

type APNsConfig struct {
	DefaultEnvironment  string `yaml:"default_environment"`
	ConnectionPoolSize  int    `yaml:"connection_pool_size"`
	MaxConcurrentPushes int    `yaml:"max_concurrent_pushes"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode)
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	content := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if cfg.Server.APIPort == 0 {
		cfg.Server.APIPort = 8080
	}
	if cfg.Server.WorkerCount == 0 {
		cfg.Server.WorkerCount = 4
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.RabbitMQ.PrefetchCount == 0 {
		cfg.RabbitMQ.PrefetchCount = 10
	}
	if cfg.JWT.ExpiryHours == 0 {
		cfg.JWT.ExpiryHours = 24
	}
	if cfg.APNs.ConnectionPoolSize == 0 {
		cfg.APNs.ConnectionPoolSize = 5
	}
	if cfg.APNs.MaxConcurrentPushes == 0 {
		cfg.APNs.MaxConcurrentPushes = 100
	}

	return &cfg, nil
}

// expandEnvVars replaces ${VAR} or $VAR with environment variable values
func expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		return os.Getenv(key)
	})
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.RabbitMQ.URL == "" {
		return fmt.Errorf("rabbitmq url is required")
	}
	if c.JWT.Secret == "" || c.JWT.Secret == "${JWT_SECRET}" {
		return fmt.Errorf("jwt secret is required (set JWT_SECRET environment variable)")
	}
	if !strings.HasPrefix(c.JWT.Secret, "${") && len(c.JWT.Secret) < 32 {
		return fmt.Errorf("jwt secret must be at least 32 characters")
	}

	return nil
}
