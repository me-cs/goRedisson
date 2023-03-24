package goRedisson

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	_ Lock = (*goRedissonLock)(nil)
)

var (
	ErrObtainLockTimeout = errors.New("obtained lock timeout")
)

type goRedissonLock struct {
	goRedissonBaseLock
}

func (m *goRedissonLock) getChannelName() string {
	return m.prefixName("redisson_lock__channel", m.getRawName())
}

func (m *goRedissonLock) lock() error {
	return m.TryLock(-1)
}

func newRedisLock(name string, goRedisson *GoRedisson) Lock {
	redisLock := &goRedissonLock{}
	redisLock.goRedissonBaseLock = *newBaseLock(goRedisson.id, name, goRedisson, redisLock)
	return redisLock
}

func (m *goRedissonLock) tryLockInner(_, leaseTime time.Duration, goroutineId uint64) (*int64, error) {
	result, err := m.goRedisson.client.Eval(context.Background(), `
if (redis.call('exists', KEYS[1]) == 0) then
	redis.call('hincrby', KEYS[1], ARGV[2], 1);
	redis.call('pexpire', KEYS[1], ARGV[1]);
	return nil;
end;
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then 
	redis.call('hincrby', KEYS[1], ARGV[2], 1); 
	redis.call('pexpire', KEYS[1], ARGV[1]); 
	return nil; 
end; 
return redis.call('pttl', KEYS[1]);
`, []string{m.getRawName()}, leaseTime.Milliseconds(), m.getLockName(goroutineId)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	if ttl, ok := result.(int64); ok {
		return &ttl, nil
	} else {
		return nil, fmt.Errorf("tryAcquireInner result converter to int64 error, value is %v", result)
	}
}

func (m *goRedissonLock) unlockInner(goroutineId uint64) (*int64, error) {
	defer m.cancelExpirationRenewal(goroutineId)
	result, err := m.goRedisson.client.Eval(context.TODO(), `
if (redis.call('hexists', KEYS[1], ARGV[3]) == 0) then 
	return nil;
end; 
local counter = redis.call('hincrby', KEYS[1], ARGV[3], -1); 
if (counter > 0) then 
	redis.call('pexpire', KEYS[1], ARGV[2]); 
	return 0; 
else 
	redis.call('del', KEYS[1]); 
	redis.call('publish', KEYS[2], ARGV[1]); 
	return 1; 
end; 
return nil;
`, []string{m.getRawName(), m.getChannelName()}, unlockMessage, m.internalLockLeaseTime.Milliseconds(), m.getLockName(goroutineId)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	if b, ok := result.(int64); ok {
		return &b, nil
	} else {
		return nil, fmt.Errorf("unlock result converter to bool error, value is %v", result)
	}
}
