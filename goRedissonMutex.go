package goRedisson

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// check goRedissonMutex implements Lock
	_ Lock = (*goRedissonMutex)(nil)
)

// goRedissonMutex is a distributed lock implementation
type goRedissonMutex struct {
	goRedissonBaseLock
}

// getChannelName returns the channel name for the lock
func (m *goRedissonMutex) getChannelName() string {
	return m.prefixName("redisson_lock__channel", m.getRawName())
}

// newGoRedissonMutex creates a new goRedissonMutex
func newGoRedissonMutex(name string, goRedisson *GoRedisson) Lock {
	redisLock := &goRedissonMutex{}
	redisLock.goRedissonBaseLock = *newBaseLock(goRedisson.id, name, goRedisson, redisLock)
	return redisLock
}

// tryLockInner tries to acquire the mutex
func (m *goRedissonMutex) tryLockInner(ctx context.Context, leaseTime time.Duration, goroutineId uint64) (*int64, error) {
	result, err := m.goRedisson.client.Eval(ctx, `
if (redis.call('exists', KEYS[1]) == 0) then
    local callVal = redis.call('hsetnx', KEYS[1], ARGV[2], 1);
    if (callVal == 0) then
        redis.call('pexpire', KEYS[1], ARGV[1]);
        return redis.call('pttl', KEYS[1]);
    end ;
    redis.call('pexpire', KEYS[1], ARGV[1]);
    return nil;
end ;
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
    local callVal = redis.call('hsetnx', KEYS[1], ARGV[2], 1);
    if (callVal == 0) then
        redis.call('pexpire', KEYS[1], ARGV[1]);
        return redis.call('pttl', KEYS[1]);
    end ;
    redis.call('pexpire', KEYS[1], ARGV[1]);
    return nil;
end ;
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

// unlockInner releases the mutex
func (m *goRedissonMutex) unlockInner(ctx context.Context, goroutineId uint64) (*int64, error) {
	defer m.cancelExpirationRenewal(goroutineId)
	result, err := m.goRedisson.client.Eval(ctx, `
if (redis.call('hexists', KEYS[1], ARGV[3]) == 0) then
    return nil;
end ;
local val = redis.call('hget', KEYS[1], ARGV[3]);
if (val ~= "1") then
    return nil;
else
    redis.call('del', KEYS[1]);
    redis.call('publish', KEYS[2], ARGV[1]);
    return 1;
end ;
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

// renewExpirationInner renews the mutex expiration
func (m *goRedissonMutex) renewExpirationInner(ctx context.Context, goroutineId uint64) (int64, error) {
	return m.goRedisson.client.Eval(ctx, `
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
    redis.call('pexpire', KEYS[1], ARGV[1]);
    return 1;
end ;
return 0;
`, []string{m.getRawName()}, m.internalLockLeaseTime.Milliseconds(), m.getLockName(goroutineId)).Int64()
}
