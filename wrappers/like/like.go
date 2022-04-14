package like

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm/v2"
	"time"
)

type ILikeWrapper interface {
	GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan
	GetInternalLikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalLikedByUserResponseChan
	GetInternalDislikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]bool]
	GetInternalUserLikes(userId int64, size int, pageState string, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalUserLikesResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type LikeWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewLikeWrapper(config boilerplate.WrapperConfig) ILikeWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://likes"

		log.Warn().Msgf("Api Url is missing for Likes. Setting as default : %v", config.ApiUrl)
	}

	return &LikeWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "likes",
	}
}

func (w *LikeWrapper) GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction,
	forceLog bool) chan LastLikedByUserResponseChan {
	respCh := make(chan LastLikedByUserResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetLastLikesByUsers", GetLatestLikedByUserRequest{
		LimitPerUser: limitPerUser,
		UserIds:      userIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := LastLikedByUserResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			items := map[int64][]LikeRecord{}

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

func (w *LikeWrapper) GetInternalLikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalLikedByUserResponseChan {
	respCh := make(chan GetInternalLikedByUserResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetInternalLikedByUserBulk", GetInternalLikedByUserRequest{
		UserId:     userId,
		ContentIds: contentIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetInternalLikedByUserResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := map[int64]bool{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Data = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *LikeWrapper) GetInternalUserLikes(userId int64, size int, pageState string, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalUserLikesResponseChan {
	respCh := make(chan GetInternalUserLikesResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetInternalUserLikes", GetInternalUserLikesRequest{
		UserId:    userId,
		Size:      size,
		PageState: pageState,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetInternalUserLikesResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := getInternalUserLikesResponse{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.LikedContentIds = data.ContentIds
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *LikeWrapper) GetInternalDislikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]bool] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]bool](w.baseWrapper, w.apiUrl, "GetInternalDislikedByUserBulk", GetInternalDislikedByUserRequest{
		UserId:     userId,
		ContentIds: contentIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}
