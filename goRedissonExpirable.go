package goRedisson

type GoRedissonExpirable struct {
	*GoRedissonObject
}

func NewGoRedissonExpirable(name string) *GoRedissonExpirable {
	return &GoRedissonExpirable{
		GoRedissonObject: NewGoRedissonObject(name),
	}
}
