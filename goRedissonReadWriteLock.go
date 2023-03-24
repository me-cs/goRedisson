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

func NewRedisReadWriteLock(name string, redisson *GoRedisson) ReadWriteLock {
	return &goRedissonReadWriteLock{
		goRedissonExpirable: NewGoRedissonExpirable(name),
		goRedisson:          redisson,
		rLock:               NewReadLock(name, redisson),
		wLock:               NewRedisWriteLock(name, redisson),
	}
}
