package config

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type DBConfig struct {
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Port     string `mapstructure:"port"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	Limiter LimiterConfig `mapstructure:"limiter"`
}

type LimiterConfig struct {
	Limit  int    `mapstructure:"limit"`
	Window int    `mapstructure:"window"`
	Prefix string `mapstructure:"prefix"`
}

type Config struct {
	Database DBConfig     `mapstructure:"db"`
	Redis    RedisConfig  `mapstructure:"redis"`
	Server   ServerConfig `mapstructure:"server"`
	// Add JWT secret if needed directly here or in its own struct
	JwtSecret string `mapstructure:"jwt_secret"`
}

func LoadConfig(configPath string) (cfg Config, err error) {

	// --- 1. Set Defaults ---
	// These are the lowest priority and used if no other source provides the value.
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "3000")
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.name", "appdb")
	viper.SetDefault("db.port", "5432")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	if configPath != "" {
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Println("Config file not found, using defaults and environment variables.")
			} else {
				log.Printf("Error reading config file: %v\n", err)
			}
		} else {
			log.Println("Using config file:", viper.ConfigFileUsed())
		}
	} else {
		log.Println("No config file path specified, using defaults and environment variables.")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AllowEmptyEnv(false)

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	if cfg.Redis.Password == "" {
		log.Println("INFO: Redis password (REDIS_PASSWORD) is not set.")
	}
	if cfg.Database.User == "" {
		log.Println("WARNING: Database user (DB_USER) is not set.")
		// Optionally return an error:
		return cfg, errors.New("database user (DB_USER) is required")
	}
	if cfg.Database.Password == "" {
		log.Println("WARNING: Database password (DB_PASSWORD) is not set.")
		// Optionally return an error
		return cfg, errors.New("database password (DB_PASSWORD) is required")
	}
	if cfg.Redis.Password == "" {
		log.Println("INFO: Redis password (REDIS_PASSWORD) is not set.")
	}
	if cfg.JwtSecret == "" {
		log.Println("CRITICAL WARNING: JWT secret (JWT_SECRET) is not set.")
		return cfg, errors.New("jwt secret (JWT_SECRET) is required")
	}
	log.Println("Configuration loaded successfully.")
	return cfg, nil
}

func (rc RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", rc.Host, rc.Port)
}
