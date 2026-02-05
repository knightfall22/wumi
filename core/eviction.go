package core

import (
	"time"

	"github.com/knightfall22/wumi/config"
)

/*
* The approximated LRU algorithm
 */
func getCurrentClock() uint32 {
	return uint32(time.Now().UnixMilli() & 0x00FFFFFF)
}

func getIdleTime(lastAccessedAt uint32) uint32 {
	c := getCurrentClock()

	if c >= lastAccessedAt {
		return c - lastAccessedAt
	}

	return uint32(0x00FFFFFF-lastAccessedAt) + c
}

func populateEvictionPool() {
	sampleSize := 5

	for k := range store {
		ePool.Push(k, store[k].LastAccessedAt)
		sampleSize--
		if sampleSize == 0 {
			break
		}
	}
}

func evictAllKeysLRU() {
	populateEvictionPool()
	evictionCount := int16(config.EvictionRatio * float64(config.KeyLimit))

	for i := 0; i < int(evictionCount) && len(ePool.pool) > 0; i++ {
		item := ePool.Pop()
		if item == nil {
			return
		}

		DEL(item.key)
	}
}

// Randomly removes key to makes space for new data added
// The number of keys removed will be sufficient to free up at least 10% of space
func evictAllKeysRandom() {
	evictionCount := int64(config.EvictionRatio * float64(config.KeyLimit))

	for k := range store {
		DEL(k)
		evictionCount--

		if evictionCount <= 0 {
			break
		}
	}
}

// evict the first key it found while iterating
func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

func evict() {
	switch config.EvictionStrategy {
	case "simple-first":
		evictFirst()
	case "allkeys-random":
		evictAllKeysRandom()
	case "allkeys-lru":
		evictAllKeysLRU()
	}
}
