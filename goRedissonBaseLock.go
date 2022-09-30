package goRedisson

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	UNLOCK_MESSAGE      int64 = 0
	READ_UNLOCK_MESSAGE int64 = 1
)

type ExpirationEntry struct {
	sync.Mutex
	goroutineIds map[uint64]int64
	cancelFunc   context.CancelFunc
}

func NewRenewEntry() *ExpirationEntry {
	return &ExpirationEntry{
		goroutineIds: make(map[uint64]int64),
	}
}

func (e *ExpirationEntry) addGoroutineId(goroutineId uint64) {
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

func (e *ExpirationEntry) removeGoroutineId(goroutineId uint64) {
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

func (e *ExpirationEntry) hasNoThreads() bool {
	return len(e.goroutineIds) == 0
}

func (e *ExpirationEntry) getFirstGoroutineId() *uint64 {
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

type goRedissonBaseLock struct {
	*goRedissonExpirable
	ExpirationRenewalMap  sync.Map
	internalLockLeaseTime time.Duration
	id                    string
	entryName             string
	lock                  InnerLocker
	goRedisson            *GoRedisson
}

func NewBaseLock(key, name string, redisson *GoRedisson, locker InnerLocker) *goRedissonBaseLock {
	baseLock := &goRedissonBaseLock{
		goRedissonExpirable:   NewGoRedissonExpirable(name),
		internalLockLeaseTime: redisson.watchDogTimeout,
		id:                    key,
		lock:                  locker,
		goRedisson:            redisson,
	}
	baseLock.entryName = baseLock.id + ":" + name
	return baseLock
}

func (m *goRedissonBaseLock) getLockName(goroutineId uint64) string {
	return m.id + ":" + strconv.FormatUint(goroutineId, 10)
}

func (m *goRedissonBaseLock) getEntryName() string {
	return m.entryName
}

func (m *goRedissonBaseLock) tryAcquire(waitTime, leaseTime time.Duration, goroutineId uint64) (*int64, error) {
	if leaseTime > 0 {
		return m.lock.tryLockInner(waitTime, leaseTime, goroutineId)
	}
	ttl, err := m.lock.tryLockInner(waitTime, m.internalLockLeaseTime, goroutineId)
	if err != nil {
		return nil, err
	}
	// lock acquired
	if ttl == nil {
		if leaseTime > 0 {
			m.internalLockLeaseTime = leaseTime
		} else {
			m.scheduleExpirationRenewal(goroutineId)
		}
	}
	return ttl, nil
}

func (m *goRedissonBaseLock) scheduleExpirationRenewal(goroutineId uint64) {
	entry := NewRenewEntry()
	oldEntry, stored := m.ExpirationRenewalMap.LoadOrStore(m.getEntryName(), entry)
	if stored {
		oldEntry.(*ExpirationEntry).addGoroutineId(goroutineId)
	} else {
		entry.addGoroutineId(goroutineId)
		m.renewExpiration()
		//todo
		// how to impl this code in java with golang?
		// if (Thread.currentThread().isInterrupted()) {
		//                    cancelExpirationRenewal(goroutineId);
		//                }
		//m.cancelExpirationRenewal(goroutineId)
	}
}

func (m *goRedissonBaseLock) renewExpirationInner(goroutineId uint64) (int64, error) {
	result, err := m.goRedisson.client.Eval(context.TODO(), `
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then
	redis.call('pexpire', KEYS[1], ARGV[1]);
	return 1;
end;
return 0;
`, []string{m.getRawName()}, m.internalLockLeaseTime.Milliseconds(), m.getLockName(goroutineId)).Result()
	if err != nil {
		return 0, err
	}
	if b, ok := result.(int64); ok {
		return b, nil
	} else {
		return 0, fmt.Errorf("try lock result converter to int64 error, value is %v", result)
	}
}

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
			goroutineId := ent.(*ExpirationEntry).getFirstGoroutineId()
			if goroutineId == nil {
				return
			}
			res, err := m.renewExpirationInner(*goroutineId)
			if err != nil {
				m.ExpirationRenewalMap.Delete(entryName)
				return
			}
			if res != 0 {
				m.renewExpiration()
			}
			m.cancelExpirationRenewal(0)
			return
		case <-ctx.Done():
			return
		}
	}(ctx)

	ee.(*ExpirationEntry).cancelFunc = cancel

}

func (m *goRedissonBaseLock) cancelExpirationRenewal(goroutineId uint64) {
	entry, ok := m.ExpirationRenewalMap.Load(m.getEntryName())
	if !ok {
		return
	}
	task := entry.(*ExpirationEntry)
	if goroutineId != 0 {
		task.removeGoroutineId(goroutineId)
	}
	if goroutineId == 0 || task.hasNoThreads() {
		if task.cancelFunc != nil {
			task.cancelFunc()
			task.cancelFunc = nil
		}
		m.ExpirationRenewalMap.Delete(m.getEntryName())
	}
}

func (m *goRedissonBaseLock) TryLock(waitTime time.Duration) error {
	wait := waitTime.Milliseconds()
	current := time.Now().UnixMilli()
	goroutineId, err := GetId()
	if err != nil {
		return err
	}
	ttl, err := m.tryAcquire(waitTime, -1, goroutineId)
	if err != nil {
		return err
	}
	// lock acquired
	if ttl == nil {
		return nil
	}
	wait -= time.Now().UnixMilli() - current
	if wait <= 0 {
		return ErrObtainLockTimeout
	}
	current = time.Now().UnixMilli()
	// PubSub
	sub := m.goRedisson.client.Subscribe(context.TODO(), m.lock.getChannelName())
	defer sub.Close()
	defer sub.Unsubscribe(context.TODO(), m.lock.getChannelName())

	wait -= time.Now().UnixMilli() - current
	if wait <= 0 {
		return ErrObtainLockTimeout
	}

	for {
		currentTime := time.Now().UnixMilli()
		ttl, err = m.tryAcquire(waitTime, -1, goroutineId)
		if err != nil {
			return err
		}
		// lock acquired
		if ttl == nil {
			return nil
		}
		wait -= time.Now().UnixMilli() - currentTime
		if wait <= 0 {
			return ErrObtainLockTimeout
		}
		currentTime = time.Now().UnixMilli()
		if *ttl >= 0 && *ttl < wait {
			tCtx, _ := context.WithTimeout(context.TODO(), time.Duration(*ttl)*time.Millisecond)
			_, err := sub.ReceiveMessage(tCtx)
			if err != nil {
				//if errors.As(err, &target) {
				//	continue
				//}
			}
		} else {
			tCtx, _ := context.WithTimeout(context.TODO(), time.Duration(wait)*time.Millisecond)
			_, err := sub.ReceiveMessage(tCtx)
			if err != nil {
				//if errors.As(err, &target) {
				//	continue
				//}
			}
		}
		wait -= time.Now().UnixMilli() - currentTime
		if wait <= 0 {
			return ErrObtainLockTimeout
		}
	}
}

func (m *goRedissonBaseLock) Unlock() error {
	goroutineId, err := GetId()
	if err != nil {
		return err
	}
	opStatus, err := m.lock.unlockInner(goroutineId)
	if err != nil {
		return err
	}
	if opStatus == nil {
		return fmt.Errorf("attempt to unlock lock, not locked by current thread by node id: %s goroutine-id: %d", m.id, goroutineId)
	}
	return nil
}
