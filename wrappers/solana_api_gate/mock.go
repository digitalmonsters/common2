package solana_api_gate

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
)

//goland:noinspection ALL
type SolanaApiGateWrapperMock struct {
	TransferTokenFn func(from string, amount string, account string, recipientType string, withdrawalTransactionId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[TransferTokenResponseData]
	CreateVestingFn func(from string, to string, amounts string, timestamps string, withdrawalTransactionId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateVestingResponseData]
}

func (m *SolanaApiGateWrapperMock) TransferToken(from string, amount string, account string, recipientType string, withdrawalTransactionId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[TransferTokenResponseData] {
	return m.TransferTokenFn(from, amount, account, recipientType, withdrawalTransactionId, ctx, forceLog)
}

func (m *SolanaApiGateWrapperMock) CreateVesting(from string, to string, amounts string, timestamps string, withdrawalTransactionId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateVestingResponseData] {
	return m.CreateVestingFn(from, to, amounts, timestamps, withdrawalTransactionId, ctx, forceLog)
}

func GetMock() ISolanaApiGateWrapper { // for compiler errors
	return &SolanaApiGateWrapperMock{}
}
