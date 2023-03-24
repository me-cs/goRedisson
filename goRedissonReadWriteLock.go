package goRedisson

type ReadWriteLock interface {
	ReadLock() Lock
	WriteLock() Lock
}

var (
	_ ReadWriteLock = (*goRedissonReadWriteLock)(nil)
)

type goRedissonReadWriteLock struct {
	*goRedissonExpirable
	goRedisson *GoRedisson
	rLock      Lock
	wLock      Lock
}

func (m *goRedissonReadWriteLock) ReadLock() Lock {
	return m.rLock
}

func (m *goRedissonReadWriteLock) WriteLock() Lock {
	return m.wLock
}

func newRedisReadWriteLock(name string, redisson *GoRedisson) ReadWriteLock {
	return &goRedissonReadWriteLock{
		goRedissonExpirable: newGoRedissonExpirable(name),
		goRedisson:          redisson,
		rLock:               newReadLock(name, redisson),
		wLock:               newRedisWriteLock(name, redisson),
	}
}
