package goRedisson

// goRedissonExpirable is the base struct for all expirable objects
type goRedissonExpirable struct {
	*goRedissonObject
}

// newGoRedissonExpirable creates a new goRedissonExpirable
func newGoRedissonExpirable(name string) *goRedissonExpirable {
	return &goRedissonExpirable{
		goRedissonObject: newGoRedissonObject(name),
	}
}
