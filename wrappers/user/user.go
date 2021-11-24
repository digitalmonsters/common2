package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/cache/inmemory_cache"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"io/ioutil"
	"net/http"
	"time"
)

type Wrapper struct {
	apiUrl            string
	defaultExpiration time.Duration
	baseWrapper       *wrappers.BaseWrapper
	cache             *inmemory_cache.Service
}

type SimpleUser struct {
	Id          int64  `json:"id"`
	Avatar      string `json:"avatar"`
	DisplayName string `json:"displayname"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	Username    string `json:"username"`
	Verified    bool   `json:"verified"`
}

type CachedUsersResponse struct {
	Error *rpc.RpcError        `json:"error"`
	Items map[int64]SimpleUser `json:"items"`
}

type IUserWrapper interface {
	GetCachedUsers(userIds []int64, apmTransaction *apm.Transaction) chan CachedUsersResponse
}

func New(apiUrl string, cacheDefaultExp time.Duration) IUserWrapper {
	return &Wrapper{
		apiUrl:            apiUrl,
		defaultExpiration: cacheDefaultExp,
		baseWrapper:       wrappers.GetBaseWrapper(),
		cache:             inmemory_cache.New(cacheDefaultExp),
	}
}

func (w *Wrapper) GetCachedUsers(userIds []int64, apmTransaction *apm.Transaction) chan CachedUsersResponse {
	respCh := make(chan CachedUsersResponse, 2)

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

		close(respCh)
		return respCh
	}

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		result := CachedUsersResponse{
			Error: nil,
		}

		//todo: api request logic
		var body []byte

		httpReq, err := http.NewRequest("POST", w.apiUrl, bytes.NewReader(body))
		if err != nil {
			result.Error = &rpc.RpcError{
				Code:    error_codes.GenericServerError,
				Message: err.Error(),
				Data:    nil,
			}
			respCh <- result
			return
		}

		httpRes, err := apm_helper.SendRequest(http.DefaultClient, httpReq, apmTransaction, true)

		var bodyResp []byte
		if httpRes != nil && httpRes.Body != nil {
			bodyResp, _ = ioutil.ReadAll(httpRes.Body)
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
			toCache := map[int64]interface{}{}
			for id, user := range users {
				finalResponse[id] = user
				toCache[id] = user
			}

			w.cache.Set(toCache, w.defaultExpiration)
			result.Items = finalResponse
		}

		respCh <- result
	})

	return respCh
}
