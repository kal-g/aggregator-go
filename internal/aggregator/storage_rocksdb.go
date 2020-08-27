package aggregator

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/tecbot/gorocksdb"
)

type rocksDBStorage struct {
	db     *gorocksdb.DB
	mtx    *sync.Mutex
	mtxMap map[string]*sync.Mutex
}

func newRocksDBStorage(path string) *rocksDBStorage {
	var rdb rocksDBStorage

	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	fmt.Printf("Path: %+v\n", path)
	db, err := gorocksdb.OpenDb(opts, path)
	if err != nil {
		panic(err)
	}
	rdb.db = db
	rdb.mtx = &sync.Mutex{}
	rdb.mtxMap = make(map[string]*sync.Mutex)
	return &rdb
}

func (s rocksDBStorage) Get(key string) StorageResult {
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

func (s rocksDBStorage) Put(key string, val int) {
	s.db.Put(gorocksdb.NewDefaultWriteOptions(), []byte(key), []byte(strconv.Itoa(val)))
}

func (s rocksDBStorage) Lock(namespace string) {
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

func (s rocksDBStorage) Unlock(namespace string) {
	s.mtxMap[namespace].Unlock()
}
