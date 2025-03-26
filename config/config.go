package config

import(
	"fmt"

	"lqkhoi-go-http-api/pkg/utils"
)

type DBConfig struct{
	DBHost string
	DBUser string
	DBPassword string
	DBName string
	DBPort string
}
type RedisConfig struct{
	Addr string
	Password string
}
type AppConfig struct{
	Host string
	Port string
}
type Config struct{
	DBConfig DBConfig
	RedisConfig RedisConfig
	AppConfig AppConfig
}
func NewConfig() *Config{
	return &Config{
	}
}

func (cfg *Config) LoadAppConfig() *Config{
	cfg.AppConfig.Host = utils.GetenvStringValue("APP_HOST","localhost")
	cfg.AppConfig.Port = utils.GetenvStringValue("APP_PORT","3000")
	return cfg
}

func (cfg *Config) LoadDBConfig() *Config {
	cfg.DBConfig.DBHost = utils.GetenvStringValue("DB_HOST","localhost")
	cfg.DBConfig.DBName = utils.GetenvStringValue("DB_NAME","appdb")
	cfg.DBConfig.DBUser = utils.GetenvStringValue("DB_USER","first_user")
	cfg.DBConfig.DBPassword = utils.GetenvStringValue("DB_PASSWORD","first_user_password")
	cfg.DBConfig.DBPort = utils.GetenvStringValue("DB_PORT","5432")
	return cfg
}

func (cfg *Config) LoadRedisConfig() *Config{
	cfg.RedisConfig.Addr = fmt.Sprintf("%s:%s",utils.GetenvStringValue("REDIS_HOST","localhost"),utils.GetenvStringValue("REDIS_PORT","6379"))
	cfg.RedisConfig.Password = utils.GetenvStringValue("REDIS_PASSWORD","redispassword")
	return cfg
}