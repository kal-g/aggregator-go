package aggregator

import "sync"

// NaiveStorage uses a simple go map
// For testing only
type naiveStorage struct {
	data map[string]int
	mtx  *sync.Mutex
}

// NewNaiveStorage create storage
func newNaiveStorage() *naiveStorage {
	var ns naiveStorage
	ns.data = make(map[string]int)
	ns.mtx = &sync.Mutex{}
	return &ns
}

// Get value
func (s naiveStorage) Get(key string) StorageResult {
	val, keyExists := s.data[key]
	if !keyExists {
		return StorageResult{Value: 0, ErrCode: 1}
	}
	return StorageResult{Value: val, ErrCode: 0}
}

// IncrBy value
func (s naiveStorage) IncrBy(key string, incr int) {
	_, keyExists := s.data[key]
	if !keyExists {
		s.data[key] = 0
	}
	s.data[key] = s.data[key] + incr
}

// Lock storage
func (s naiveStorage) Lock(string) {
	s.mtx.Lock()
}

// Unlock storage
func (s naiveStorage) Unlock(string) {
	s.mtx.Unlock()
}
