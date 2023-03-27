package goRedisson

import (
	"context"
	"errors"
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
`, []string{m.getRawName()}, leaseTime.Milliseconds(), m.getLockName(goroutineId)).Int64()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return &result, err
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
`, []string{m.getRawName(), m.getChannelName()}, unlockMessage, m.internalLockLeaseTime.Milliseconds(), m.getLockName(goroutineId)).Int64()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return &result, err
}

func (m *goRedissonLock) renewExpirationInner(goroutineId uint64) (int64, error) {
	return m.goRedisson.client.Eval(context.TODO(), `
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
	redis.call('pexpire', KEYS[1], ARGV[1]);
	return 1;
end;
return 0;
`, []string{m.getRawName()}, m.internalLockLeaseTime.Milliseconds(), m.getLockName(goroutineId)).Int64()
}
