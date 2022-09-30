package goRedisson

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
)

type AtomicDouble interface {
	GetAndDecrement() (float64, error)
	AddAndGet(float642 float64) float64
	CompareAndSet(float64, float64) (bool, error)
	Get() (float64, error)
	GetAndDelete() (float64, error)
	GetAndAdd(float64) (float64, error)
	GetAndSet(float64) (float64, error)
	IncrementAndGet() float64
	GetAndIncrement() (float64, error)
	Set(float64) error
	DecrementAndGet() float64
}

var (
	_ AtomicDouble = (*goRedissonAtomicDouble)(nil)
)

type goRedissonAtomicDouble struct {
	*goRedissonExpirable
	goRedisson *GoRedisson
}

func NewGoRedissonAtomicDouble(goRedisson *GoRedisson, name string) *goRedissonAtomicDouble {
	return &goRedissonAtomicDouble{
		goRedissonExpirable: NewGoRedissonExpirable(name),
		goRedisson:          goRedisson,
	}
}

func (m *goRedissonAtomicDouble) AddAndGet(delta float64) float64 {
	return m.goRedisson.client.IncrByFloat(context.Background(), m.getRawName(), delta).Val()
}

func (m *goRedissonAtomicDouble) CompareAndSet(expect float64, update float64) (bool, error) {
	r, err := m.goRedisson.client.Eval(context.Background(), `
local value = redis.call('get', KEYS[1]);
if (value == false and tonumber(ARGV[1]) == 0) or (tonumber(value) == tonumber(ARGV[1])) then
     redis.call('set', KEYS[1], ARGV[2]);
     return 1
   else
return 0 end
`, []string{m.getRawName()}, strconv.FormatFloat(expect, 'e', -1, 64), strconv.FormatFloat(update, 'e', -1, 64)).Int()
	if err != nil {
		return false, err
	}
	return r == 1, nil
}

func (m *goRedissonAtomicDouble) DecrementAndGet() float64 {
	return m.goRedisson.client.IncrByFloat(context.Background(), m.getRawName(), -1).Val()
}

func (m *goRedissonAtomicDouble) Get() (float64, error) {
	r, err := m.goRedisson.client.Get(context.Background(), m.getRawName()).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return r, err
}

func (m *goRedissonAtomicDouble) GetAndDelete() (float64, error) {
	r, err := m.goRedisson.client.Eval(context.Background(), `
local currValue = redis.call('get', KEYS[1]);
redis.call('del', KEYS[1]);
return currValue;
`, []string{m.getRawName()}, m.getRawName()).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return r, err
}

func (m *goRedissonAtomicDouble) GetAndAdd(delta float64) (float64, error) {
	v, err := m.goRedisson.client.Do(context.Background(), "INCRBYFLOAT", m.getRawName(), delta).Float64()
	if err != nil {
		return 0, err
	}
	return v - delta, nil
}

func (m *goRedissonAtomicDouble) GetAndSet(newValue float64) (float64, error) {
	f, err := m.goRedisson.client.GetSet(context.Background(), m.getRawName(), strconv.FormatFloat(newValue, 'e', -1, 64)).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return f, err
}

func (m *goRedissonAtomicDouble) IncrementAndGet() float64 {
	return m.goRedisson.client.IncrByFloat(context.Background(), m.getRawName(), 1).Val()
}

func (m *goRedissonAtomicDouble) GetAndIncrement() (float64, error) {
	return m.GetAndAdd(1)
}

func (m *goRedissonAtomicDouble) GetAndDecrement() (float64, error) {
	return m.GetAndAdd(-1)
}

func (m *goRedissonAtomicDouble) Set(newValue float64) error {
	return m.goRedisson.client.Do(context.Background(), "SET", m.getRawName(), strconv.FormatFloat(newValue, 'e', -1, 64)).Err()
}
