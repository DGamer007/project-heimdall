package types

type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type PostgresConfig struct {
	Host string `mapstructure:"host" validate:"required"`
	Port string `mapstructure:"port" validate:"required,numeric,min=1,max=65535"`
	User string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	DBName string `mapstructure:"db_name" validate:"required"`
	SSLMode string `mapstructure:"ssl_mode" validate:"required"`
}
