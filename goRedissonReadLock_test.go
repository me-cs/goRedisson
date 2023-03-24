package goRedisson

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestReadLock_TryLock(t *testing.T) {
	g := getGodisson()
	mutex := g.GetReadWriteLock("TestReadLock_TryLock")

	err := mutex.ReadLock().TryLock(5 * time.Second)
	if err != nil {
		panic(err)
	}

	time.Sleep(40 * time.Second)
	err = mutex.ReadLock().Unlock()
	if err != nil {
		panic(err)
	}

}

func TestWWLockUnlock(t *testing.T) {
	g := getGodisson()
	var wl ReadWriteLock
	wl = g.GetReadWriteLock("TestWWLockUnlock")
	go func() {
		err := wl.WriteLock().TryLock(3 * time.Second)

		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		err = wl.WriteLock().Unlock()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)
	err := wl.WriteLock().TryLock(4 * time.Second)
	if err != nil {
		panic(err)
	}
	err = wl.WriteLock().Unlock()

	if err != nil {
		panic(err)
	}
}

func TestWRLockUnlock(t *testing.T) {
	g := getGodisson()
	var wl ReadWriteLock
	wl = g.GetReadWriteLock("TestWRLockUnlock")
	go func() {
		err := wl.ReadLock().TryLock(3 * time.Second)
		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		err = wl.ReadLock().Unlock()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)
	err := wl.WriteLock().TryLock(4 * time.Second)
	if err != nil {
		panic(err)
	}
	err = wl.WriteLock().Unlock()

	if err != nil {
		panic(err)
	}
}

func TestRWLockUnlock(t *testing.T) {
	g := getGodisson()
	var wl ReadWriteLock
	wl = g.GetReadWriteLock("TestRWLockUnlock")
	go func() {
		err := wl.WriteLock().TryLock(3 * time.Second)

		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
		err = wl.WriteLock().Unlock()
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)
	err := wl.ReadLock().TryLock(4 * time.Second)
	if err != nil {
		panic(err)
	}
	err = wl.ReadLock().Unlock()

	if err != nil {
		panic(err)
	}
}

func TestRWLock(t *testing.T) {
	g := getGodisson()
	var wl ReadWriteLock
	wl = g.GetReadWriteLock("TestRWLock")
	err := wl.WriteLock().TryLock(3 * time.Second)
	if err != nil {
		panic(err)
	}
	defer wl.WriteLock().Unlock()

	err = wl.ReadLock().TryLock(3 * time.Second)
	if err == nil {
		panic("it should not be nil")
	}
}

func TestRRLock(t *testing.T) {
	g := getGodisson()
	var wl ReadWriteLock
	wl = g.GetReadWriteLock("TestRRLock")
	err := wl.ReadLock().TryLock(3 * time.Second)
	if err != nil {
		panic(err)
	}
	err = wl.ReadLock().TryLock(3 * time.Second)
	if err != nil {
		panic("err")
	}
	err = wl.ReadLock().Unlock()
	if err != nil {
		panic(err)
	}
	err = wl.ReadLock().Unlock()
	if err != nil {
		panic(err)
	}
}

func TestReadLock(t *testing.T) {
	g := getGodisson()
	key := strconv.FormatInt(int64(rand.Int31n(1000000)), 10)
	l := g.GetReadWriteLock(key)
	innerWg := sync.WaitGroup{}
	for i := 0; i < 200; i++ {
		innerWg.Add(1)
		go func() {
			defer innerWg.Done()
			err := l.ReadLock().TryLock(4 * time.Second)
			if err != nil {
				panic(err)
			}
			err = l.ReadLock().Unlock()
			if err != nil {
				panic(err)
			}
		}()
	}
	innerWg.Wait()

}
