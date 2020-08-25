package aggregator

import (
	"strconv"
	"sync"

	"github.com/tecbot/gorocksdb"
)

type RocksDBStorage struct {
	db  *gorocksdb.DB
	mtx *sync.Mutex
}

func NewRocksDBStorage(path string) *RocksDBStorage {
	var rdb RocksDBStorage

	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)

	db, err := gorocksdb.OpenDb(opts, path)
	if err != nil {
		panic(err)
	}
	rdb.db = db
	rdb.mtx = &sync.Mutex{}
	return &rdb
}

func (s RocksDBStorage) Get(key string) StorageResult {
	val, err := s.db.Get(gorocksdb.NewDefaultReadOptions(), []byte(key))
	if err != nil {
		return StorageResult{Value: 0, ErrCode: 1}
	}

	data := val.Data()
	value, err := strconv.Atoi(string(data))
	if err != nil {
		return StorageResult{Value: 0, ErrCode: 2}
	}
	return StorageResult{Value: value, ErrCode: 0}
}

func (s RocksDBStorage) Put(key string, val int) {
	s.db.Put(gorocksdb.NewDefaultWriteOptions(), []byte(key), []byte(strconv.Itoa(val)))
}

func (s RocksDBStorage) Lock() {
	s.mtx.Lock()
}

func (s RocksDBStorage) Unlock() {
	s.mtx.Unlock()
}
