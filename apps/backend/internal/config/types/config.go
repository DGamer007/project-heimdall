package types

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth AuthConfig `mapstructure:"auth"`
}
