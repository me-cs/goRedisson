package goRedisson

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	// unlockMessage is the message sent to the channel when the lock is unlocked
	unlockMessage int64 = 0
	// readUnlockMessage is the message sent to the channel when the lock is unlocked for read
	readUnlockMessage int64 = 1
)

// expirationEntry is a struct that holds the goroutine ids that are waiting for the lock to expire
type expirationEntry struct {
	sync.Mutex
	goroutineIds map[uint64]int64
	cancelFunc   context.CancelFunc
}

// newRenewEntry creates a new expirationEntry
func newRenewEntry() *expirationEntry {
	return &expirationEntry{
		goroutineIds: make(map[uint64]int64),
	}
}

// addGoroutineId adds a goroutine id to the expirationEntry
func (e *expirationEntry) addGoroutineId(goroutineId uint64) {
	e.Lock()
	defer e.Unlock()
	count, ok := e.goroutineIds[goroutineId]
	if ok {
		count++
	} else {
		count = 1
	}
	e.goroutineIds[goroutineId] = count
}

// removeGoroutineId removes a goroutine id from the expirationEntry
func (e *expirationEntry) removeGoroutineId(goroutineId uint64) {
	e.Lock()
	defer e.Unlock()

	count, ok := e.goroutineIds[goroutineId]
	if !ok {
		return
	}
	count--
	if count == 0 {
		delete(e.goroutineIds, goroutineId)
	} else {
		e.goroutineIds[goroutineId] = count
	}
}

// hasNoThreads returns true if there are no goroutines waiting for the lock to expire
func (e *expirationEntry) hasNoThreads() bool {
	e.Lock()
	defer e.Unlock()
	return len(e.goroutineIds) == 0
}

// getFirstGoroutineId returns the first goroutine id in the expirationEntry
func (e *expirationEntry) getFirstGoroutineId() *uint64 {
	e.Lock()
	defer e.Unlock()
	if len(e.goroutineIds) == 0 {
		return nil
	}

	var first = uint64(1<<64 - 1)
	for key := range e.goroutineIds {
		if key <= first {
			first = key
		}
	}
	return &first
}

// goRedissonBaseLock is the base lock struct
type goRedissonBaseLock struct {
	*goRedissonExpirable
	ExpirationRenewalMap  sync.Map
	internalLockLeaseTime time.Duration
	id                    string
	entryName             string
	lock                  innerLocker
	goRedisson            *GoRedisson
}

// newBaseLock creates a new goRedissonBaseLock
func newBaseLock(key, name string, redisson *GoRedisson, locker innerLocker) *goRedissonBaseLock {
	baseLock := &goRedissonBaseLock{
		goRedissonExpirable:   newGoRedissonExpirable(name),
		internalLockLeaseTime: redisson.watchDogTimeout,
		id:                    key,
		lock:                  locker,
		goRedisson:            redisson,
	}
	baseLock.entryName = baseLock.id + ":" + name
	return baseLock
}

// getLockName returns the lock name
func (m *goRedissonBaseLock) getLockName(goroutineId uint64) string {
	return m.id + ":" + strconv.FormatUint(goroutineId, 10)
}

// getEntryName returns the entry name
func (m *goRedissonBaseLock) getEntryName() string {
	return m.entryName
}

// tryAcquire tries to acquire the lock
func (m *goRedissonBaseLock) tryAcquire(ctx context.Context, goroutineId uint64) (*int64, error) {
	ttl, err := m.lock.tryLockInner(ctx, m.internalLockLeaseTime, goroutineId)
	if err != nil {
		return nil, err
	}
	// lock acquired
	if ttl == nil {
		m.scheduleExpirationRenewal(goroutineId)
	}
	return ttl, nil
}

// scheduleExpirationRenewal schedules the expiration renewal
func (m *goRedissonBaseLock) scheduleExpirationRenewal(goroutineId uint64) {
	entry := newRenewEntry()
	oldEntry, stored := m.ExpirationRenewalMap.LoadOrStore(m.getEntryName(), entry)
	if stored {
		oldEntry.(*expirationEntry).addGoroutineId(goroutineId)
	} else {
		entry.addGoroutineId(goroutineId)
		m.renewExpiration()
	}
}

