package boilerplate

import (
	"context"
	"encoding/json"
	"github.com/digitalmonsters/go-common/wrappers/configurator"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffexporter"
	"go.elastic.co/apm"
	"log"
	"time"
)

func InitFeatureFlags(wrapper configurator.IConfiguratorWrapper, exportData bool, notifiers ...ffclient.NotifierConfig) error {
	var cfg = ffclient.Config{
		PollingInterval:         3 * time.Second,
		Logger:                  nil, //todo
		Context:                 context.Background(),
		Retriever:               NewHttpFlagsRetriever(wrapper),
		FileFormat:              "json",
		Notifiers:               notifiers,
		StartWithRetrieverError: false,
	}
	if exportData {
		cfg.DataExporter = ffclient.DataExporter{
			FlushInterval:    10 * time.Second,
			MaxEventInMemory: 1000,
			Exporter:         NewFlagsExporter(wrapper),
		}
	}
	return ffclient.Init(cfg)
}

type HttpFlagsRetriever struct {
	wrapper configurator.IConfiguratorWrapper
}

func NewHttpFlagsRetriever(wrapper configurator.IConfiguratorWrapper) *HttpFlagsRetriever {
	return &HttpFlagsRetriever{wrapper: wrapper}
}

func (r *HttpFlagsRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	res := <-r.wrapper.GetFeatureFlags(apm.TransactionFromContext(ctx), false)
	if res.Error != nil {
		return nil, res.Error.ToError()
	}
	js, err := json.Marshal(res.Response)
	if err != nil {
		return nil, err
	}
	return js, nil
}

type FlagsExporter struct {
	wrapper configurator.IConfiguratorWrapper
}

func NewFlagsExporter(wrapper configurator.IConfiguratorWrapper) *FlagsExporter {
	return &FlagsExporter{wrapper: wrapper}
}

func (f *FlagsExporter) Export(ctx context.Context, logger *log.Logger, featureEvents []ffexporter.FeatureEvent) error {
	var mappedEvents []configurator.FeatureEvent
	for _, ev := range featureEvents {
		mappedEvents = append(mappedEvents, configurator.FeatureEvent{
			Kind:         ev.Kind,
			ContextKind:  ev.ContextKind,
			UserKey:      ev.UserKey,
			CreationDate: ev.CreationDate,
			Key:          ev.Key,
			Variation:    ev.Variation,
			Value:        ev.Value,
			Default:      ev.Default,
			Version:      ev.Version,
		})
	}
	res := <-f.wrapper.CreateFeatureFlagEvents(mappedEvents, apm.TransactionFromContext(ctx), false)
	if res.Error != nil {
		return res.Error.ToError()
	}
	return nil
}

func (f *FlagsExporter) IsBulk() bool {
	return true
}
