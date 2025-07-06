package config

import (
	internal_errors "heimdall/backend/pkg/errors"
	internal_struct "heimdall/backend/pkg/struct"
	"strings"

	types "heimdall/backend/internal/config/types"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var AppConfig *types.Config

func LoadConfig(envFileNames ...string) (*types.Config, error) {
	if err := godotenv.Overload(envFileNames...); err != nil {
		return nil, internal_errors.NewInfrastructureError("Failed to load .env file", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".","_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, internal_errors.NewInfrastructureError("Failed to read config file", err)
	}

	var config types.Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, internal_errors.NewInfrastructureError("Failed to unmarshal config", err)
	}

	if err := internal_struct.Validate(&config); err != nil {
		return nil, err
	}

	AppConfig = &config
	return &config, nil
}
