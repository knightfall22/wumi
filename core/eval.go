package core

import (
	"errors"
	"io"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")

func evalPING(args []string, c io.ReadWriter) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func evalSET(args []string, c io.ReadWriter) error {
	if len(args) <= 1 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}

	var expireMs int64 = -1
	key, value := args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++

			if i == len(args) {
				return errors.New("(error) ERR syntax error")
			}

			exDurationSec, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return errors.New("(error) ERR value is not an integer or out of range")
			}

			expireMs = exDurationSec * 1000

		default:
			return errors.New("ERR syntax error")

		}
	}

	PUT(key, NewObj(value, expireMs))
	_, err := c.Write([]byte("+OK\r\n"))
	return err
}

func evalGet(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'get' command")
	}

	key := args[0]

	//Get the key from hash table
	obj := GET(key)

	//if key does not exist return RESP encoded nil
	if obj == nil {
		_, err := c.Write(RESP_NIL)
		return err
	}

	//if the key as already expired return nil
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		_, err := c.Write(RESP_NIL)
		return err
	}

	// return RESP encoded values
	_, err := c.Write(Encode(obj.Value, false))
	return err
}

func evalTTL(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'ttl' command")
	}

	key := args[0]

	//Get the key from hash table
	obj := GET(key)

	// if key doesn't exist return RESP encoded -2
	if obj == nil {
		_, err := c.Write([]byte(":-2\r\n"))
		return err
	}

	// if key does exist but ttl is not set return RESP encoded -1
	if obj.ExpiresAt == -1 {
		_, err := c.Write([]byte(":-1\r\n"))
		return err
	}

	//compute the time remaining for the key to expire and
	//return the RESP encodedd form of it
	durationMs := obj.ExpiresAt - time.Now().UnixMilli()

	//if the key is expired return -2
	if durationMs <= 0 {
		_, err := c.Write([]byte(":-2\r\n"))
		return err
	}

	_, err := c.Write(Encode(int64(durationMs/1000), false))
	return err
}

func EvalAndRespond(cmd *RedisCmd, c io.ReadWriter) error {
	switch cmd.Cmd {
	case "PING":
		return evalPING(cmd.Args, c)
	case "SET":
		return evalSET(cmd.Args, c)
	case "GET":
		return evalGet(cmd.Args, c)
	case "TTL":
		return evalTTL(cmd.Args, c)
	default:
		return evalPING(cmd.Args, c)
	}
}
