package goRedisson

import "time"

type innerLocker interface {
	tryLockInner(time.Duration, time.Duration, uint64) (*int64, error)
	unlockInner(uint64) (*int64, error)
	getChannelName() string
	renewExpirationInner(uint64) (int64, error)
}

// A Lock represents an object that can be locked and unlocked.
type Lock interface {
	TryLock(time.Duration) error
	Unlock() error
}
