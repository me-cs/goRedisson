package goRedisson

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/satori/go.uuid"
)

type GoRedisson struct {
	client          *redis.Client
	watchDogTimeout time.Duration
	id              string
}

var DefaultWatchDogTimeout = 30 * time.Second

func NewGoRedisson(redisClient *redis.Client, opts ...OptionFunc) *GoRedisson {
	g := &GoRedisson{
		client:          redisClient,
		id:              uuid.NewV4().String(),
		watchDogTimeout: DefaultWatchDogTimeout,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

type OptionFunc func(g *GoRedisson)

func WithWatchDogTimeout(t time.Duration) OptionFunc {
	return func(g *GoRedisson) {
		if t.Seconds() < 30 {
			t = DefaultWatchDogTimeout
			log.Println("watchDogTimeout is too small, so config default ")
		}
		g.watchDogTimeout = t
	}
}

func (g *GoRedisson) GetLock(key string) Lock {
	return newRedisLock(key, g)
}

func (g *GoRedisson) GetReadWriteLock(key string) ReadWriteLock {
	return newRedisReadWriteLock(key, g)
}
