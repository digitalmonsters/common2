package http_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"io/ioutil"
	"time"
)

type HttpClient struct {
	cl                *req.Client
	targetServiceName string
}

type HttpRequest struct {
	*req.Request
}

type AsyncChan struct {
	Resp *req.Response
	Err  error
}

var DefaultHttpClient = NewHttpClient()

type forceLogKey struct {
}

func NewHttpClient() *HttpClient {
	client := req.C().SetTimeout(30 * time.Second)

	h := &HttpClient{
		cl: client,
	}

	client.OnBeforeRequest(func(client *req.Client, request *req.Request) error {
		if parentApm := apm.TransactionFromContext(request.Context()); parentApm != nil {
			targetServiceName := h.targetServiceName

			if len(targetServiceName) == 0 {
				targetServiceName = request.URL.Hostname()
			}
			span := parentApm.StartSpan(fmt.Sprintf("HTTP [%v] [%v]", request.RawRequest.Method, request.RawURL),
				targetServiceName, nil)

			request.SetContext(apm.ContextWithSpan(request.Context(), span))
		}

		return nil
	})

	client.OnAfterResponse(func(client *req.Client, response *req.Response) error {
		forceLog, _ := response.Request.Context().Value(forceLogKey{}).(bool)

		if response.IsError() {
			forceLog = true
		}

		ctx := response.Request.Context()

		if span := apm.SpanFromContext(ctx); span != nil {
			finalStatement := ""

			if forceLog {
				var rawBodyRequest []byte
				var rawBodyResponse []byte

				if r, err := ioutil.ReadAll(response.Request.RawRequest.Body); err != nil {
					log.Ctx(ctx).Err(err).Send()
				} else {
					rawBodyRequest = r
				}

				if r, err := ioutil.ReadAll(response.Response.Body); err != nil {
					log.Ctx(ctx).Err(err).Send()
				} else {
					rawBodyResponse = r
				}

				if data, err := json.Marshal(map[string]interface{}{
					"request":  rawBodyRequest,
					"response": rawBodyResponse,
				}); err != nil {
					log.Ctx(ctx).Err(err).Send()

					finalStatement = fmt.Sprintf("request [%v] || response [%v]", rawBodyRequest, rawBodyResponse)
				} else {
					finalStatement = string(data)
				}
			}

			span.Context.SetDatabase(apm.DatabaseSpanContext{
				Instance:  response.Request.URL.Hostname(),
				Type:      response.Request.URL.Hostname(),
				Statement: finalStatement,
			})

			span.End()
		}
		return nil
	})

	return h
}

func (h *HttpClient) WithServiceName(serviceName string) *HttpClient {
	h.targetServiceName = serviceName

	return h
}

func (h *HttpClient) WithTimeout(duration time.Duration) *HttpClient {
	h.cl.SetTimeout(duration)

	return h
}

func (h HttpClient) NewRequest(ctx context.Context) *HttpRequest {
	return &HttpRequest{h.cl.R().SetContext(ctx)}
}

func (h HttpClient) NewRequestWithTimeout(ctx context.Context, timeout time.Duration) *HttpRequest {
	return &HttpRequest{h.cl.Clone().SetTimeout(timeout).R().SetContext(ctx)}
}

func (r *HttpRequest) WithForceLog() *HttpRequest {
	r.SetContext(context.WithValue(r.Context(), forceLogKey{}, true))

	return r
}

func (r *HttpRequest) GetAsync(url string) chan AsyncChan {
	return r.doAsync(func() (*req.Response, error) {
		return r.Get(url)
	})
}

func (r *HttpRequest) PostAsync(url string) chan AsyncChan {
	return r.doAsync(func() (*req.Response, error) {
		return r.Post(url)
	})
}

func (r HttpRequest) doAsync(fn func() (*req.Response, error)) chan AsyncChan {
	ch := make(chan AsyncChan, 2)

	go func() {
		defer func() {
			close(ch)
		}()

		r, e := fn()

		ch <- AsyncChan{
			Resp: r,
			Err:  e,
		}
	}()

	return ch
}
