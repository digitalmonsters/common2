package configurator

import (
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

type ConfiguratorWrapperMock struct {
	GetFeatureFlagsFn         func(apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]FeatureToggleConfig]
	CreateFeatureFlagEventsFn func(events []interface{}, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]interface{}]
}

func (m *ConfiguratorWrapperMock) GetFeatureFlags(apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]FeatureToggleConfig] {
	return m.GetFeatureFlagsFn(apmTransaction, forceLog)
}

func (m *ConfiguratorWrapperMock) CreateFeatureFlagEvents(events []interface{}, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]interface{}] {
	return m.CreateFeatureFlagEventsFn(events, apmTransaction, forceLog)
}
