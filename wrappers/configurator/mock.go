package configurator

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
)

type ConfiguratorWrapperMock struct {
	GetFeatureFlagsFn         func(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[string]FeatureToggleConfig]
	CreateFeatureFlagEventsFn func(ctx context.Context, events []FeatureEvent, forceLog bool) chan wrappers.GenericResponseChan[map[string]interface{}]
}

func (m *ConfiguratorWrapperMock) GetFeatureFlags(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[string]FeatureToggleConfig] {
	return m.GetFeatureFlagsFn(ctx, forceLog)
}

func (m *ConfiguratorWrapperMock) CreateFeatureFlagEvents(ctx context.Context, events []FeatureEvent, forceLog bool) chan wrappers.GenericResponseChan[map[string]interface{}] {
	return m.CreateFeatureFlagEventsFn(ctx, events, forceLog)
}
