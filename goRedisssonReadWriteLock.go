package goRedisson

type ReadWriteLock interface {
	readLock() Lock
	writeLock() Lock
}

var (
	_ ReadWriteLock = (*goRedissonReadWriteLock)(nil)
)

type goRedissonReadWriteLock struct {
	*goRedissonExpirable
	goRedisson *GoRedisson
}

func (m *goRedissonReadWriteLock) readLock() Lock {
	return NewReadLock(m.name, m.goRedisson)
}

func (m *goRedissonReadWriteLock) writeLock() Lock {
	return NewRedisWriteLock(m.name, m.goRedisson)
}

func NewRedisReadWriteLock(name string, redisson *GoRedisson) *goRedissonReadWriteLock {
	return &goRedissonReadWriteLock{
		goRedissonExpirable: NewGoRedissonExpirable(name),
		goRedisson:          redisson,
	}
}
