package config

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"lqkhoi-go-http-api/pkg/utils"

	"github.com/spf13/viper"
)

type DBConfig struct {
	Host     string `mapstructure:"host"     validate:"required,min=3"`
	User     string `mapstructure:"user"     validate:"required,min=3"`
	Password string `mapstructure:"password" validate:"required,min=7"`
	Name     string `mapstructure:"name"     validate:"required,min=3"`
	Port     string `mapstructure:"port"     validate:"required"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"     validate:"required,min=3"`
	Port     string `mapstructure:"port"     validate:"required"`
	Password string `mapstructure:"password" validate:"required,min=7"`
}

type ServerConfig struct {
	Host    string        `mapstructure:"host" validate:"required,min=3"`
	Port    string        `mapstructure:"port" validate:"required"`
	Limiter LimiterConfig `mapstructure:"limiter"`
}

type LimiterConfig struct {
	Limit  int    `mapstructure:"limit"  validate:"required,min=5,max=20"`
	Window int    `mapstructure:"window" validate:"required,min=1,max=10"`
	Prefix string `mapstructure:"prefix" validate:"required"`
}

type DateTimeConfig struct {
	Format string `mapstructure:"format" validate:"required"`
}

type Config struct {
	Database  DBConfig       `mapstructure:"db"`
	Redis     RedisConfig    `mapstructure:"redis"`
	Server    ServerConfig   `mapstructure:"server"`
	JwtSecret string         `mapstructure:"jwt_secret" validate:"required,min=15"`
	DateTime  DateTimeConfig `mapstructure:"date_time"`
}

func LoadConfig(configPath string) (cfg Config, err error) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	if configPath != "" {
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				slog.Error("Config file not found, using defaults and environment variables.", "error", err)
			} else {
				slog.Error("Error reading config file", "error", err)
			}
		} else {
			slog.Error("Using config file", "file", viper.ConfigFileUsed())
		}
	} else {
		slog.Info("No config file path specified, using defaults and environment variables.")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.Unmarshal(&cfg)
	if err != nil {
		slog.Error("Unable to decode config into struct", "error", err)
		return cfg, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	slog.Debug("Loaded configuration", "config", cfg)

	slog.Info("Validating configuration")

	err = utils.ValidateStructForConfig(cfg)

	if err != nil {
		return cfg, err
	}

	log.Println("Configuration loaded successfully.")
	return cfg, nil
}

func (rc RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", rc.Host, rc.Port)
}
