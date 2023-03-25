package goRedisson

import "strings"

type goRedissonObject struct {
	name string
}

func (m *goRedissonObject) prefixName(prefix string, name string) string {
	if strings.Contains(name, "{") {
		return prefix + ":" + name
	}
	return prefix + ":{" + name + "}"
}

func (m *goRedissonObject) suffixName(name, suffix string) string {
	if strings.Contains(name, "{") {
		return name + ":" + suffix
	}
	return "{" + name + "}:" + suffix
}

func (m *goRedissonObject) getRawName() string {
	return m.name
}

func newGoRedissonObject(name string) *goRedissonObject {
	return &goRedissonObject{name: name}
}
