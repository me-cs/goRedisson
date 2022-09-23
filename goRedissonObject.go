package goRedisson

import "strings"

type GoRedissonObject struct {
	name string
}

func (m *GoRedissonObject) prefixName(prefix string, name string) string {
	if strings.Contains(name, "{") {
		return prefix + ":" + name

	}
	return prefix + ":{" + name + "}"
}

func (m *GoRedissonObject) suffixName(name, suffix string) string {
	if strings.Contains(name, "{") {
		return name + ":" + suffix
	}
	return "{" + name + "}" + suffix
}

func (m *GoRedissonObject) getRawName() string {
	return m.name
}

func NewGoRedissonObject(name string) *GoRedissonObject {
	return &GoRedissonObject{name: name}
}
