# goRedisson

*Redisson golang implementation(Continuous coding in progress)*

## Description
TODO

### Example use:

```go
package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/me-cs/goRedisson"
	"log"
	"time"
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
	return
}

```

## Contributing
Contributing is done with commit code. There is no help that is too small! :) 

If you wish to contribute to this project, please branch and issue a pull request against master ("[GitHub Flow](https://guides.github.com/introduction/flow/)")
