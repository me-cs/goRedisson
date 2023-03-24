package goRedisson

import "time"

type InnerLocker interface {
	tryLockInner(time.Duration, time.Duration, uint64) (*int64, error)
	unlockInner(uint64) (*int64, error)
	getChannelName() string
	renewExpirationInner(uint64) (int64, error)
}

type Lock interface {
	TryLock(time.Duration) error
	Unlock() error
}
