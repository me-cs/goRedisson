package goRedisson

type goRedissonExpirable struct {
	*goRedissonObject
}

func newGoRedissonExpirable(name string) *goRedissonExpirable {
	return &goRedissonExpirable{
		goRedissonObject: newGoRedissonObject(name),
	}
}
