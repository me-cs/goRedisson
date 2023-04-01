# goRedisson

*Redisson go implementation*

[![Go](https://github.com/me-cs/goRedisson/workflows/Go/badge.svg?branch=main)](https://github.com/me-cs/goRedisson/actions)
[![codecov](https://codecov.io/gh/me-cs/goRedisson/branch/main/graph/badge.svg)](https://codecov.io/gh/me-cs/goRedisson)
[![Release](https://img.shields.io/github/v/release/me-cs/goRedisson.svg?style=flat-square)](https://github.com/me-cs/goRedisson)
[![Go Report Card](https://goreportcard.com/badge/github.com/me-cs/goRedisson)](https://goreportcard.com/report/github.com/me-cs/goRedisson)
[![Go Reference](https://pkg.go.dev/badge/github.com/me-cs/goRedisson.svg)](https://pkg.go.dev/github.com/me-cs/goRedisson)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Description
redis mutex rwmutex go implementation with watchdog

English | [简体中文](README-CN.md)

### Example use:

```go
package main

import (
	"context"
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
	ctx,cancel:=context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := mutex.LockContext(ctx)
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
	testRwMutex()
	return
}

func testRwMutex() {
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
				ctx,cancel:=context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := l.WriteLock().LockContext(ctx)
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
				ctx,cancel:=context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				err := l.ReadLock().LockContext(ctx)
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
