package content

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/cache/inmemory_cache"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

type SimpleContent struct {
	Id            int64    `json:"id"`
	Duration      int      `json:"duration"`
	AgeRestricted bool     `json:"age_restricted"`
	AuthorId      int64    `json:"author_id"`
	CategoryId    null.Int `json:"category_id"`
	Hashtags      []string `json:"hashtags"`
}

//goland:noinspection ALL
type ContentGetInternalResponseChan struct {
	Error *rpc.RpcError           `json:"error"`
	Items map[int64]SimpleContent `json:"items"`
}

type ContentGetInternalRequest struct {
	IncludeDeleted bool    `json:"include_deleted"`
	ContentIds     []int64 `json:"content_ids"`
}

type IContentWrapper interface {
	GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type ContentWrapper struct {
	baseWrapper       *wrappers.BaseWrapper
	defaultTimeout    time.Duration
	defaultExpiration time.Duration
	cache             *inmemory_cache.Service
	l                 *kafka_listener.BatchListener
	apiUrl            string
	serviceName       string
}

func NewContentWrapper(apiUrl string, cacheExpiration time.Duration, kafkaConfig boilerplate.KafkaListenerConfiguration, ctx context.Context) IContentWrapper {
	w := &ContentWrapper{
		baseWrapper:       wrappers.GetBaseWrapper(),
		defaultTimeout:    5 * time.Second,
		apiUrl:            common.StripSlashFromUrl(apiUrl),
		defaultExpiration: cacheExpiration,
		serviceName:       "content-backend",
		cache:             inmemory_cache.New(cacheExpiration),
	}

	w.InitKafkaListener(kafkaConfig, ctx)

	return w
}

func (w *ContentWrapper) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction,
	forceLog bool) chan ContentGetInternalResponseChan {
	respCh := make(chan ContentGetInternalResponseChan, 2)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		finalResponse := map[int64]SimpleContent{}

		cachedContent, missingInCache := w.cache.Get(contentIds)
		for id, record := range cachedContent {
			content, ok := record.(SimpleContent)
			if !ok {
				apm_helper.CaptureApmError(errors.New("cannot convert interface from cache"), apmTransaction)
				continue
			}

			finalResponse[id] = content
		}

		if len(missingInCache) == 0 {
			respCh <- ContentGetInternalResponseChan{
				Error: nil,
				Items: finalResponse,
			}

			return
		}

		respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "ContentGetInternal", ContentGetInternalRequest{
			ContentIds:     missingInCache,
			IncludeDeleted: includeDeleted,
		}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		resp := <-respChan

		result := ContentGetInternalResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			items := map[int64]SimpleContent{}

			if err := json.Unmarshal(resp.Result, &items); err != nil {
				result.Error = &rpc.RpcError{
					Code:    error_codes.GenericMappingError,
					Message: err.Error(),
					Data:    nil,
				}
			} else {
				for k, v := range items {
					w.cache.SetInt64Key(k, v, w.defaultExpiration)
					finalResponse[k] = v
				}

				result.Items = finalResponse
			}
		}

		respCh <- result
	})

	return respCh
}
