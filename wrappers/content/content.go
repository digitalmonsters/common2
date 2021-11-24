package content

import (
	"encoding/json"
	"errors"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/cache/inmemory_cache"
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
type ContentGetInternalResponse struct {
	Error *rpc.RpcError           `json:"error"`
	Items map[int64]SimpleContent `json:"items"`
}

type ContentGetInternalRequest struct {
	IncludeDeleted bool    `json:"include_deleted"`
	ContentIds     []int64 `json:"content_ids"`
}

type IContentWrapper interface {
	GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponse
}

//goland:noinspection GoNameStartsWithPackageName
type ContentWrapper struct {
	baseWrapper       *wrappers.BaseWrapper
	defaultTimeout    time.Duration
	defaultExpiration time.Duration
	cache             *inmemory_cache.Service
	apiUrl            string
	serviceName       string
}

func NewContentWrapper(apiUrl string, cacheExpiration time.Duration) IContentWrapper {
	return &ContentWrapper{
		baseWrapper:       wrappers.GetBaseWrapper(),
		defaultTimeout:    5 * time.Second,
		apiUrl:            apiUrl,
		defaultExpiration: cacheExpiration,
		serviceName:       "content-backend",
		cache:             inmemory_cache.New(cacheExpiration),
	}
}

func (w *ContentWrapper) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponse {
	respCh := make(chan ContentGetInternalResponse, 2)

	finalResponse := map[int64]SimpleContent{}

	cachedContent, missingInCache := w.cache.Get(contentIds)
	for id, iface := range cachedContent {
		content, ok := iface.(SimpleContent)
		if !ok {
			apm_helper.CaptureApmError(errors.New("cannot convert interface from cache"), apmTransaction)
			continue
		}

		finalResponse[id] = content
	}

	if len(missingInCache) == 0 {
		respCh <- ContentGetInternalResponse{
			Error: nil,
			Items: finalResponse,
		}

		close(respCh)
		return respCh
	}

	respChan := w.baseWrapper.SendRequest(w.apiUrl, "ContentGetInternal", ContentGetInternalRequest{
		ContentIds:     contentIds,
		IncludeDeleted: includeDeleted,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := ContentGetInternalResponse{
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
				toCache := map[int64]interface{}{}
				for id, content := range items {
					finalResponse[id] = content
					toCache[id] = content
				}

				w.cache.Set(toCache, w.defaultExpiration)
				result.Items = finalResponse
			}
		}

		respCh <- result
	})

	return respCh
}
