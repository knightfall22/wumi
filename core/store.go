package core

import (
	"time"

	"github.com/knightfall22/wumi/config"
)

var store = make(map[string]*Obj)

func NewObj(v any, durationMs int64, oType, oEnc uint8) *Obj {
	var expiresAt int64 = -1
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	return &Obj{
		Value:        v,
		TypeEncoding: oType | oEnc,
		ExpiresAt:    expiresAt,
	}
}

func PUT(k string, object *Obj) {
	if len(store) >= config.KeyLimit {
		evict()
	}
	store[k] = object
}

func GET(k string) *Obj {
	v := store[k]

	if v != nil {
		if v.ExpiresAt != -1 && v.ExpiresAt <= time.Now().UnixMilli() {
			delete(store, k)
			return nil
		}
	}
	return v
}

func DEL(k string) bool {
	if _, ok := store[k]; ok {
		delete(store, k)
		return true
	}
	return false
}
