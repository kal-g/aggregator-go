package aggregator

import (
	"strconv"
	"sync"

	"github.com/tecbot/gorocksdb"
)

// RocksDBStorage with concurrency
type RocksDBStorage struct {
	db     *gorocksdb.DB
	mtx    *sync.Mutex
	mtxMap map[string]*sync.Mutex
}

// NewRocksDBStorage creates and inits a new rocksdb instance. Must not already exist
func NewRocksDBStorage(path string) *RocksDBStorage {
	var rdb RocksDBStorage

	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	opts.SetErrorIfExists(true)

	db, err := gorocksdb.OpenDb(opts, path)
	if err != nil {
		panic(err)
	}
	rdb.db = db
	rdb.mtx = &sync.Mutex{}
	rdb.mtxMap = make(map[string]*sync.Mutex)
	return &rdb
}

// Get for storage
func (s RocksDBStorage) Get(key string) StorageResult {
	val, err := s.db.Get(gorocksdb.NewDefaultReadOptions(), []byte(key))
	if err != nil {
		return StorageResult{Value: 0, ErrCode: 1}
	}

	data := val.Data()
	if len(data) == 0 {
		return StorageResult{Value: 0, ErrCode: 1}
	}
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return StorageResult{Value: 0, ErrCode: 2}
	}
	return StorageResult{Value: value, ErrCode: 0}
}

// Put for storage
func (s RocksDBStorage) Put(key string, val int) {
	s.db.Put(gorocksdb.NewDefaultWriteOptions(), []byte(key), []byte(strconv.Itoa(val)))
}

// Lock for storage per namespace
func (s RocksDBStorage) Lock(namespace string) {
	mtx, mtxExists := s.mtxMap[namespace]
	if !mtxExists {
		s.mtx.Lock()
		s.mtxMap[namespace] = &sync.Mutex{}
		s.mtxMap[namespace].Lock()
		s.mtx.Unlock()
	} else {
		mtx.Lock()
	}

}

// Unlock for storage per namespace
func (s RocksDBStorage) Unlock(namespace string) {
	s.mtxMap[namespace].Unlock()
}
