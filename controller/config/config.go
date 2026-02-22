package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Storage struct {
		Type          string `mapstructure:"type"`
		RetentionDays int    `mapstructure:"retention_days"`
		Sqlite        struct {
			Dsn string `mapstructure:"dsn"`
		} `mapstructure:"sqlite"`
		Victoria struct {
			Url string `mapstructure:"url"`
		} `mapstructure:"victoria"`
	} `mapstructure:"storage"`
	Http struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"http"`
}

var Cfg *Config

func LoadConfig(path string) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 默认值
	viper.SetDefault("storage.type", "sqlite")
	viper.SetDefault("storage.retention_days", 30)
	viper.SetDefault("http.port", ":8080")
	viper.SetDefault("storage.sqlite.dsn", "./geegee.db")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found or error parsing (%s), using defaults. Err: %v\n", path, err)
	}

	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}
	log.Printf("Config Loaded: StorageType=[%s], HttpPort=[%s]", Cfg.Storage.Type, Cfg.Http.Port)
}
