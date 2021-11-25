package content

import (
	"context"
	"encoding/json"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/go-common/kafka_listener/structs"
	"github.com/segmentio/kafka-go"
	"time"
)

func (w *ContentWrapper) InitKafkaListener(kafkaConfig boilerplate.KafkaListenerConfiguration, ctx context.Context) {
	w.l = kafka_listener.NewBatchListener(kafkaConfig, structs.NewCommand("user_cache_service",
		func(executionData structs.ExecutionData, request ...kafka.Message) (interface{}, error) {
			mapped, appErrors := w.mapKafkaMessages(request)

			if len(appErrors) > 0 {
				if len(appErrors) == len(request) {
					return nil, appErrors[0]
				}
			}

			for _, e := range mapped {
				w.cache.SetInt64Key(e.Id, e, w.defaultExpiration)
			}

			return nil, nil
		}, false), ctx, time.Duration(kafkaConfig.ListenerDuration)*time.Second, kafkaConfig.MaxBatchSize)

	go func() {
		w.l.Listen()
	}()
}

func (w *ContentWrapper) mapKafkaMessages(messages []kafka.Message) ([]SimpleContent, []error) {
	eventsMap := make([]SimpleContent, len(messages))
	var appErrors []error

	for i, message := range messages {
		var event SimpleContent

		if err := json.Unmarshal(message.Value, &event); err != nil {
			appErrors = append(appErrors, err)
			continue
		}

		eventsMap[i] = event
	}

	return eventsMap, appErrors
}
