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
}


type Config struct {
	Database DBConfig     `mapstructure:"database"`
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
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.name", "appdb")
	viper.SetDefault("database.port", "5432")
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

	// Define a replacer to map env vars like DB_HOST to database.host
	// It replaces dots (".") with underscores ("_") when looking for env vars.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))


	// Optional: Set a prefix if your env vars are like APP_DB_HOST
	// viper.SetEnvPrefix("APP") // Uncomment if you use a prefix like APP_

	// Note: Viper automatically converts env var keys to lowercase for matching.
	// So, DB_HOST -> db_host -> (via replacer) -> database.host

	// Specifically read standalone secrets not nested under prefixes if needed
	// Viper's AutomaticEnv + Replacer usually handles this if names align.
	// e.g., env var JWT_SECRET will map to "jwt_secret" key if no prefix.
	// If you used SetEnvPrefix("APP"), it would look for APP_JWT_SECRET.


	// --- 4. Unmarshal all configuration sources into the Config struct ---
	// Viper applies precedence: Env Vars > Config File > Defaults
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// --- 5. (Optional but Recommended) Validation ---
	// Ensure critical secrets are present
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