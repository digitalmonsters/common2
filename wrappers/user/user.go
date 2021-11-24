package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/cache/inmemory_cache"
	"go.elastic.co/apm"
	"io/ioutil"
	"net/http"
	"time"
)

type Wrapper struct {
	apiUrl            string
	defaultExpiration time.Duration
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

type IUserWrapper interface {
	GetCachedUsers(userIds []int64, apmTransaction *apm.Transaction) (map[int64]SimpleUser, error)
}

func New(apiUrl string, cacheDefaultExp time.Duration) IUserWrapper {
	return &Wrapper{
		apiUrl:            apiUrl,
		defaultExpiration: cacheDefaultExp,
		cache:             inmemory_cache.New(cacheDefaultExp),
	}
}

func (w *Wrapper) GetCachedUsers(userIds []int64, apmTransaction *apm.Transaction) (map[int64]SimpleUser, error) {
	result := map[int64]SimpleUser{}

	cachedUsers, missingInCache := w.cache.Get(userIds)
	for id, iface := range cachedUsers {
		user, ok := iface.(SimpleUser)
		if !ok {
			apm_helper.CaptureApmError(errors.New("cannot convert interface from cache"), apmTransaction)
			continue
		}

		result[id] = user
	}

	if len(missingInCache) == 0 {
		return result, nil
	}

	//todo: api request logic
	var body []byte

	httpReq, err := http.NewRequest("POST", w.apiUrl, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpRes, err := apm_helper.SendRequest(http.DefaultClient, httpReq, apmTransaction, true)

	var bodyResp []byte
	if httpRes != nil && httpRes.Body != nil {
		bodyResp, _ = ioutil.ReadAll(httpRes.Body)
	}

	var users []SimpleUser

	err = json.Unmarshal(bodyResp, &users)
	if err != nil {
		return nil, err
	}

	toCache := map[int64]interface{}{}
	for _, user := range users {
		result[user.Id] = user
		toCache[user.Id] = user
	}

	w.cache.Set(toCache, w.defaultExpiration)

	return result, nil
}
