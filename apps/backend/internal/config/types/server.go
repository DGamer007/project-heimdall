package types

type ServerConfig struct {
	GinMode        string   `mapstructure:"gin_mode" validate:"omitempty,oneof=debug release test"`
	Port           string   `mapstructure:"port" validate:"required,numeric,min=1,max=65535"`
}
