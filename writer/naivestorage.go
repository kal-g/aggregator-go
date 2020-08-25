package aggregator

import "sync"

type NaiveStorage struct {
	data map[string]int
	mtx  *sync.Mutex
}

func NewNaiveStorage() *NaiveStorage {
	var ns NaiveStorage
	ns.data = make(map[string]int)
	ns.mtx = &sync.Mutex{}
	return &ns
}

func (s NaiveStorage) Get(key string) StorageResult {
	val, key_exists := s.data[key]
	if !key_exists {
		return StorageResult{Value: 0, ErrCode: 1}
	}
	return StorageResult{Value: val, ErrCode: 0}
}

func (s NaiveStorage) Put(key string, val int) {
	s.data[key] = val
}

func (s NaiveStorage) Lock() {
	s.mtx.Lock()
}

func (s NaiveStorage) Unlock() {
	s.mtx.Unlock()
}
