package config

import(
	"fmt"

	"lqkhoi-go-http-api/pkg"
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
type Config struct{
	DBConfig DBConfig
	RedisConfig RedisConfig
}
func NewConfig() *Config{
	return &Config{
	}
}
func (cfg *Config) LoadDBConfig() *Config {
	cfg.DBConfig.DBHost = pkg.GetenvStringValue("DB_HOST","localhost")
	cfg.DBConfig.DBName = pkg.GetenvStringValue("DB_NAME","appdb")
	cfg.DBConfig.DBUser = pkg.GetenvStringValue("DB_USER","first_user")
	cfg.DBConfig.DBPassword = pkg.GetenvStringValue("DB_PASSWORD","first_user_password")
	cfg.DBConfig.DBPort = pkg.GetenvStringValue("DB_PORT","5432")
	return cfg
}
func (cfg *Config)LoadRedisConfig() *Config{
	cfg.RedisConfig.Addr = fmt.Sprintf("%s:%s",pkg.GetenvStringValue("REDIS_HOST","localhost"),pkg.GetenvStringValue("REDIS_PORT","6379"))
	cfg.RedisConfig.Password = pkg.GetenvStringValue("REDIS_PASSWORD","redispassword")
	return cfg
}