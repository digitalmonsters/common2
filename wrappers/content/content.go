package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/cache"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
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
	dataCache         cache.ICache
	apiUrl            string
	serviceName       string
}

func NewContentWrapper(apiUrl string, dataCache cache.ICache) IContentWrapper {
	w := &ContentWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: 5 * time.Second,
		apiUrl:         common.StripSlashFromUrl(apiUrl),
		serviceName:    "content-backend",
		dataCache:      dataCache,
	}

	if w.dataCache == nil {
		w.dataCache = cache.NoopCache{}
	}

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

		cachedContent, missingInCache := w.dataCache.GetByInt64s(contentIds)
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
					w.dataCache.SetByString(fmt.Sprint(k), v, cache.DefaultExpiration)

					finalResponse[k] = v
				}

				result.Items = finalResponse
			}
		}

		respCh <- result
	})

	return respCh
}
