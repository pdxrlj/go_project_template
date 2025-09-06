package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App      *AppConfig      `yaml:"app"`
	Database *DatabaseConfig `yaml:"database"`
}

type AppConfig struct {
	Port      string        `yaml:"port"`
	Host      string        `yaml:"host"`
	LogLevel  string        `yaml:"logLevel"`
	LogOutput string        `yaml:"logOutput"`
	Logger    *LoggerConfig `yaml:"logger"`
}

type LoggerConfig struct {
	Rotation      string `yaml:"rotation"`
	RotationSize  int    `yaml:"rotationSize"`
	RotationCount int    `yaml:"rotationCount"`
	RotationTime  string `yaml:"rotationTime"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func InitConfig() *Config {
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	config := &Config{}

	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}
	return config
}

func (c *Config) GetAppConfig() *AppConfig {
	return c.App
}

func (c *Config) GetLoggerConfig() *LoggerConfig {
	return c.App.Logger
}

func (c *Config) GetDatabaseConfig() *DatabaseConfig {
	return c.Database
}
