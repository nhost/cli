/*
This package implements the main business logic in the cli
*/

package controller

import (
	"context"
	"fmt"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/nhost/cli/v2/nhostclient/credentials"
	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
)

type Printer interface {
	Printf(string, ...any)
	Println(...any)
	Print(...any)
}

type NhostClientAuth interface {
	Login(ctx context.Context, email string, password string) (credentials.Session, error)
	LoginPAT(ctx context.Context, pat string) (credentials.Session, error)
	Logout(ctx context.Context, tokenType string, accessToken string) error
	CreatePAT(ctx context.Context, accessToken string) (credentials.Credentials, error)
}

type NhostClientSecrets interface {
	GetSecrets(
		ctx context.Context, appID string, interceptors ...clientv2.RequestInterceptor,
	) (*graphql.GetSecrets, error)
	UpdateSecret(
		ctx context.Context, appID string, name string, value string, interceptors ...clientv2.RequestInterceptor,
	) (*graphql.UpdateSecret, error)
	CreateSecret(
		ctx context.Context, appID string, name string, value string, interceptors ...clientv2.RequestInterceptor,
	) (*graphql.CreateSecret, error)
	DeleteSecret(
		ctx context.Context, appID string, name string, interceptors ...clientv2.RequestInterceptor,
	) (*graphql.DeleteSecret, error)
}

type NhostClient interface {
	NhostClientAuth
	NhostClientSecrets

	GetWorkspacesApps(ctx context.Context, interceptors ...clientv2.RequestInterceptor) (*graphql.GetWorkspacesApps, error)
	GetConfigRawJSON(
		ctx context.Context, appID string, interceptors ...clientv2.RequestInterceptor,
	) (*graphql.GetConfigRawJSON, error)
	GetHasuraAdminSecret(
		ctx context.Context, appID string, interceptors ...clientv2.RequestInterceptor,
	) (*graphql.GetHasuraAdminSecret, error)
}

type CredentialsFunc func() (credentials.Credentials, error)

func GetNhostSession(ctx context.Context, cl NhostClient) (credentials.Session, error) {
	creds, err := project.AuthFromDisk()
	if err != nil {
		return credentials.Session{}, err
	}

	session, err := cl.LoginPAT(ctx, creds.PersonalAccessToken)
	if err != nil {
		return credentials.Session{}, fmt.Errorf("failed to login: %w", err)
	}

	return session, nil
}
