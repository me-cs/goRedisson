package goRedisson

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func getGodisson() *GoRedisson {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return NewGoRedisson(redisDB)
}

func TestMutex_TryLock(t *testing.T) {
	g := getGodisson()
	mutex := g.GetLock("TestMutex_TryLock")

	err := mutex.TryLock(5 * time.Second)
	if err != nil {
		panic(err)
	}

	time.Sleep(40 * time.Second)
	err = mutex.Unlock()
	if err != nil {
		panic(err)
	}

}

func singleLockUnlockTest(times int32, variableName string, g *GoRedisson) error {
	mutex := g.GetLock("plus_" + variableName)
	a := 0
	wg := sync.WaitGroup{}
	total := int32(0)
	for i := int32(0); i < times; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := mutex.TryLock(1 * time.Second)
			if err != nil {
				return
			}
			a++
			err = mutex.Unlock()
			if err != nil {
				panic("unlock failed")
			}
			atomic.AddInt32(&total, 1)
		}()
	}
	wg.Wait()
	log.Println(variableName, "=", a)
	if int32(a) != total {
		return fmt.Errorf("mutex lock and unlock test failed, %s shoule equal %d,but equal %d", variableName, total, a)
	}
	return nil
}

func TestMutex_LockUnlock(t *testing.T) {
	testCase := []int32{1, 10, 100, 200, 300, 330}
	for _, v := range testCase {
		if err := singleLockUnlockTest(v, "variable_1", getGodisson()); err != nil {
			log.Fatalf("err=%v", err)
		}
	}
}

func TestMultiMutex(t *testing.T) {
	testCases := []int32{1, 10, 100, 200}
	id := 0
	getVariableId := func() int {
		id++
		return id
	}
	for _, v := range testCases {
		wg := sync.WaitGroup{}
		numOfFailures := int32(0)
		for i := int32(0); i < v; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := singleLockUnlockTest(10, fmt.Sprintf("variable_%d", getVariableId()), getGodisson())
				if err != nil {
					t.Logf("test failed,err=%v", err)
					atomic.AddInt32(&numOfFailures, 1)
					return
				}
			}()
			wg.Wait()
		}
		if numOfFailures != 0 {
			t.Fatalf("multi mutex test failed, numOfFailures should equal 0,but equal %d", numOfFailures)
		}
	}
}

func TestMutexFairness(t *testing.T) {
	g := getGodisson()
	mu := g.GetLock("TestMutexFairness")
	stop := make(chan bool)
	defer close(stop)
	go func() {
		for {
			err := mu.TryLock(60 * time.Second)
			if err != nil {
				panic(err)
			}
			time.Sleep(100 * time.Microsecond)
			err = mu.Unlock()
			if err != nil {
				panic(err)
			}
			select {
			case <-stop:
				return
			default:
			}
		}
	}()
	done := make(chan bool, 1)
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Microsecond)
			err := mu.TryLock(60 * time.Second)
			if err != nil {
				panic(err)
			}
			err = mu.Unlock()
			if err != nil {
				panic(err)
			}
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("can't acquire Mutex in 10 seconds")
	}
}
