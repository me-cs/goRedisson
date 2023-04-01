package goRedisson

import "strings"

// goRedissonObject is the base struct for all objects
type goRedissonObject struct {
	name string
}

// prefixName prefixes the name with the given prefix
func (m *goRedissonObject) prefixName(prefix string, name string) string {
	if strings.Contains(name, "{") {
		return prefix + ":" + name
	}
	return prefix + ":{" + name + "}"
}

// suffixName suffixes the name with the given suffix
func (m *goRedissonObject) suffixName(name, suffix string) string {
	if strings.Contains(name, "{") {
		return name + ":" + suffix
	}
	return "{" + name + "}:" + suffix
}

// getRawName returns the raw name
func (m *goRedissonObject) getRawName() string {
	return m.name
}

// newGoRedissonObject creates a new goRedissonObject
func newGoRedissonObject(name string) *goRedissonObject {
	return &goRedissonObject{name: name}
}
