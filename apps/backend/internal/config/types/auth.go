package types

import (
	"time"
)

type AuthConfig struct {
	JWT JWTConfig `mapstructure:"jwt"`
	Providers ProvidersConfig `mapstructure:"providers"`
}

type JWTConfig struct {
	Issuer string `mapstructure:"issuer" validate:"required"`
	AccessToken TokenConfig `mapstructure:"access_token"`
	RefreshToken TokenConfig `mapstructure:"refresh_token"`
}

type TokenConfig struct {
	Secret string `mapstructure:"secret" validate:"required"`
	TTL time.Duration `mapstructure:"ttl" validate:"required"`
}

type ProvidersConfig struct {
	Local LocalAuthProviderConfig `mapstructure:"local"`
}

type LocalAuthProviderConfig struct {
	Enabled bool `mapstructure:"enabled"`
}
