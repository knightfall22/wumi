package core_test

import (
	"fmt"
	"testing"

	"github.com/knightfall22/wumi/core"
)

func TestReadSimpleString(t *testing.T) {
	cases := map[string]string{
		"+OK\r\n": "OK",
	}

	for k, v := range cases {
		value, _ := core.Decode([]byte(k))
		if value != v {
			t.Fail()
		}
	}
}

func TestReadError(t *testing.T) {
	cases := map[string]string{
		"-Error Message\r\n": "Error Message",
	}

	for k, v := range cases {
		value, _ := core.Decode([]byte(k))
		if value != v {
			t.Fail()
		}
	}
}

func TestInt64(t *testing.T) {
	cases := map[string]int64{
		":0\r\n":    0,
		":1000\r\n": 1000,
	}

	for k, v := range cases {
		value, _ := core.Decode([]byte(k))
		if value != v {
			t.Fail()
		}
	}
}

func TestReadBulkString(t *testing.T) {
	cases := map[string]string{
		"$5\r\nHello\r\n": "Hello",
		"$0\r\n\r\n":      "",
	}

	for k, v := range cases {
		value, _ := core.Decode([]byte(k))
		if value != v {
			t.Logf("%s != %s", value, v)
		}
	}
}

func TestArrayDecode(t *testing.T) {
	cases := map[string][]any{
		"*0\r\n":                               {},
		"*2\r\n$5\r\nHello\r\n$5\r\nWorld\r\n": {"Hello", "World"},
		"*3\r\n:1\r\n$5\r\nHello\r\n$5\r\nWorld\r\n":           {int64(1), "Hello", "World"},
		"*2\r\n*2\r\n:1\r\n$5\r\nHello\r\n\r\n$5\r\nWorld\r\n": {[]any{int64(1), "Hello"}, "World"},
	}

	for k, v := range cases {
		value, _ := core.Decode([]byte(k))
		array := value.([]any)

		if len(array) != len(v) {
			t.Fail()
		}

		for i := range array {
			if fmt.Sprintf("%v", v[i]) != fmt.Sprintf("%v", array[i]) {
				t.Fail()
			}
		}
	}
}
