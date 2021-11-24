package user

import "go.elastic.co/apm"

type UserWrapperMock struct {
	GetCachedUsersFn func(userIds []int64, apmTransaction *apm.Transaction) (map[int64]SimpleUser, error)
}

func (w *UserWrapperMock) GetCachedUsers(userIds []int64, apmTransaction *apm.Transaction) (map[int64]SimpleUser, error) {
	return w.GetCachedUsersFn(userIds, apmTransaction)
}

func GetMock() IUserWrapper {
	return &UserWrapperMock{}
}
