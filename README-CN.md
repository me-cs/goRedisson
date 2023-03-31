# goRedisson

*Redisson go语言实现*

[![Go](https://github.com/me-cs/goRedisson/workflows/Go/badge.svg?branch=main)](https://github.com/me-cs/goRedisson/actions)
[![codecov](https://codecov.io/gh/me-cs/goRedisson/branch/main/graph/badge.svg)](https://codecov.io/gh/me-cs/goRedisson)
[![Go Report Card](https://goreportcard.com/badge/github.com/me-cs/goRedisson)](https://goreportcard.com/report/github.com/me-cs/goRedisson)
[![Go Reference](https://pkg.go.dev/badge/github.com/me-cs/goRedisson.svg)](https://pkg.go.dev/github.com/me-cs/goRedisson)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Description
带有看门狗的 redis 互斥锁/读写锁 的go语言实现


### 示例:

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
	// 创建redis 客户端
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

	//你的业务代码

	err = mutex.Unlock()
	if err != nil {
		log.Print(err)
		return
	}

	// 或者你也可是使用读写锁
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

## 贡献
贡献是通过提交代码完成的。没有什么帮助是太小的！:)

如果你想为这个项目做贡献，请在主干线上分支并发出拉动请求 ("[GitHub Flow](https://guides.github.com/introduction/flow/)")
