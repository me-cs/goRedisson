package goRedisson

type goRedissonExpirable struct {
	*goRedissonObject
}

func NewGoRedissonExpirable(name string) *goRedissonExpirable {
	return &goRedissonExpirable{
		goRedissonObject: NewGoRedissonObject(name),
	}
}