// renewExpiration renews the expiration
func (m *goRedissonBaseLock) renewExpiration() {
	entryName := m.getEntryName()
	ee, ok := m.ExpirationRenewalMap.Load(entryName)
	if !ok {
		return
	}
	ti := time.NewTimer(m.internalLockLeaseTime / 3)

	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		select {
		case <-ti.C:
			ent, ok := m.ExpirationRenewalMap.Load(entryName)
			if !ok {
				return
			}
			goroutineId := ent.(*expirationEntry).getFirstGoroutineId()
			if goroutineId == nil {
				return
			}
			res, err := m.lock.renewExpirationInner(ctx, *goroutineId)
			if err != nil {
				m.ExpirationRenewalMap.Delete(entryName)
				return
			}
			if res != 0 {
				m.renewExpiration()
				return
			}
			m.cancelExpirationRenewal(0)
			return
		case <-ctx.Done():
			return
		}
	}(ctx)
	ee.(*expirationEntry).Lock()
	ee.(*expirationEntry).cancelFunc = cancel
	ee.(*expirationEntry).Unlock()
}

// cancelExpirationRenewal cancels the expiration renewal
func (m *goRedissonBaseLock) cancelExpirationRenewal(goroutineId uint64) {
	entry, ok := m.ExpirationRenewalMap.Load(m.getEntryName())
	if !ok {
		return
	}
	task := entry.(*expirationEntry)
	if goroutineId != 0 {
		task.removeGoroutineId(goroutineId)
	}
	if goroutineId == 0 || task.hasNoThreads() {
		task.Lock()
		if task.cancelFunc != nil {
			task.cancelFunc()
			task.cancelFunc = nil
		}
		task.Unlock()
		m.ExpirationRenewalMap.Delete(m.getEntryName())
	}
}

// Lock locks m. Lock returns when locking is successful or when an exception is encountered.
func (m *goRedissonBaseLock) Lock() error {
	return m.LockContext(context.Background())
}

// LockContext locks m. Lock Returns when locking is successful or when the context timeout or an exception is encountered.
func (m *goRedissonBaseLock) LockContext(ctx context.Context) error {
	goroutineId, err := getId()
	if err != nil {
		return err
	}
	// PubSub
	sub := m.goRedisson.client.Subscribe(ctx, m.lock.getChannelName())
	defer sub.Close()
	defer sub.Unsubscribe(context.TODO(), m.lock.getChannelName())
	ttl := new(int64)
	// fire
	// setting ttl to 0 will allow the for loop to start properly
	*ttl = 0
	for {
		select {
		// obtain lock timeout
		case <-ctx.Done():
			return ErrObtainLockTimeout
		// indicates that the lock has ttl milliseconds to expire
		case <-time.After(time.Duration(*ttl) * time.Millisecond):
			ttl, err = m.tryAcquire(ctx, goroutineId)
		// a lock has been released
		case <-sub.Channel():
			ttl, err = m.tryAcquire(ctx, goroutineId)
		}
		if err != nil {
			return err
		}
		// lock acquired
		if ttl == nil {
			return nil
		}
	}
}

// Unlock unlocks m. Unlock returns when unlocking is successful or when an exception is encountered.
func (m *goRedissonBaseLock) Unlock() error {
	return m.UnlockContext(context.Background())
}

// UnlockContext unlocks m. UnlockContext Returns when unlocking is successful or when the context timeout or an exception is encountered.
func (m *goRedissonBaseLock) UnlockContext(ctx context.Context) error {
	goroutineId, err := getId()
	if err != nil {
		return err
	}
	opStatus, err := m.lock.unlockInner(ctx, goroutineId)
	if err != nil {
		return err
	}
	if opStatus == nil {
		return fmt.Errorf("attempt to unlock lock, not locked by current thread by node id: %s goroutine-id: %d", m.id, goroutineId)
	}
	return nil
}
