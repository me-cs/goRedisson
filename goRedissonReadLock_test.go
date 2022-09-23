package goRedisson

import (
	"testing"
	"time"
)

func TestReadLock_TryLock(t *testing.T) {
	g := getGodisson()
	mutex := g.GetReadWriteLock("TestReadLock_TryLock")

	err := mutex.readLock().TryLock(5 * time.Second)
	if err != nil {
		panic(err)
	}

	time.Sleep(40 * time.Second)
	err = mutex.readLock().Unlock()
	if err != nil {
		panic(err)
	}

}

func TestWWLockUnlock(t *testing.T) {
	g := getGodisson()
	var wl RReadWriteLock
	wl = g.GetReadWriteLock("TestWWLockUnlock")
	go func() {
		err := wl.writeLock().TryLock(3 * time.Second)

		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		err = wl.writeLock().Unlock()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)
	err := wl.writeLock().TryLock(4 * time.Second)
	if err != nil {
		panic(err)
	}
	err = wl.writeLock().Unlock()

	if err != nil {
		panic(err)
	}
}

func TestWRLockUnlock(t *testing.T) {
	g := getGodisson()
	var wl RReadWriteLock
	wl = g.GetReadWriteLock("TestWRLockUnlock")
	go func() {
		err := wl.readLock().TryLock(3 * time.Second)
		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		err = wl.readLock().Unlock()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)
	err := wl.writeLock().TryLock(4 * time.Second)
	if err != nil {
		panic(err)
	}
	err = wl.writeLock().Unlock()

	if err != nil {
		panic(err)
	}
}

func TestRWLockUnlock(t *testing.T) {
	g := getGodisson()
	var wl RReadWriteLock
	wl = g.GetReadWriteLock("TestRWLockUnlock")
	go func() {
		err := wl.writeLock().TryLock(3 * time.Second)

		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		err = wl.writeLock().Unlock()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)
	err := wl.readLock().TryLock(4 * time.Second)
	if err != nil {
		panic(err)
	}
	err = wl.readLock().Unlock()

	if err != nil {
		panic(err)
	}
}

func TestRWLock(t *testing.T) {
	g := getGodisson()
	var wl RReadWriteLock
	wl = g.GetReadWriteLock("TestRWLock")
	err := wl.writeLock().TryLock(3 * time.Second)
	if err != nil {
		panic(err)
	}
	defer wl.writeLock().Unlock()

	err = wl.readLock().TryLock(3 * time.Second)
	if err == nil {
		panic("it should not be nil")
	}
}

func TestRRLock(t *testing.T) {
	g := getGodisson()
	var wl RReadWriteLock
	wl = g.GetReadWriteLock("TestRRLock")
	err := wl.readLock().TryLock(3 * time.Second)
	if err != nil {
		panic(err)
	}
	err = wl.readLock().TryLock(3 * time.Second)
	if err != nil {
		panic("err")
	}
	err = wl.readLock().Unlock()
	if err != nil {
		panic(err)
	}
	err = wl.readLock().Unlock()
	if err != nil {
		panic(err)
	}
}
