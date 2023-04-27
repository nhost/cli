package workflows

import (
	"context"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/nhost/cli/v2/nhostclient/credentials"
	"github.com/nhost/cli/v2/nhostclient/graphql"
)

type Printer interface {
	Printf(string, ...any)
	Println(...any)
	Print(...any)
}

type NhostClientAuth interface {
	Login(ctx context.Context, email string, password string) (credentials.Session, error)
	GetWorkspacesApps(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*graphql.GetWorkspacesApps, error)
	LoginPAT(ctx context.Context, pat string) (credentials.Session, error)
	Logout(ctx context.Context, tokenType string, accessToken string) error
	CreatePAT(ctx context.Context, accessToken string) (credentials.Credentials, error)
}
