package goRedisson

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type goRedissonWriteLock struct {
	goRedissonBaseLock
}

func (m *goRedissonWriteLock) getChannelName() string {
	return m.prefixName("go_redisson_rwlock", m.getRawName())
}

func (m *goRedissonWriteLock) getLockName(goroutineId uint64) string {
	return m.goRedissonBaseLock.getLockName(goroutineId) + ":write"
}

func newRedisWriteLock(name string, goRedisson *GoRedisson) Lock {
	redisWriteLock := &goRedissonWriteLock{}
	redisWriteLock.goRedissonBaseLock = *newBaseLock(goRedisson.id, name, goRedisson, redisWriteLock)
	return redisWriteLock
}

func (m *goRedissonWriteLock) tryLockInner(_, leaseTime time.Duration, goroutineId uint64) (*int64, error) {
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

func (m *goRedissonWriteLock) unlockInner(goroutineId uint64) (*int64, error) {
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
`, []string{m.getRawName(), m.getChannelName()}, readUnlockMessage, m.internalLockLeaseTime.Milliseconds(), m.getLockName(goroutineId)).Result()
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

func (m *goRedissonWriteLock) getKeyPrefix(goroutineId uint64, timeoutPrefix string) string {
	return strings.Split(timeoutPrefix, ":"+m.getLockName(goroutineId))[0]
}

func (m *goRedissonWriteLock) getReadWriteTimeoutNamePrefix(goroutineId uint64) string {
	return m.suffixName(m.getRawName(), m.getLockName(goroutineId)) + ":rwlock_timeout"
}

func (m *goRedissonWriteLock) renewExpirationInner(goroutineId uint64) (int64, error) {
	timeoutPrefix := m.getReadWriteTimeoutNamePrefix(goroutineId)
	keyPrefix := m.getKeyPrefix(goroutineId, timeoutPrefix)

	result, err := m.goRedisson.client.Eval(context.TODO(), `
local counter = redis.call('hget', KEYS[1], ARGV[2]);
if (counter ~= false) then
    redis.call('pexpire', KEYS[1], ARGV[1]);
    
    if (redis.call('hlen', KEYS[1]) > 1) then
        local keys = redis.call('hkeys', KEYS[1]); 
        for n, key in ipairs(keys) do 
            counter = tonumber(redis.call('hget', KEYS[1], key)); 
            if type(counter) == 'number' then 
                for i=counter, 1, -1 do 
                    redis.call('pexpire', KEYS[2] .. ':' .. key .. ':rwlock_timeout:' .. i, ARGV[1]); 
                end; 
            end; 
        end;
    end;
    
    return 1;
end;
return 0;
`, []string{m.getRawName(), keyPrefix}, m.internalLockLeaseTime.Milliseconds(), m.getLockName(goroutineId)).Result()
	if err != nil {
		return 0, err
	}
	if b, ok := result.(int64); ok {
		return b, nil
	} else {
		return 0, fmt.Errorf("try lock result converter to int64 error, value is %v", result)
	}
}
