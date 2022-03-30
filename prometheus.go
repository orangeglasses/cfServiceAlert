package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PrometheusClient struct { //assumes proxy through grafana
	api v1.API
}

func NewPrometheusClient(pURL string, apiKey string) (*PrometheusClient, error) {
	client, err := api.NewClient(api.Config{
		Address: pURL,

		RoundTripper: NewApiKeyAuthRoundTripper(apiKey, api.DefaultRoundTripper),
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	return &PrometheusClient{
		api: v1.NewAPI(client),
	}, nil
}

func (p *PrometheusClient) Query(q string) (model.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := p.api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	return result, nil
}

type ApiKeyAuthRoundTripper struct {
	apiKey string
	rt     http.RoundTripper
}

func NewApiKeyAuthRoundTripper(apiKey string, rt http.RoundTripper) http.RoundTripper {
	return &ApiKeyAuthRoundTripper{apiKey, rt}
}

func (rt *ApiKeyAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		return rt.rt.RoundTrip(req)
	}

	r2 := new(http.Request)
	*r2 = *req
	// Deep copy of the Header.
	r2.Header = make(http.Header)
	for k, s := range req.Header {
		r2.Header[k] = s
	}

	r2.Header.Add("Authorization", fmt.Sprintf("Bearer %s", rt.apiKey))

	return rt.rt.RoundTrip(r2)
}

type closeIdler interface {
	CloseIdleConnections()
}

func (rt *ApiKeyAuthRoundTripper) CloseIdleConnections() {
	if ci, ok := rt.rt.(closeIdler); ok {
		ci.CloseIdleConnections()
	}
}
