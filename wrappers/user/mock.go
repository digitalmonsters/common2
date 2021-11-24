package user

import "go.elastic.co/apm"

type UserWrapperMock struct {
	GetCachedUsersFn func(userIds []int64, apmTransaction *apm.Transaction) chan CachedUsersResponse
}

func (w *UserWrapperMock) GetCachedUsers(userIds []int64, apmTransaction *apm.Transaction) chan CachedUsersResponse {
	return w.GetCachedUsersFn(userIds, apmTransaction)
}

func GetMock() IUserWrapper {
	return &UserWrapperMock{}
}
