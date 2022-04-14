package user_go

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm/v2"
)

//goland:noinspection ALL
type UserGoWrapperMock struct {
	GetUsersFn func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan

	GetUsersDetailFn func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserDetailRecord]
	GetUserDetailsFn func(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserDetailRecord]

	GetProfileBulkFn                      func(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
	GetUsersActiveThresholdsFn            func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan
	GetUserIdsFilterByUsernameFn          func(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan
	GetUsersTagsFn                        func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan
	AuthGuestFn                           func(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AuthGuestResp]
	GetBlockListFn                        func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string][]int64]
	GetUserBlockFn                        func(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[UserBlockData]
	UpdateUserMetadataAfterRegistrationFn func(request UpdateUserMetaDataRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord]
	ForceResetUserWithNewGuestIdentityFn  func(deviceId string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[ForceResetUserIdentityWithNewGuestResponse]
	VerifyUserFn                          func(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord]
}

func (m *UserGoWrapperMock) GetUserDetails(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserDetailRecord] {
	return m.GetUserDetailsFn(userId, ctx, forceLog)
}

func (m *UserGoWrapperMock) VerifyUser(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord] {
	return m.VerifyUserFn(userId, ctx, forceLog)
}

func (m *UserGoWrapperMock) ForceResetUserWithNewGuestIdentity(deviceId string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[ForceResetUserIdentityWithNewGuestResponse] {
	return m.ForceResetUserWithNewGuestIdentityFn(deviceId, ctx, forceLog)
}

func (m *UserGoWrapperMock) UpdateUserMetadataAfterRegistration(request UpdateUserMetaDataRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord] {
	return m.UpdateUserMetadataAfterRegistrationFn(request, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan {
	return m.GetUsersFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUsersDetails(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserDetailRecord] {
	return m.GetUsersDetailFn(userIds, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan {
	return m.GetProfileBulkFn(currentUserId, userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUsersActiveThresholds(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan {
	return m.GetUsersActiveThresholdsFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUserIdsFilterByUsername(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan {
	return m.GetUserIdsFilterByUsernameFn(userIds, searchQuery, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUsersTags(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan {
	return m.GetUsersTagsFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AuthGuestResp] {
	return m.AuthGuestFn(deviceId, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetBlockList(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string][]int64] {
	return m.GetBlockListFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[UserBlockData] {
	return m.GetUserBlockFn(blockedTo, blockedBy, apmTransaction, forceLog)
}

func GetMock() IUserGoWrapper { // for compiler errors
	return &UserGoWrapperMock{}
}
