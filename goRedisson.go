package goRedisson

import (
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/satori/go.uuid"
)

type GoRedisson struct {
	// client redis client
	client *redis.Client
	//watchDogTimeout timeout for watchdog
	watchDogTimeout time.Duration
	//id goRedisson unique uuid
	id string
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

// GetLock returns a Lock named "key" which can be used to lock and unlock the resource "key".
// A Lock can be copied after first use, but most of the time it is advisable to keep instances of Lock.
func (g *GoRedisson) GetLock(key string) Lock {
	return newRedisLock(key, g)
}

// GetReadWriteLock returns a ReadWriteLock named "key" which can be used to lock and unlock the resource "key" when reading or writing.
// A ReadWriteLock can be copied after first use, but most of the time it is advisable to keep instances of ReadWriteLock.
func (g *GoRedisson) GetReadWriteLock(key string) ReadWriteLock {
	return newRedisReadWriteLock(key, g)
}
