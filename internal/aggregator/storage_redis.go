package aggregator

import (
	"context"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	db  *redis.Client
	mtx *sync.Mutex
}

func NewRedisStorage(redisURL string) *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	rs := &RedisStorage{
		db:  rdb,
		mtx: &sync.Mutex{},
	}
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	// TODO Only if new master, clean out the DB
	rdb.FlushAll(ctx)
	return rs
}

// Get for storage
func (s RedisStorage) Get(key string) StorageResult {
	ctx := context.Background()
	val, err := s.db.Get(ctx, key).Result()
	if err != nil {
		return StorageResult{Value: 0, Err: &StorageKeyNotFoundError{}}
	}

	if len(val) == 0 {
		return StorageResult{Value: 0, Err: &StorageKeyNotFoundError{}}
	}
	value, err := strconv.Atoi(val)
	if err != nil {
		return StorageResult{Value: 0, Err: &StorageInvalidDataError{}}
	}
	return StorageResult{Value: value, Err: nil}
}

// IncrBy value
func (s RedisStorage) IncrBy(key string, incr int) {

	ctx := context.Background()
	res := s.db.IncrBy(ctx, key, int64(incr))
	err := res.Err()
	if err != nil {
		panic(err)
	}
}

// Lock storage
func (s RedisStorage) Lock(string) {
	s.mtx.Lock()
}

// Unlock storage
func (s RedisStorage) Unlock(string) {
	s.mtx.Unlock()
}
