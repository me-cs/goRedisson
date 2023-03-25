package goRedisson

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func getGodisson() *GoRedisson {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     "200.200.107.249:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return NewGoRedisson(redisDB)
}

func TestMutexRenew(t *testing.T) {
	g := getGodisson()
	mutex := g.GetLock("TestMutexRenew")

	err := mutex.TryLock(5 * time.Second)
	if err != nil {
		panic(err)
	}

	time.Sleep(45 * time.Second)
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
	if int32(a) != total {
		return fmt.Errorf("mutex lock and unlock test failed, %s shoule equal %d,but equal %d", variableName, total, a)
	}
	return nil
}

func TestMutex_LockUnlock(t *testing.T) {
	testCase := []int32{1, 10, 100, 200, 300, 330}
	for _, v := range testCase {
		if err := singleLockUnlockTest(v, "variable_1", getGodisson()); err != nil {
			t.Fatalf("err=%v", err)
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

func benchmarkMutex(b *testing.B, slack, work bool) {
	mu := getGodisson().GetLock("benchmarkMutex")
	if slack {
		b.SetParallelism(10)
	}
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			err := mu.TryLock(5 * time.Second)
			if err != nil {
				panic(err)
			}
			err = mu.Unlock()
			if err != nil {
				panic(err)
			}
			if work {
				for i := 0; i < 100; i++ {
					foo *= 2
					foo /= 2
				}
			}
		}
		_ = foo
	})
}

func BenchmarkMutex(b *testing.B) {
	benchmarkMutex(b, false, false)
}

func BenchmarkMutexSlack(b *testing.B) {
	benchmarkMutex(b, true, false)
}

func BenchmarkMutexWork(b *testing.B) {
	benchmarkMutex(b, false, true)
}

func BenchmarkMutexWorkSlack(b *testing.B) {
	benchmarkMutex(b, true, true)
}

func HammerMutex(m Lock, loops int, cdone chan bool) {
	for i := 0; i < loops; i++ {
		if i%3 == 0 {
			if m.TryLock(time.Second) == nil {
				err := m.Unlock()
				if err != nil {
					panic(err)
				}
			}
			continue
		}
		err := m.TryLock(time.Second)
		if err != nil {
			panic(err)
		}
		err = m.Unlock()
		if err != nil {
			panic(err)
		}
	}
	cdone <- true
}

func TestMutex(t *testing.T) {
	if n := runtime.SetMutexProfileFraction(1); n != 0 {
		t.Logf("got mutexrate %d expected 0", n)
	}
	defer runtime.SetMutexProfileFraction(0)

	m := getGodisson().GetLock("TestMutex")

	c := make(chan bool)
	for i := 0; i < 10; i++ {
		go HammerMutex(m, 1000, c)
	}
	for i := 0; i < 10; i++ {
		<-c
	}
}
