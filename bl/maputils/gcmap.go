package maputils

type GCMap[K comparable, V any] struct {
	m        map[K]V
	accessed map[K]struct{}
}

// NewGC Creates a map that can GC unaccessed items
func NewGC[K comparable, V any](m map[K]V) GCMap[K, V] {
	return GCMap[K, V]{
		m:        m,
		accessed: map[K]struct{}{},
	}
}

// Looks up an item from the map
func (ngc GCMap[K, V]) Lookup(k K) (V, bool) {
	val, ok := ngc.m[k]

	ngc.accessed[k] = struct{}{}

	return val, ok
}

// Returns only the items in the map that were looked up.
func (ngc GCMap[K, V]) PurgeUnaccessed() map[K]V {
	lookedUp := map[K]V{}
	for k, v := range ngc.m {
		if _, ok := ngc.accessed[k]; ok {
			lookedUp[k] = v
		}
	}

	return lookedUp
}

func (ngc *GCMap[K, V]) GetUnaccessed() map[K]V {
	unaccessed := map[K]V{}

	for k, v := range ngc.m {
		if _, ok := ngc.accessed[k]; !ok {
			unaccessed[k] = v
		}
	}
	return unaccessed
}
