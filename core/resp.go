package core

import (
	"errors"
	"fmt"
)

func Decode(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}

	value, _, err := DecodeOne(data)
	return value, err
}

func DecodeArrayString(data []byte) ([]string, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	ts := value.([]any)
	tokens := make([]string, len(ts))

	for i := range tokens {
		tokens[i] = ts[i].(string)
	}

	return tokens, nil
}

func DecodeOne(data []byte) (any, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}

	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInt64(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}

	return nil, 0, nil
}

// read a RESP encoded simple string from data and returns
// the string, delta and the error
func readSimpleString(data []byte) (string, int, error) {
	pos := 1

	for ; data[pos] != '\r'; pos++ {
	}

	return string(data[1:pos]), pos + 2, nil
}

// read a RESP encoded error from data and returns
// the error string, delta and the error
func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// read a RESP encoded integer from data and returns
// the integer, delta and the error
func readInt64(data []byte) (int64, int, error) {
	var value int64
	pos := 1

	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}

	return value, pos + 2, nil

}

// read a RESP encoded string from data and returns
// the integer, delta and the error
func readBulkString(data []byte) (string, int, error) {
	pos := 1

	//Reading len and forwarding postion by the length of the integer + the special character
	len, delta := readLength(data[pos:])
	pos += delta

	return string(data[pos:(pos + len)]), pos + len + 2, nil
}

// read a RESP encoded string from data and returns
// the integer, delta and the error
func readArray(data []byte) ([]any, int, error) {
	pos := 1
	count, delta := readLength(data[pos:])
	pos += delta

	elems := make([]any, count)
	for i := range elems {
		elem, delta, err := DecodeOne(data[pos:])

		if err != nil {
			return []any{}, 0, err
		}

		elems[i] = elem
		pos += delta
	}

	return elems, pos + 2, nil
}

// reads the length typically the first integer of the string
// until hit a non digit byte then returns
// the integer and the delta = length + 2
func readLength(data []byte) (int, int) {
	length, pos := 0, 0

	for pos = range data {
		b := data[pos]

		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}

		length = length*10 + int(b-'0')
	}

	return 0, 0
}

func Encode(value any, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}

		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	}

	return []byte{}
}
