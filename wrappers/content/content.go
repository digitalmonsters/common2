package content

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm/v2"
	"gopkg.in/guregu/null.v4"
	"time"
)

type IContentWrapper interface {
	GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleContent]
	GetTopNotFollowingUsers(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetTopNotFollowingUsersResponse]
	GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool, apmTransaction *apm.Transaction,
		shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[HashtagResponseData]

	GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool,
		apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[CategoryResponseData]
	GetAllCategories(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]AllCategoriesResponseItem]
	GetUserBlacklistedCategories(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetUserBlacklistedCategoriesResponse]
	GetUserLikes(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[LikedContent]
	GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]string]
}

//goland:noinspection GoNameStartsWithPackageName
type ContentWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewContentWrapper(config boilerplate.WrapperConfig) IContentWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://content"

		log.Warn().Msgf("Api Url is missing for Content. Setting as default : %v", config.ApiUrl)
	}

	return &ContentWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "content",
	}
}

func (w *ContentWrapper) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction,
	forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleContent] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]SimpleContent](w.baseWrapper, w.apiUrl, "ContentGetInternal", ContentGetInternalRequest{
		ContentIds:     contentIds,
		IncludeDeleted: includeDeleted,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetTopNotFollowingUsers(userId int64, limit int, offset int, apmTransaction *apm.Transaction,
	forceLog bool) chan wrappers.GenericResponseChan[GetTopNotFollowingUsersResponse] {

	return wrappers.ExecuteRpcRequestAsync[GetTopNotFollowingUsersResponse](w.baseWrapper, w.apiUrl, "GetTopNotFollowingUsers", GetTopNotFollowingUsersRequest{
		UserId: userId,
		Limit:  limit,
		Offset: offset,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool,
	apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[HashtagResponseData] {
	return wrappers.ExecuteRpcRequestAsync[HashtagResponseData](w.baseWrapper, w.apiUrl, "GetHashtagsInternal", GetHashtagsInternalRequest{
		Hashtags:               hashtags,
		OmitHashtags:           omitHashtags,
		Limit:                  limit,
		WithViews:              withViews,
		Offset:                 offset,
		ShouldHaveValidContent: shouldHaveValidContent,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetUserBlacklistedCategories(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetUserBlacklistedCategoriesResponse] {
	return wrappers.ExecuteRpcRequestAsync[GetUserBlacklistedCategoriesResponse](w.baseWrapper, w.apiUrl, "GetUserBlacklistedCategoriesInternal", GetUserBlacklistedCategoriesRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool,
	apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[CategoryResponseData] {
	return wrappers.ExecuteRpcRequestAsync[CategoryResponseData](w.baseWrapper, w.apiUrl, "GetCategoryInternal", GetCategoryInternalRequest{
		CategoryIds:            categoryIds,
		Limit:                  limit,
		Offset:                 offset,
		OmitCategoryIds:        omitCategoryIds,
		WithViews:              withViews,
		OnlyParent:             onlyParent,
		ShouldHaveValidContent: shouldHaveValidContent,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetAllCategories(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]AllCategoriesResponseItem] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]AllCategoriesResponseItem](w.baseWrapper, w.apiUrl, "GetAllCategories", GetAllCategoriesRequest{
		CategoryIds:    categoryIds,
		IncludeDeleted: includeDeleted,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetUserLikes(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[LikedContent] {
	return wrappers.ExecuteRpcRequestAsync[LikedContent](w.baseWrapper, w.apiUrl, "InternalGetUserLikes", GetUserLikesRequest{
		UserId: userId,
		Limit:  limit,
		Offset: offset,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]string] {
	return wrappers.ExecuteRpcRequestAsync[map[string]string](w.baseWrapper, w.apiUrl, "InternalGetConfigValues", GetConfigValuesRequest{Properties: properties},
		map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}
