/*
This package provides functionality to interact with the Nhost API.
*/
package nhostclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/nhost/cli/v2/nhostclient/graphql"
)

const (
	retryerMaxAttempts = 3
	retryerBaseDelay   = 2
)

type Client struct {
	baseURL string
	client  *http.Client
	Graphql *graphql.Client
	retryer BasicRetryer
}

func WithAccessToken(accessToken string) clientv2.RequestInterceptor {
	return func(
		ctx context.Context,
		req *http.Request,
		gqlInfo *clientv2.GQLRequestInfo,
		res interface{},
		next clientv2.RequestInterceptorFunc,
	) error {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		return next(ctx, req, gqlInfo, res)
	}
}

func New(domain string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("https://%s/v1/auth", domain),
		client:  &http.Client{}, //nolint:exhaustruct
		Graphql: graphql.NewClient(
			&http.Client{}, //nolint:exhaustruct
			fmt.Sprintf("https://%s/v1/graphql", domain),
			&clientv2.Options{}, //nolint:exhaustruct
		),
		retryer: NewBasicRetryer(retryerMaxAttempts, retryerBaseDelay),
	}
}
