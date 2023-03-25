package goRedisson

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
)

type parseUint64Test struct {
	in  []byte
	out uint64
	err error
}

// ErrRange indicates that a value is out of range for the target type.
var ErrRange = errors.New("value out of range")

// ErrSyntax indicates that a value does not have the right syntax for the target type.
var ErrSyntax = errors.New("invalid syntax")

var parseUint64Tests = []parseUint64Test{
	{[]byte(""), 0, ErrSyntax},
	{[]byte("0"), 0, nil},
	{[]byte("1"), 1, nil},
	{[]byte("12345"), 12345, nil},
	{[]byte("012345"), 12345, nil},
	{[]byte("12345x"), 0, ErrSyntax},
	{[]byte("98765432100"), 98765432100, nil},
	{[]byte("18446744073709551615"), 1<<64 - 1, nil},
	{[]byte("18446744073709551616"), 1<<64 - 1, ErrRange},
	{[]byte("18446744073709551620"), 1<<64 - 1, ErrRange},
	{[]byte("1_2_3_4_5"), 0, ErrSyntax}, // base=10 so no underscores allowed
	{[]byte("_12345"), 0, ErrSyntax},
	{[]byte("1__2345"), 0, ErrSyntax},
	{[]byte("12345_"), 0, ErrSyntax},
	{[]byte("-0)"), 0, ErrSyntax},
	{[]byte("-1)"), 0, ErrSyntax},
	{[]byte("+1)"), 0, ErrSyntax},
}

func TestParseUint64(t *testing.T) {
	for i := range parseUint64Tests {
		test := &parseUint64Tests[i]
		out, err := parseUintBytes(test.in, 10, 64)
		if test.out != out || !reflect.DeepEqual(test.err, err) {
			t.Errorf("ParseUint(%q, 10, 64) = %v, %v want %v, %v",
				test.in, out, err, test.out, test.err)
		}
	}
}

func init() {
	// The parse routines return NumErrors wrapping
	// the error and the string. Convert the tables above.
	for i := range parseUint64Tests {
		test := &parseUint64Tests[i]
		if test.err != nil {
			test.err = &strconv.NumError{Func: "ParseUint", Num: string(test.in), Err: test.err}
		}
	}
}
