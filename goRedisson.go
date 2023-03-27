package goRedisson

import (
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/satori/go.uuid"
)

type GoRedisson struct {
	client          *redis.Client
	watchDogTimeout time.Duration
	id              string
}

// DefaultWatchDogTimeout
// The default watchdog timeout, the watchdog will go every 1/3 of the DefaultWatchDogTimeout to renew the lock held by the current thread.
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
