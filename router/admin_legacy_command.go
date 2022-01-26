package router

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"strings"
)

type LegacyAdminCommand struct {
	methodName                string
	accessLevel               common.AccessLevel
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
}

func NewLegacyAdminCommand(methodName string, fn CommandFunc) ICommand {
	return &LegacyAdminCommand{
		methodName:                strings.ToLower(methodName),
		accessLevel:               common.AccessLevelWrite,
		forceLog:                  true,
		fn:                        fn,
		requireIdentityValidation: true,
	}
}

func (a LegacyAdminCommand) CanExecute(ctx *fasthttp.RequestCtx, apmTransaction *apm.Transaction, auth auth_go.IAuthGoWrapper) (int64, *rpc.RpcError) {
	userId, err := publicCanExecuteLogic(ctx, a.requireIdentityValidation)

	if err != nil {
		return 0, err
	}

	if userId <= 0 {
		return 0, &rpc.RpcError{
			Code:        error_codes.MissingJwtToken,
			Message:     "legacy admin method requires identity validation",
			Hostname:    hostName,
			ServiceName: hostName,
		}
	}

	resp := <-auth.CheckLegacyAdmin(userId, apmTransaction, false)

	if resp.Error != nil {
		return 0, resp.Error
	}

	if resp.Resp.IsAdmin || resp.Resp.IsSuperAdmin {
		return userId, nil
	}

	return 0, &rpc.RpcError{
		Code:        error_codes.InvalidJwtToken,
		Message:     "user is not marked as admin",
		Stack:       "",
		Hostname:    hostName,
		ServiceName: hostName,
	}
}

func (a LegacyAdminCommand) GetPath() string {
	return a.GetMethodName()
}

func (a LegacyAdminCommand) GetHttpMethod() string {
	return "post"
}

func (a LegacyAdminCommand) ForceLog() bool {
	return a.forceLog
}

func (a LegacyAdminCommand) GetObj() string {
	return ""
}

func (a LegacyAdminCommand) RequireIdentityValidation() bool {
	return a.requireIdentityValidation
}

func (a LegacyAdminCommand) AccessLevel() common.AccessLevel {
	return a.accessLevel
}

func (a LegacyAdminCommand) GetMethodName() string {
	return a.methodName
}

func (a LegacyAdminCommand) GetFn() CommandFunc {
	return a.fn
}