package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

const DefaultExpiration = cache.DefaultExpiration

type ICache interface {
	SetByInt64s(data map[int64]interface{}, expiration time.Duration)
	SetByString(key string, value interface{}, expiration time.Duration)
	GetByInt64s(keys []int64) (results map[int64]interface{}, missing []int64)
	GetByStrings(keys []string) (results map[string]interface{}, missing []string)
}

type NoopCache struct {
}

func (NoopCache) SetByInt64s(data map[int64]interface{}, expiration time.Duration) {
	return
}

func (NoopCache) SetByString(key string, value interface{}, expiration time.Duration) {
	return
}

func (NoopCache) GetByInt64s(keys []int64) (results map[int64]interface{}, missing []int64) {
	return map[int64]interface{}{}, keys
}

func (NoopCache) GetByStrings(keys []string) (results map[string]interface{}, missing []string) {
	return map[string]interface{}{}, keys
}
