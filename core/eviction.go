package core

// evict the first key it found while iterating
func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

func evict() {
	evictFirst()
}
