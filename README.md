# goRedisson

*Redisson golang implementation*

[![codecov](https://codecov.io/gh/me-cs/goRedisson/branch/master/graph/badge.svg)](https://codecov.io/gh/me-cs/goRedisson)
[![Go Report Card](https://goreportcard.com/badge/github.com/me-cs/goRedisson)](https://goreportcard.com/report/github.com/me-cs/goRedisson)

## Description
redis mutex rwmutex golang implementation with watchdog

### Example use:

```go
package main

import (
	"log"
	"sync"
	"time"
    
	"github.com/me-cs/goRedisson"
	"github.com/redis/go-redis/v9"
)

func main() {
	// create redis client
	redisDB := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer redisDB.Close()

	g := goRedisson.NewGoRedisson(redisDB)

	mutex := g.GetLock("example")
	err := mutex.TryLock(time.Second)
	if err != nil {
		log.Print(err)
		return
	}

	//Your business code

	err = mutex.Unlock()
	if err != nil {
		log.Print(err)
		return
	}

	// or you can use a rwlock
	testRwMutest()
	return
}

func testRwMutest() {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer redisDB.Close()

	g := goRedisson.NewGoRedisson(redisDB)
	l := g.GetReadWriteLock("testRwMutest")
	a := 0
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		innerWg := sync.WaitGroup{}
		for i := 0; i < 100; i++ {
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				err := l.WriteLock().TryLock(15 * time.Second)
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

	go func() {
		defer wg.Done()
		innerWg := sync.WaitGroup{}
		for i := 0; i < 100; i++ {
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				err := l.ReadLock().TryLock(15 * time.Second)
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
	}()

	wg.Wait()
	if a != 100 {
		panic(a)
	}
}

```

## Contributing
Contributing is done with commit code. There is no help that is too small! :) 

If you wish to contribute to this project, please branch and issue a pull request against master ("[GitHub Flow](https://guides.github.com/introduction/flow/)")
