package category

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Wrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

type ICategoryWrapper interface {
	GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan
}

func NewCategoryWrapper(config boilerplate.WrapperConfig) ICategoryWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &Wrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "content",
	}
}

func (w *Wrapper) GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan {
	respCh := make(chan CategoryGetInternalResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetCategoryInternal", GetCategoryInternalRequest{
		CategoryIds:     categoryIds,
		Limit:           limit,
		Offset:          offset,
		OmitCategoryIds: omitCategoryIds,
		OnlyParent:      onlyParent,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := CategoryGetInternalResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := &ResponseData{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:     error_codes.GenericMappingError,
					Message:  err.Error(),
					Data:     nil,
					Hostname: w.baseWrapper.GetHostName(),
				}
			} else {
				result.Data = data
			}
		}

		respCh <- result
	})

	return respCh
}
