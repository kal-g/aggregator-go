package aggregator

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	db *redis.Client
}

func NewRedisStorage(addr string) *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// TODO If new master, clean out the DB
	rs := &RedisStorage{
		db: rdb,
	}
	return rs
}

// Get for storage
func (s RedisStorage) Get(key string) StorageResult {
	ctx := context.Background()
	val, err := s.db.Get(ctx, key).Result()
	if err != nil {
		return StorageResult{Value: 0, ErrCode: 1}
	}

	if len(val) == 0 {
		return StorageResult{Value: 0, ErrCode: 1}
	}
	value, err := strconv.Atoi(val)
	if err != nil {
		return StorageResult{Value: 0, ErrCode: 2}
	}
	return StorageResult{Value: value, ErrCode: 0}
}
