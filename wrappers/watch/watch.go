package watch

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"time"
)

type IWatchWrapper interface {
	GetLastWatchesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastWatcherByUserResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type WatchWrapper struct {
	apiUrl         string
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	serviceName    string
}

func NewWatchWrapper(config boilerplate.WrapperConfig) IWatchWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &WatchWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "views",
	}
}

func (w *WatchWrapper) GetLastWatchesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction,
	forceLog bool) chan LastWatcherByUserResponseChan {
	respCh := make(chan LastWatcherByUserResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetLastWatchesByUsers", GetLatestWatchesByUserRequest{
		LimitPerUser: limitPerUser,
		UserIds:      userIds,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := LastWatcherByUserResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			items := map[int64][]LastWatchesByUserRecord{}

			if err := json.Unmarshal(resp.Result, &items); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = items
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *WatchWrapper) GetCategoriesByViews(limit int64, offset int64, apmTransaction *apm.Transaction) chan GetCategoriesResponseChan {
	respCh := make(chan GetCategoriesResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetCategoriesByViews", GetCategoriesByViewsRequest{
		Limit:  limit,
		Offset: offset,
	}, w.defaultTimeout, apmTransaction, w.serviceName, false)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetCategoriesResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			items := make([]CategoryInfo, 0)

			if err := json.Unmarshal(resp.Result, &items); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = items
			}
		}

		respCh <- result
	}()

	return respCh
}
