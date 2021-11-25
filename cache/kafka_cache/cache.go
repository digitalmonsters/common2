package kafka_cache

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/go-common/kafka_listener/structs"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/segmentio/kafka-go"
	"github.com/thoas/go-funk"
	"time"
)

type MapKafkaMessageFn func(message kafka.Message) string

type KafkaCache struct {
	memoryCache    *cache.Cache
	listenerConfig boilerplate.KafkaListenerConfiguration
	ctx            context.Context
	mapFn          MapKafkaMessageFn
}

func NewKafkaCache(listenerConfig boilerplate.KafkaListenerConfiguration, ctx context.Context,
	mapFn MapKafkaMessageFn) *KafkaCache {
	listenerConfig.GroupId = uuid.New().String()

	return &KafkaCache{listenerConfig: listenerConfig, ctx: ctx, mapFn: mapFn,
		memoryCache: cache.New(15*time.Minute, 20*time.Minute)}
}

func (k *KafkaCache) startListener() {
	l := kafka_listener.NewBatchListener(k.listenerConfig, structs.NewCommand(
		"cache listener", func(executionData structs.ExecutionData, request ...kafka.Message) (interface{}, error) {
			for _, r := range request {
				k.memoryCache.Delete(k.mapFn(r))
			}
			return nil, nil
		}, false), k.ctx, 5*time.Second, 100)

	go func() {
		l.Listen()
	}()
}

func (k *KafkaCache) GetByStrings(keys []string) (results map[string]interface{}, missing []string) {
	results = map[string]interface{}{}

	for _, id := range keys {
		cached, ok := k.memoryCache.Get(id)
		if !ok {
			if !funk.ContainsString(missing, id) {
				missing = append(missing, id)
			}
			continue
		}

		results[id] = cached
	}

	return results, missing
}

func (k *KafkaCache) GetByInt64s(keys []int64) (results map[int64]interface{}, missing []int64) {
	results = map[int64]interface{}{}

	for _, id := range keys {
		cached, ok := k.memoryCache.Get(fmt.Sprint(id))
		if !ok {
			if !funk.ContainsInt64(missing, id) {
				missing = append(missing, id)
			}
			continue
		}

		results[id] = cached
	}

	return results, missing
}


func (k *KafkaCache) SetByInt64s(data map[int64]interface{}, expiration time.Duration) {
	for key, value := range data {
		k.SetByString(key, value, expiration)
	}
}

func (k *KafkaCache) SetByString(key int64, value interface{}, expiration time.Duration) {
	k.memoryCache.Set(fmt.Sprint(key), value, expiration)
}
