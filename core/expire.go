package core

import (
	"log"
	"time"
)

func expireSample() float32 {
	limit := 20
	expiredCount := 0

	// assuming iteration of golang hash table is randomized
	for key, obj := range store {
		if obj.ExpiresAt != -1 {
			limit--

			if obj.ExpiresAt <= time.Now().UnixMilli() {
				delete(store, key)
				expiredCount++
			}
		}

		//once we hace iterated the keys that have expiration set
		//we break
		if limit == 0 {
			break
		}
	}

	return float32(expiredCount) / float32(20.0)

}

func DeleteExpiredKeys() {
	for {
		frac := expireSample()

		//if the sample had less than 25% of sampled key
		//breakt the loop
		if frac < .25 {
			break
		}
	}

	log.Println("deleted the expired but undeleted keys. total keys", len(store))
}
