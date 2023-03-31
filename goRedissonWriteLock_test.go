package goRedisson

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestWriteLockRenew(t *testing.T) {
	g := getGodisson()
	mutex := g.GetReadWriteLock("TestWriteLockRenew")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := mutex.WriteLock().TryLock(ctx)
	if err != nil {
		panic(err)
	}
	time.Sleep(15 * time.Second)
	err = mutex.WriteLock().Unlock()
	if err != nil {
		panic(err)
	}
}

func testWriteLock(times int) {
	l := getGodisson().GetReadWriteLock("TestWriteLock")
	a := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		innerWg := sync.WaitGroup{}
		for i := 0; i < times; i++ {
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				err := l.WriteLock().TryLock(ctx)
				if err != nil {
					panic(err)
				}
				a++
				err = l.WriteLock().Unlock()
				if err != nil {
					panic(err)
				}
			}()
		}
		innerWg.Wait()
	}()
	wg.Wait()
	if a != times {
		panic(a)
	}
}

func TestWriteLock(t *testing.T) {
	for _, v := range []int{1, 10, 100, 200} {
		testWriteLock(v)
	}
}
