package goRedisson

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type RedisReadLock struct {
	BaseLock
}

func NewReadLock(name string, goRedisson *GoRedisson) *RedisReadLock {
	redisReadLock := &RedisReadLock{}
	redisReadLock.BaseLock = *NewBaseLock(goRedisson.id, name, goRedisson, redisReadLock)
	return redisReadLock
}

func (m *RedisReadLock) getChannelName() string {
	return m.prefixName("go_redisson_rwlock", m.getRawName())
}

func (m *RedisReadLock) getWriteLockName(goroutineId uint64) string {
	return m.getLockName(goroutineId) + ":write"
}

func (m *RedisReadLock) getReadWriteTimeoutNamePrefix(goroutineId uint64) string {
	return m.suffixName(m.getRawName(), m.getLockName(goroutineId)) + ":rwlock_timeout"
}

func (m *RedisReadLock) tryAcquire(waitTime, leaseTime time.Duration, goroutineId uint64) (*int64, error) {
	if leaseTime > 0 {
		return m.tryLockInner(waitTime, leaseTime, goroutineId)
	}
	ttl, err := m.tryLockInner(waitTime, m.internalLockLeaseTime, goroutineId)
	if err != nil {
		return nil, err
	}
	// lock acquired
	if ttl == nil {
		if leaseTime > 0 {
			m.internalLockLeaseTime = leaseTime
		} else {
			m.scheduleExpirationRenewal(goroutineId)
		}
	}
	return ttl, nil
}

func (m *RedisReadLock) Unlock() error {
	goroutineId, err := GetId()
	if err != nil {
		return err
	}
	opStatus, err := m.unlockInner(goroutineId)
	if err != nil {
		return err
	}
	if opStatus == nil {
		return fmt.Errorf("attempt to unlock lock, not locked by current thread by node id: %s goroutine-id: %d", m.id, goroutineId)
	}
	return nil
}

func (m *RedisReadLock) tryLockInner(_, leaseTime time.Duration, goroutineId uint64) (*int64, error) {
	result, err := m.goRedisson.client.Eval(context.Background(), `
local mode = redis.call('hget', KEYS[1], 'mode');
if (mode == false) then
 redis.call('hset', KEYS[1], 'mode', 'read');
 redis.call('hset', KEYS[1], ARGV[2], 1);
 redis.call('set', KEYS[2] .. ':1', 1);
 redis.call('pexpire', KEYS[2] .. ':1', ARGV[1]);
 redis.call('pexpire', KEYS[1], ARGV[1]);
 return nil;
end;
if (mode == 'read') or (mode == 'write' and redis.call('hexists', KEYS[1], ARGV[3]) == 1) then
 local ind = redis.call('hincrby', KEYS[1], ARGV[2], 1); 
 local key = KEYS[2] .. ':' .. ind
 redis.call('set', key, 1);
 redis.call('pexpire', key, ARGV[1]);
 local remainTime = redis.call('pttl', KEYS[1]);
 redis.call('pexpire', KEYS[1], math.max(remainTime, ARGV[1]));
 return nil;
end;
return redis.call('pttl', KEYS[1]);
`, []string{m.getRawName(), m.getReadWriteTimeoutNamePrefix(goroutineId)}, leaseTime.Milliseconds(),
		m.getLockName(goroutineId), m.getWriteLockName(goroutineId)).Result()
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

func (m *RedisReadLock) getKeyPrefix(goroutineId uint64, timeoutPrefix string) string {
	return strings.Split(timeoutPrefix, ":"+m.getLockName(goroutineId))[0]
}

func (m *RedisReadLock) unlockInner(goroutineId uint64) (*int64, error) {
	defer m.cancelExpirationRenewal(goroutineId)

	timeoutPrefix := m.getReadWriteTimeoutNamePrefix(goroutineId)
	keyPrefix := m.getKeyPrefix(goroutineId, timeoutPrefix)

	result, err := m.goRedisson.client.Eval(context.TODO(), `
local mode = redis.call('hget', KEYS[1], 'mode');
if (mode == false) then
   redis.call('publish', KEYS[2], ARGV[1]);
   return 1;
end;
local lockExists = redis.call('hexists', KEYS[1], ARGV[2]);
if (lockExists == 0) then
   return nil
end;
   
local counter = redis.call('hincrby', KEYS[1], ARGV[2], -1);
if (counter == 0) then
   redis.call('hdel', KEYS[1], ARGV[2]); 
end;
redis.call('del', KEYS[3] .. ':' .. (counter+1));

if (redis.call('hlen', KEYS[1]) > 1) then
   local maxRemainTime = -3;
   local keys = redis.call('hkeys', KEYS[1]); 
   for n, key in ipairs(keys) do
       counter = tonumber(redis.call('hget', KEYS[1], key)); 
       if type(counter) == 'number' then
          for i=counter, 1, -1 do 
             local remainTime = redis.call('pttl', KEYS[4] .. ':' .. key .. ':rwlock_timeout:' .. i); 
             maxRemainTime = math.max(remainTime, maxRemainTime); 
          end; 
       end; 
   end;
           
   if maxRemainTime > 0 then
      redis.call('pexpire', KEYS[1], maxRemainTime);
      return 0;
   end; 
       
   if mode == 'write' then 
      return 0; 
   end;
end;
   
redis.call('del', KEYS[1]);
redis.call('publish', KEYS[2], ARGV[1]);
return 1; 
`, []string{m.getRawName(), m.getChannelName(), timeoutPrefix, keyPrefix}, UNLOCK_MESSAGE, m.getLockName(goroutineId)).Result()
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

func (m *RedisReadLock) renewExpirationInner(goroutineId uint64) (int64, error) {
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
