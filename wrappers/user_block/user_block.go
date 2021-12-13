package user_block

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

type IUserBlockWrapper interface {
	GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserBlockResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type UserBlockWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewUserBlockWrapper(config boilerplate.WrapperConfig) IUserBlockWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &UserBlockWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "user-block",
	}
}

func (w *UserBlockWrapper) GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction,
	forceLog bool) chan GetUserBlockResponseChan {
	resChan := make(chan GetUserBlockResponseChan, 2)

	w.baseWrapper.GetPool().Submit(func() {
		finalResponse := GetUserBlockResponseChan{}

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(
			fmt.Sprintf("%v/mobile/v1/user/block_relations", w.apiUrl),
			"POST",
			"application/json",
			"get user block",
			GetUserBlockRequest{
				BlockBy:   blockedBy,
				BlockedTo: blockedTo,
			}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse.Error = rpcInternalResponse.Error

		if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
			var userBlockData UserBlockData

			if err := json.Unmarshal(rpcInternalResponse.Result, &userBlockData); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:     error_codes.GenericMappingError,
					Message:  err.Error(),
					Data:     nil,
					Hostname: w.baseWrapper.GetHostName(),
				}
			} else {
				finalResponse.Data = userBlockData
			}
		}

		resChan <- finalResponse
	})

	return resChan
}