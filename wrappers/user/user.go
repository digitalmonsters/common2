package user

import (
	"context"
	"encoding/json"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/cache/inmemory_cache"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"time"
)

type Wrapper struct {
	apiUrl            string
	defaultExpiration time.Duration
	baseWrapper       *wrappers.BaseWrapper
	cache             *inmemory_cache.Service
	ctx               context.Context
	l                 *kafka_listener.BatchListener
}

type IUserWrapper interface {
	GetCachedUsers(userIds []int64, apmTransaction *apm.Transaction) chan CachedUsersResponse
}

func New(apiUrl string, cacheDefaultExp time.Duration, kafkaConfig boilerplate.KafkaListenerConfiguration, ctx context.Context) IUserWrapper {
	w := &Wrapper{
		apiUrl:            apiUrl,
		ctx:               ctx,
		defaultExpiration: cacheDefaultExp,
		baseWrapper:       wrappers.GetBaseWrapper(),
		cache:             inmemory_cache.New(cacheDefaultExp),
	}

	w.InitKafkaListener(kafkaConfig, ctx)

	return w
}

func (w *Wrapper) GetCachedUsers(userIds []int64, apmTransaction *apm.Transaction) chan CachedUsersResponse {
	respCh := make(chan CachedUsersResponse, 2)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		finalResponse := map[int64]SimpleUser{}

		cachedUsers, missingInCache := w.cache.Get(userIds)
		for id, iface := range cachedUsers {
			content, ok := iface.(SimpleUser)
			if !ok {
				apm_helper.CaptureApmError(errors.New("cannot convert interface from cache"), apmTransaction)
				continue
			}

			finalResponse[id] = content
		}

		if len(missingInCache) == 0 {
			respCh <- CachedUsersResponse{
				Error: nil,
				Items: finalResponse,
			}

			return
		}

		result := CachedUsersResponse{
			Error: nil,
		}

		//todo: api request logic
		var body []byte

		req := &fasthttp.Request{}
		req.SetRequestURI(w.apiUrl)
		req.Header.SetMethod("POST")
		req.SetBody(body)

		var resp *fasthttp.Response

		err := apm_helper.SendHttpRequest(nil, req, resp, apmTransaction, time.Minute, true)
		if err != nil {
			result.Error = &rpc.RpcError{
				Code:    error_codes.GenericServerError,
				Message: err.Error(),
				Data:    nil,
			}
			respCh <- result
			return
		}

		bodyResp, err := common.UnpackFastHttpBody(resp)
		if err != nil {
			result.Error = &rpc.RpcError{
				Code:    error_codes.GenericServerError,
				Message: err.Error(),
				Data:    nil,
			}
			respCh <- result
			return
		}

		var users map[int64]SimpleUser

		err = json.Unmarshal(bodyResp, &users)
		if err != nil {
			result.Error = &rpc.RpcError{
				Code:    error_codes.GenericMappingError,
				Message: err.Error(),
				Data:    nil,
			}
			respCh <- result
			return
		}

		if len(users) > 0 {
			for id, user := range users {
				finalResponse[id] = user
				w.cache.SetInt64Key(id, user, w.defaultExpiration)
			}

			result.Items = finalResponse
		}

		respCh <- result
	})

	return respCh
}
