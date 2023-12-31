package notification_handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
)

func NewNotificationHandlerWrapper(config boilerplate.WrapperConfig) INotificationHandlerWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://notification-handler"

		log.Warn().Msgf("Api Url is missing for NotificationHandler. Setting as default : %v", config.ApiUrl)
	}

	w := &NotificationHandlerWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "notification-handler",
	}

	env := boilerplate.GetCurrentEnvironment().ToString()

	w.publisher = eventsourcing.NewKafkaEventPublisher(
		boilerplate.KafkaWriterConfiguration{
			Hosts: "kafka-0.kafka-headless.kafka.svc.cluster.local:9092,kafka-1.kafka-headless.kafka.svc.cluster.local:9092",
			Tls:   false,
		}, boilerplate.KafkaTopicConfig{
			Name:              fmt.Sprintf("%v.notifications.handler_sending_queue", env),
			NumPartitions:     24,
			ReplicationFactor: 2,
		})

	w.customPublisher = eventsourcing.NewKafkaEventPublisher(
		boilerplate.KafkaWriterConfiguration{
			Hosts: "kafka-0.kafka-headless.kafka.svc.cluster.local:9092,kafka-1.kafka-headless.kafka.svc.cluster.local:9092",
			Tls:   false,
		}, boilerplate.KafkaTopicConfig{
			Name:              fmt.Sprintf("%v.notifications.handler_sending_queue_custom", env),
			NumPartitions:     24,
			ReplicationFactor: 2,
		})

	return w
}

func (h *NotificationHandlerWrapper) EnqueueNotificationWithTemplate(templateName string, userId int64,
	renderingVars map[string]string, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult {
	ch := make(chan EnqueueMessageResult, 2)

	go func() {
		var resp EnqueueMessageResult

		defer func() {
			ch <- resp
			close(ch)
		}()

		templateName = strings.ToLower(templateName)

		if len(templateName) == 0 {
			resp.Error = errors.New("invalid template name")
			return
		}

		if h.publisher == nil {
			resp.Error = errors.New("publisher is nil")
			return
		}

		id := boilerplate.GetGenerator().Generate().String()

		if err := h.publisher.Publish(apm.TransactionFromContext(ctx),
			SendNotificationWithTemplate{
				Id:                 id,
				TemplateName:       templateName,
				RenderingVariables: renderingVars,
				UserId:             userId,
				CustomData:         customData,
			}); len(err) > 0 {
			resp.Error = err[0]

			return
		}

		resp.Id = id
	}()

	return ch
}

func (h *NotificationHandlerWrapper) EnqueueNotificationWithCustomTemplate(title, body, headline string, userId int64, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult {
	ch := make(chan EnqueueMessageResult, 2)

	go func() {
		var resp EnqueueMessageResult

		defer func() {
			ch <- resp
			close(ch)
		}()

		if len(title) == 0 && len(body) == 0 && len(headline) == 0 {
			resp.Error = errors.New("message is empty")
			return
		}

		if h.publisher == nil {
			resp.Error = errors.New("publisher is nil")
			return
		}

		id := boilerplate.GetGenerator().Generate().String()

		if err := h.customPublisher.Publish(apm.TransactionFromContext(ctx),
			SendNotificationWithCustomTemplate{
				Id:         id,
				UserId:     userId,
				Title:      title,
				Body:       body,
				Headline:   headline,
				CustomData: customData,
			}); len(err) > 0 {
			resp.Error = err[0]

			return
		}
		resp.Id = id
	}()

	return ch
}

func (h *NotificationHandlerWrapper) GetNotificationsReadCount(notificationIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]int64] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]int64](h.baseWrapper, h.apiUrl,
		"GetNotificationsReadCount", GetNotificationsReadCountRequest{NotificationIds: notificationIds},
		map[string]string{}, h.defaultTimeout, apm.TransactionFromContext(ctx), h.serviceName, forceLog)
}

func (h *NotificationHandlerWrapper) DisableUnregisteredTokens(tokens []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[[]string] {
	return wrappers.ExecuteRpcRequestAsync[[]string](h.baseWrapper, h.apiUrl,
		"DisableUnregisteredTokens", DisableUnregisteredTokensRequest{Tokens: tokens},
		map[string]string{}, h.defaultTimeout, apm.TransactionFromContext(ctx), h.serviceName, forceLog)
}
