package inmemory_cache

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/thoas/go-funk"
	"time"
)

type Service struct {
	cache      *cache.Cache
	expiration time.Duration
}

func New(expiration time.Duration) *Service {
	return &Service{
		expiration: expiration,
		cache:      cache.New(expiration, expiration),
	}
}

func (s *Service) Get(ids []int64) (map[int64]interface{}, []int64) {
	var missingData []int64

	result := map[int64]interface{}{}
	for _, id := range ids {
		cached, ok := s.cache.Get(fmt.Sprint(id))
		if !ok {
			if !funk.ContainsInt64(missingData, id) {
				missingData = append(missingData, id)
			}
			continue
		}

		result[id] = cached
	}

	return result, missingData
}

func (s *Service) Set(data map[int64]interface{}, expiration time.Duration) {
	for key, iface := range data {
		s.cache.Set(fmt.Sprint(key), iface, expiration)
	}
}
