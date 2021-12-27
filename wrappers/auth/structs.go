package auth

import "github.com/digitalmonsters/go-common/rpc"

type AuthParseTokenRequest struct {
	Token            string `json:"token"`
	IgnoreExpiration bool   `json:"ignore_expiration"`
}

type AuthParseTokenResponseChan struct {
	Resp  AuthParseTokenResponse
	Error *rpc.RpcError
}

type AuthParseTokenResponse struct {
	UserId       int64 `json:"user_id"`
	Expired      bool  `json:"expired"`
	IsAdmin      bool  `json:"is_admin"`
	IsSuperAdmin bool  `json:"is_super_admin"`
}

type GenerateTokenResponseChan struct {
	Resp  GenerateTokenResponse
	Error *rpc.RpcError
}

type GenerateTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
