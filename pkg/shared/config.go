package shared

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Host     string `mapstructure:"host" default:"0.0.0.0"`
	Port     int    `mapstructure:"port" default:"7000"`
	User     string `mapstructure:"user" default:"docker"`
	DBName   string `mapstructure:"dbname" default"docker"`
	Password string `mapstructure:"password" default:""`
}

func GetConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("/historymap-config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file not found; ignore error if desired
			return nil, err
		}
	}
	config := &Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func GetLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}
