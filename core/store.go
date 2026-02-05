package core

import (
	"time"

	"github.com/knightfall22/wumi/config"
)

var store = make(map[string]*Obj)
var expires = make(map[*Obj]uint64)

func setExpiry(obj *Obj, expDurationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli() + expDurationMs)
}

func getExpiry(obj *Obj) (int64, bool) {
	exp, ok := expires[obj]
	return int64(exp), ok
}

func hasExpired(obj *Obj) bool {
	v, ok := expires[obj]

	if !ok {
		return false
	}

	return v <= uint64(time.Now().UnixMilli())
}

func NewObj(v any, durationMs int64, oType, oEnc uint8) *Obj {
	obj := &Obj{
		Value:          v,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}

	if durationMs > 0 {
		setExpiry(obj, durationMs)
	}

	return obj
}

func PUT(k string, object *Obj) {
	if len(store) >= config.KeyLimit {
		evict()
	}

	object.LastAccessedAt = getCurrentClock()
	store[k] = object

	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}

	KeyspaceStat[0]["keys"]++
}

func GET(k string) *Obj {
	v := store[k]

	if v != nil {
		if hasExpired(v) {
			delete(store, k)
			return nil
		}
	}
	v.LastAccessedAt = getCurrentClock()
	return v
}

func DEL(k string) bool {
	if v, ok := store[k]; ok {
		delete(store, k)
		delete(expires, v)
		KeyspaceStat[0]["keys"]--
		return true
	}
	return false
}
