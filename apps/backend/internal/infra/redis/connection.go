package redis

import (
	"context"
	internal_errors "heimdall/backend/pkg/errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       int
}

func NewConnection(config Config) (*redis.Client, error) {
	conn := redis.NewClient(&redis.Options{
		Addr: getConnectionAddress(config.Host,config.Port),
		Password: config.Password,
		Username: config.User,
		DB: config.DB,
		Protocol: 2,
		MaxActiveConns: 50,
		ConnMaxLifetime: 5 * time.Minute,
		MaxIdleConns: 5,
		ConnMaxIdleTime: 1 * time.Minute,
	})

	err := conn.Ping(context.Background()).Err()
	if err != nil {
		return nil, internal_errors.NewInfrastructureError("Failed to ping Redis", err)
	}

	return conn, nil
}

func getConnectionAddress(host, port string) string {
	return host + ":" + port
}
