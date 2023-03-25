package goRedisson

import "testing"

func TestPrefixName(t *testing.T) {
	o := newGoRedissonObject("test")
	name := o.prefixName("prefix", "{goRedisson}")
	if name != "prefix:{goRedisson}" {
		t.Fatal(name)
	}
	name = o.prefixName("prefix", "goRedisson")
	if name != "prefix:{goRedisson}" {
		t.Fatal(name)
	}
}

func TestSuffixName(t *testing.T) {
	o := newGoRedissonObject("test")
	name := o.suffixName("{goRedisson}", "suffix")
	if name != "{goRedisson}:suffix" {
		t.Fatal(name)
	}
	name = o.suffixName("goRedisson", "suffix")
	if name != "{goRedisson}:suffix" {
		t.Fatal(name)
	}
}
