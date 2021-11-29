package watch

import "github.com/digitalmonsters/go-common/rpc"

type LastWatchesByUserRecord struct {
	ContentId  int64   `json:"content_id"`
	Duration   int     `json:"duration"`
	IsFullView bool    `json:"is_full_view"`
	Percent    float64 `json:"percent"`
}

type LastWatcherByUserResponseChan struct {
	Error *rpc.RpcError                       `json:"error"`
	Items map[int64][]LastWatchesByUserRecord `json:"items"`
}

type GetLatestWatchesByUserRequest struct {
	LimitPerUser int     `json:"limit_per_user"`
	UserIds      []int64 `json:"user_ids"`
	MinPercent   float64 `json:"min_percent"`
}
