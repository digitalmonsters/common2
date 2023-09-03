package elastic_search

import (
	"context"
	"fmt"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type QueryBody struct {
	Query Query                  `json:"query"`
	Aggs  map[string]interface{} `json:"aggs,omitempty"`
	Sort  []interface{}          `json:"sort,omitempty"`
}

type Query struct {
	Bool struct {
		Should  []interface{} `json:"should,omitempty"`
		Must    []interface{} `json:"must,omitempty"`
		MustNot []interface{} `json:"must_not,omitempty"`
		Filter  []interface{} `json:"filter,omitempty"`
	} `json:"bool,omitempty"`
}

type Sort struct {
	Sort []interface{} `json:"sort"`
}

type Aggs struct {
	Aggs map[string]interface{} `json:"aggs"`
}

type ElasticRequest struct {
	Query        *types.Query
	Aggregations *types.Aggregations
	Sort         *types.Sort
	Request      *search.Request
	Size         int
	Page         int
	IsPagination bool
	indexClient  *search.Search
}

type ArrayValues interface {
	[]int | []string | []float32 | []float64
}

func (e *ElasticRequest) New(indexClient *search.Search) {
	e.Query = new(types.Query)
	e.Aggregations = new(types.Aggregations)
	e.Sort = new(types.Sort)
	e.Request = new(search.Request)

	e.Query.Bool = new(types.BoolQuery)
	e.Query.Bool.Must = make([]types.Query, 0)
	e.Query.Bool.MustNot = make([]types.Query, 0)
	e.Query.Bool.Filter = make([]types.Query, 0)
	e.Query.Bool.Should = make([]types.Query, 0)
	e.indexClient = indexClient
	e.Aggregations.Aggregations = make(map[string]types.Aggregations)

}

func (e *ElasticRequest) BuildPagination(size, page int) {
	if e.Request == nil {
		e.Request = new(search.Request)
	}
	if !e.IsPagination {
		e.Size = 0
		e.Page = 1
	} else {
		e.Size = size
		e.Page = page
	}
	e.Request.Size = &e.Size
	e.Request.From = NewInt((e.Page - 1) * e.Size)
}

func NewFloat(i int64) *types.Float64 {
	f := types.Float64(i)
	return &f
}

func AToF(s string) *types.Float64 {
	d := types.Float64(0)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return &d
	}
	d = types.Float64(f)
	return &d
}

func NewString(base interface{}, value string) *string {
	s := fmt.Sprintf("%v%s", base, value)
	return &s
}

func NewInt(i int) *int {
	return &i
}

func NewBool(b bool) *bool {
	return &b
}

func Sanitize(i []int) []string {
	ret := make([]string, 0)
	for _, d := range i {
		ret = append(ret, strconv.Itoa(d))
	}
	return ret
}

func (e *ElasticRequest) BuildMatchQuery(field string, value string) map[string]types.MatchQuery {
	return map[string]types.MatchQuery{
		field: {
			Query: value,
		},
	}
}

func (e *ElasticRequest) Do() (*search.Response, error) {
	e.Request.Query = e.Query
	e.Request.Aggregations = e.Aggregations.Aggregations
	if e.Sort != nil {
		e.Request.Sort = append(e.Request.Sort, *e.Sort...)
	}
	e.Request.TrackScores = NewBool(true)
	e.Request.TrackTotalHits = true
	return e.indexClient.Request(e.Request).Do(context.Background())
}
