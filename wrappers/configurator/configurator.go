package configurator

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IConfiguratorWrapper interface {
	GetFeatureFlags(apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]FeatureToggleConfig]
	CreateFeatureFlagEvents(events []FeatureEvent, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]interface{}]
}

//goland:noinspection GoNameStartsWithPackageName
type ConfiguratorWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func (c *ConfiguratorWrapper) GetFeatureFlags(apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]FeatureToggleConfig] {
	return wrappers.ExecuteRpcRequestAsync[map[string]FeatureToggleConfig](c.baseWrapper, c.apiUrl, "InternalGetFeatureToggles", nil, map[string]string{}, c.defaultTimeout,
		apmTransaction, c.serviceName, forceLog)
}

func (c *ConfiguratorWrapper) CreateFeatureFlagEvents(events []FeatureEvent, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]interface{}] {
	return wrappers.ExecuteRpcRequestAsync[map[string]interface{}](c.baseWrapper, c.apiUrl, "InternalCreateFeatureToggleEvent", CreateFeatureToggleEventsRequest{Events: events},
		map[string]string{}, c.defaultTimeout, apmTransaction, c.serviceName, forceLog)
}

func NewConfiguratorWrapper(config boilerplate.WrapperConfig) IConfiguratorWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://configurator"

		log.Warn().Msgf("Api Url is missing for Content. Setting as default : %v", config.ApiUrl)
	}

	return &ConfiguratorWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "configurator",
	}
}
