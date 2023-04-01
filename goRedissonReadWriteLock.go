package goRedisson

// ReadWriteLock is a interface for read/write lock
type ReadWriteLock interface {
	ReadLock() Lock
	WriteLock() Lock
}

var (
	// check if goRedissonReadWriteLock implements ReadWriteLock
	_ ReadWriteLock = (*goRedissonReadWriteLock)(nil)
)

// goRedissonReadWriteLock is the implementation of ReadWriteLock
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

// newRedisReadWriteLock creates a new goRedissonReadWriteLock
func newRedisReadWriteLock(name string, redisson *GoRedisson) ReadWriteLock {
	return &goRedissonReadWriteLock{
		goRedissonExpirable: newGoRedissonExpirable(name),
		goRedisson:          redisson,
		rLock:               newReadLock(name, redisson),
		wLock:               newRedisWriteLock(name, redisson),
	}
}
