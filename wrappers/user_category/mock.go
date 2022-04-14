package user_category

import "go.elastic.co/apm/v2"

//goland:noinspection ALL
type UserCategoryWrapperMock struct {
	GetUserCategorySubscriptionStateBulkFn func(categoryIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserCategorySubscriptionStateResponseChan
}

func (m *UserCategoryWrapperMock) GetUserCategorySubscriptionStateBulk(categoryIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserCategorySubscriptionStateResponseChan {
	return m.GetUserCategorySubscriptionStateBulkFn(categoryIds, userId, apmTransaction, forceLog)
}

func GetMock() IUserCategoryWrapper { // for compiler errors
	return &UserCategoryWrapperMock{}
}
