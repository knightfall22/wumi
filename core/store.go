package core

import (
	"time"
)

type Obj struct {
	Value     any
	ExpiresAt int64
}

var store = make(map[string]*Obj)

func NewObj(v any, durationMs int64) *Obj {
	var expiresAt int64 = -1
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	return &Obj{
		Value:     v,
		ExpiresAt: expiresAt,
	}
}

func PUT(k string, object *Obj) {
	store[k] = object
}

func GET(k string) *Obj {
	return store[k]
}
