package core

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_MINUS_ONE []byte = []byte(":-1\r\n")
var RESP_MINUS_TWO []byte = []byte(":-2\r\n")

func evalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b
}

func evalSET(args []string) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'set' command"), false)
	}

	var expireMs int64 = -1
	key, value := args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++

			if i == len(args) {
				return Encode(errors.New("(error) ERR syntax error"), false)
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
			}

			expireMs = exDurationSec * 1000

		default:
			return Encode(errors.New("ERR syntax error"), false)

		}
	}

	PUT(key, NewObj(value, expireMs))
	return RESP_OK
}

func evalGet(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'get' command"), false)
	}

	key := args[0]

	//Get the key from hash table
	obj := GET(key)

	//if key does not exist return RESP encoded nil
	if obj == nil {
		return RESP_NIL
	}

	//if the key as already expired return nil
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		return RESP_NIL
	}

	// return RESP encoded values
	return Encode(obj.Value, false)
}

func evalTTL(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'ttl' command"), false)
	}

	key := args[0]

	//Get the key from hash table
	obj := GET(key)

	// if key doesn't exist return RESP encoded -2
	if obj == nil {
		return RESP_MINUS_TWO
	}

	// if key does exist but ttl is not set return RESP encoded -1
	if obj.ExpiresAt == -1 {
		return RESP_MINUS_ONE
	}

	//compute the time remaining for the key to expire and
	//return the RESP encodedd form of it
	durationMs := obj.ExpiresAt - time.Now().UnixMilli()

	//if the key is expired return -2
	if durationMs <= 0 {
		return RESP_MINUS_TWO
	}

	return (Encode(int64(durationMs/1000), false))
}

func evalDel(args []string) []byte {
	deletedCount := 0

	for i := range args {
		if exist := DEL(args[i]); exist {
			deletedCount++
		}
	}

	return Encode(int64(deletedCount), false)

}

func evalExpire(args []string) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'expire' command"), false)
	}

	key := args[0]

	//Get the key from hash table
	obj := GET(key)

	//0 if timeout was not set. e.g. key does not exist, or operation skipped due to provided argument
	if obj == nil {
		return RESP_ZERO
	}

	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("(error) ERR value is not an integer or out of range"), false)
	}

	expireMs := exDurationSec * 1000

	obj.ExpiresAt = expireMs + time.Now().UnixMilli()

	return RESP_ONE
}

func EvalAndRespond(cmds RedisCmds, c io.ReadWriter) error {
	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPING(cmd.Args))
		case "SET":
			buf.Write(evalSET(cmd.Args))
		case "GET":
			buf.Write(evalGet(cmd.Args))
		case "TTL":
			buf.Write(evalTTL(cmd.Args))
		case "DEL":
			buf.Write(evalDel(cmd.Args))
		case "EX", "EXPIRE":
			buf.Write(evalExpire(cmd.Args))
		default:
			buf.Write(evalPING(cmd.Args))
		}
	}

	_, err := c.Write(buf.Bytes())
	return err
}
