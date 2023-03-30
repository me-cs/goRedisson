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
	rLock      Lock //the readLock instance
	wLock      Lock //the writeLock instance
}

// ReadLock return a readLock that can locks/unlocks for reading.
func (m *goRedissonReadWriteLock) ReadLock() Lock {
	return m.rLock
}

// WriteLock return a writeLock that can locks/unlocks for writing.
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
