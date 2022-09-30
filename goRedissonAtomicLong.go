package goRedisson

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type AtomicLong interface {
	GetAndDecrement() (int64, error)
	AddAndGet(int642 int64) int64
	CompareAndSet(int64, int64) (bool, error)
	Get() (int64, error)
	GetAndDelete() (int64, error)
	GetAndAdd(int64) (int64, error)
	GetAndSet(int64) (int64, error)
	IncrementAndGet() int64
	GetAndIncrement() (int64, error)
	Set(int64) error
	DecrementAndGet() int64
}

var (
	_ AtomicLong = (*goRedissonAtomicLong)(nil)
)

type goRedissonAtomicLong struct {
	*goRedissonExpirable
	goRedisson *GoRedisson
}

func NewGoRedissonAtomicLong(goRedisson *GoRedisson, name string) *goRedissonAtomicLong {
	return &goRedissonAtomicLong{
		goRedissonExpirable: NewGoRedissonExpirable(name),
		goRedisson:          goRedisson,
	}
}

func (m *goRedissonAtomicLong) AddAndGet(delta int64) int64 {
	return m.goRedisson.client.IncrBy(context.Background(), m.getRawName(), delta).Val()
}

func (m *goRedissonAtomicLong) CompareAndSet(expect int64, update int64) (bool, error) {
	r, err := m.goRedisson.client.Eval(context.Background(), `
local currValue = redis.call('get', KEYS[1]);
if currValue == ARGV[1]
     or (tonumber(ARGV[1]) == 0 and currValue == false) then
 redis.call('set', KEYS[1], ARGV[2]);
 return 1
else
 return 0
end
`, []string{m.getRawName()}, expect, update).Int()
	if err != nil {
		return false, err
	}
	return r == 1, nil
}

func (m *goRedissonAtomicLong) DecrementAndGet() int64 {
	return m.goRedisson.client.IncrBy(context.Background(), m.getRawName(), -1).Val()
}

func (m *goRedissonAtomicLong) Get() (int64, error) {
	r, err := m.goRedisson.client.Get(context.Background(), m.getRawName()).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return r, err
}

func (m *goRedissonAtomicLong) GetAndDelete() (int64, error) {
	r, err := m.goRedisson.client.Eval(context.Background(), `
local currValue = redis.call('get', KEYS[1]);
redis.call('del', KEYS[1]);
return currValue;
`, []string{m.getRawName()}, m.getRawName()).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return r, err
}

func (m *goRedissonAtomicLong) GetAndAdd(delta int64) (int64, error) {
	v, err := m.goRedisson.client.Do(context.Background(), "INCRBY", m.getRawName(), delta).Int64()
	if err != nil {
		return 0, err
	}
	return v - delta, nil
}

func (m *goRedissonAtomicLong) GetAndSet(newValue int64) (int64, error) {
	f, err := m.goRedisson.client.GetSet(context.Background(), m.getRawName(), newValue).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return f, err
}

func (m *goRedissonAtomicLong) IncrementAndGet() int64 {
	return m.goRedisson.client.IncrBy(context.Background(), m.getRawName(), 1).Val()
}

func (m *goRedissonAtomicLong) GetAndIncrement() (int64, error) {
	return m.GetAndAdd(1)
}

func (m *goRedissonAtomicLong) GetAndDecrement() (int64, error) {
	return m.GetAndAdd(-1)
}

func (m *goRedissonAtomicLong) Set(newValue int64) error {
	return m.goRedisson.client.Do(context.Background(), "SET", m.getRawName(), newValue).Err()
}
