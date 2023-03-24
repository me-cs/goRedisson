package goRedisson

import (
	"sync"
	"testing"
	"time"
)

func TestWriteLock(t *testing.T) {
	l := getGodisson().GetReadWriteLock("WriteLockTest")
	a := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		innerWg := sync.WaitGroup{}
		for i := 0; i < 300; i++ {
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				err := l.WriteLock().TryLock(4 * time.Second)
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
	if a != 300 {
		panic(a)
	}
}
