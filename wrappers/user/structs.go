package user

import "github.com/digitalmonsters/go-common/rpc"

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
