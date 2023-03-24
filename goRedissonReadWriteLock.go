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
	rlock      Lock
	wlock      Lock
}

func (m *goRedissonReadWriteLock) ReadLock() Lock {
	return m.rlock
}

func (m *goRedissonReadWriteLock) WriteLock() Lock {
	return m.wlock
}

func NewRedisReadWriteLock(name string, redisson *GoRedisson) ReadWriteLock {
	return &goRedissonReadWriteLock{
		goRedissonExpirable: NewGoRedissonExpirable(name),
		goRedisson:          redisson,
		rlock:               NewReadLock(name, redisson),
		wlock:               NewRedisWriteLock(name, redisson),
	}
}
