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

type parseUint64BaseTest struct {
	in   []byte
	base int
	out  uint64
	err  error
}

var parseUint64BaseTests = []parseUint64BaseTest{
	{[]byte(""), 0, 0, ErrSyntax},
	{[]byte("0"), 0, 0, nil},
	{[]byte("0x"), 0, 0, ErrSyntax},
	{[]byte("0X"), 0, 0, ErrSyntax},
	{[]byte("1"), 0, 1, nil},
	{[]byte("12345"), 0, 12345, nil},
	{[]byte("012345"), 0, 012345, nil},
	{[]byte("0x12345"), 0, 0x12345, nil},
	{[]byte("0X12345"), 0, 0x12345, nil},
	{[]byte("12345x"), 0, 0, ErrSyntax},
	{[]byte("0xabcdefg123"), 0, 0, ErrSyntax},
	{[]byte("123456789abc"), 0, 0, ErrSyntax},
	{[]byte("98765432100"), 0, 98765432100, nil},
	{[]byte("18446744073709551615"), 0, 1<<64 - 1, nil},
	{[]byte("18446744073709551616"), 0, 1<<64 - 1, ErrRange},
	{[]byte("18446744073709551620"), 0, 1<<64 - 1, ErrRange},
	{[]byte("0xFFFFFFFFFFFFFFFF"), 0, 1<<64 - 1, nil},
	{[]byte("0x10000000000000000"), 0, 1<<64 - 1, ErrRange},
	{[]byte("01777777777777777777777"), 0, 1<<64 - 1, nil},
	{[]byte("01777777777777777777778"), 0, 0, ErrSyntax},
	{[]byte("02000000000000000000000"), 0, 1<<64 - 1, ErrRange},
	{[]byte("0200000000000000000000"), 0, 1 << 61, nil},
	{[]byte("0b"), 0, 0, ErrSyntax},
	{[]byte("0B"), 0, 0, ErrSyntax},
	{[]byte("0o"), 0, 0, ErrSyntax},
	{[]byte("0O"), 0, 0, ErrSyntax},

	// underscores allowed with base == 0 only
	{[]byte("_12345"), 0, 0, ErrSyntax},
	{[]byte("1__2345"), 0, 0, ErrSyntax},
	{[]byte("12345_"), 0, 0, ErrSyntax},

	{[]byte("1_2_3_4_5"), 10, 0, ErrSyntax}, // base 10
	{[]byte("_12345"), 10, 0, ErrSyntax},
	{[]byte("1__2345"), 10, 0, ErrSyntax},
	{[]byte("12345_"), 10, 0, ErrSyntax},

	{[]byte("_0x12345"), 0, 0, ErrSyntax},
	{[]byte("0x__12345"), 0, 0, ErrSyntax},
	{[]byte("0x1__2345"), 0, 0, ErrSyntax},
	{[]byte("0x1234__5"), 0, 0, ErrSyntax},
	{[]byte("0x12345_"), 0, 0, ErrSyntax},

	{[]byte("1_2_3_4_5"), 16, 0, ErrSyntax}, // base 16
	{[]byte("_12345"), 16, 0, ErrSyntax},
	{[]byte("1__2345"), 16, 0, ErrSyntax},
	{[]byte("1234__5"), 16, 0, ErrSyntax},
	{[]byte("12345_"), 16, 0, ErrSyntax},
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
	for i := range parseUint64BaseTests {
		test := &parseUint64BaseTests[i]
		if test.err != nil {
			test.err = &strconv.NumError{Func: "ParseUint", Num: string(test.in), Err: test.err}
		}
	}
}

func TestParseUintBytes(t *testing.T) {
	for i := range parseUint64Tests {
		test := &parseUint64Tests[i]
		out, err := parseUintBytes(test.in, 10, 64)
		if test.out != out || !reflect.DeepEqual(test.err, err) {
			t.Errorf("parseUintBytes(%q, 10, 64) = %v, %v want %v, %v",
				test.in, out, err, test.out, test.err)
		}
	}
}

func TestParseUintBytesBase(t *testing.T) {
	for i := range parseUint64BaseTests {
		test := &parseUint64BaseTests[i]
		out, err := parseUintBytes(test.in, test.base, 64)
		if test.out != out || !reflect.DeepEqual(test.err, err) {
			t.Errorf("parseUintBytes(%q, %v, 64) = %v, %v want %v, %v",
				test.in, test.base, out, err, test.out, test.err)
		}
	}
}

func TestParseUintBytes0BitSize(t *testing.T) {
	for i := range parseUint64Tests {
		test := &parseUint64Tests[i]
		out, err := parseUintBytes(test.in, 10, 0)
		if test.out != out || !reflect.DeepEqual(test.err, err) {
			t.Errorf("ParseUint(%q, 10, 0) = %v, %v want %v, %v",
				test.in, out, err, test.out, test.err)
		}
	}
}

func TestCutoff64(t *testing.T) {
	if cutoff64(1) != 0 {
		panic("cutoff64(1)!=0")
	}
}
