package redis

import (
	"fmt"
	"git.jetbrains.space/orbi/fcsd/kit/log"
	"github.com/go-redis/redis"
	"time"
)

type Redis struct {
	Instance *redis.Client
	Ttl      time.Duration
	logger   log.CLoggerFunc
}

// Config redis config
type Config struct {
	Host     string
	Port     string
	Password string
	Ttl      uint
}

func Open(params *Config, logger log.CLoggerFunc) (*Redis, error) {

	l := logger().Cmp("redis").Mth("open")

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", params.Host, params.Port),
		Password: params.Password,
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, ErrRedisPingErr(err)
	}

	l.Inf("ok")
	return &Redis{
		Instance: client,
		Ttl:      time.Duration(params.Ttl) * time.Second,
		logger:   logger,
	}, nil
}

func (r *Redis) Close() {
	if r.Instance != nil {
		_ = r.Instance.Close()
	}
}
