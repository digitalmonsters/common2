package elastic_search

import (
	"strings"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
)

type ElasticSearch struct {
	EsClient *elasticsearch.Client
	Config   *boilerplate.ElasticConfig
	settings elasticsearch.Config
}

type ElasticSearchService struct {
	Service *search.Search
	Index   Index
}

func Init(cfg *boilerplate.ElasticConfig) (*ElasticSearch, error) {
	// create Elasticsearch client
	esClient := &ElasticSearch{}

	esClient.settings = elasticsearch.Config{
		Username:            cfg.UserName,
		Password:            cfg.Password,
		Addresses:           strings.Split(cfg.Hosts, ","),
		CompressRequestBody: true,
	}
	var err error
	esClient.EsClient, err = elasticsearch.NewClient(esClient.settings)
	if err != nil {
		// Coz Initializations happens at the start. Sooooo if this fails, we panic
		panic(err)
	}

	return esClient, nil

}

func (es *ElasticSearch) InitIndex(index Index) *ElasticSearchService {
	esClientTyped, err := elasticsearch.NewTypedClient(es.settings)
	if err != nil {
		// Coz Initializations happens at the start. Sooooo if this fails, we panic
		panic(err)
	}
	return &ElasticSearchService{
		Service: esClientTyped.Search().Index(string(index)),
		Index:   index,
	}
}
