package goRedisson

type RReadWriteLock interface {
	readLock() RLock
	writeLock() RLock
}

var (
	_ RReadWriteLock = (*RedisReadWriteLock)(nil)
)

type RedisReadWriteLock struct {
	*GoRedissonExpirable
	goRedisson *GoRedisson
}

func (m *RedisReadWriteLock) readLock() RLock {
	return NewReadLock(m.name, m.goRedisson)
}

func (m *RedisReadWriteLock) writeLock() RLock {
	return NewRedisWriteLock(m.name, m.goRedisson)
}

func NewRedisReadWriteLock(name string, redisson *GoRedisson) *RedisReadWriteLock {
	return &RedisReadWriteLock{
		GoRedissonExpirable: NewGoRedissonExpirable(name),
		goRedisson:          redisson,
	}
}
