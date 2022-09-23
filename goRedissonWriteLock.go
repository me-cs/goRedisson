package goRedisson

import (
	"context"
	"fmt"
	"time"
)

type RedisWriteLock struct {
	BaseLock
}

func (m *RedisWriteLock) getChannelName() string {
	return m.prefixName("go_redisson_rwlock", m.getRawName())
}

func (m *RedisWriteLock) getReadLockName(goroutineId uint64) string {
	return m.getLockName(goroutineId) + ":write"
}

func NewRedisWriteLock(name string, goRedisson *GoRedisson) *RedisWriteLock {
	redisWriteLock := &RedisWriteLock{}
	redisWriteLock.BaseLock = *NewBaseLock(goRedisson.id, name, goRedisson, redisWriteLock)
	return redisWriteLock
}

func (m *RedisWriteLock) tryLockInner(_, leaseTime time.Duration, goroutineId uint64) (*int64, error) {
	result, err := m.goRedisson.client.Eval(context.Background(), `
local mode = redis.call('hget', KEYS[1], 'mode');
if (mode == false) then
     redis.call('hset', KEYS[1], 'mode', 'write');
     redis.call('hset', KEYS[1], ARGV[2], 1);
     redis.call('pexpire', KEYS[1], ARGV[1]);
     return nil;
end;
if (mode == 'write') then
    if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
        redis.call('hincrby', KEYS[1], ARGV[2], 1); 
        local currentExpire = redis.call('pttl', KEYS[1]);
        redis.call('pexpire', KEYS[1], currentExpire + ARGV[1]);
        return nil;
    end;
end;
return redis.call('pttl', KEYS[1]);
`, []string{m.getRawName()}, leaseTime.Milliseconds(), m.getLockName(goroutineId)).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
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

func (m *RedisWriteLock) unlockInner(goroutineId uint64) (*int64, error) {
	defer m.cancelExpirationRenewal(goroutineId)

	result, err := m.goRedisson.client.Eval(context.TODO(), `
local mode = redis.call('hget', KEYS[1], 'mode');
if (mode == false) then
   redis.call('publish', KEYS[2], ARGV[1]);
   return 1;
end;
if (mode == 'write') then
   local lockExists = redis.call('hexists', KEYS[1], ARGV[3]);
   if (lockExists == 0) then
      return nil;
   else
       local counter = redis.call('hincrby', KEYS[1], ARGV[3], -1);
       if (counter > 0) then
           redis.call('pexpire', KEYS[1], ARGV[2]);
           return 0;
       else
           redis.call('hdel', KEYS[1], ARGV[3]);
           if (redis.call('hlen', KEYS[1]) == 1) then
              redis.call('del', KEYS[1]);
              redis.call('publish', KEYS[2], ARGV[1]); 
           else
              redis.call('hset', KEYS[1], 'mode', 'read');
           end;
           return 1;
       end;
   end;
end;
return nil;
`, []string{m.getRawName(), m.getChannelName()}, UNLOCK_MESSAGE, m.internalLockLeaseTime.Milliseconds(), m.getLockName(goroutineId)).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
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
