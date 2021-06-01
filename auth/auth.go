package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc"
	"github.com/oidc-proxy-ecosystem/oidc-tools/config"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	Provider *oidc.Provider
	Config   oauth2.Config
	Ctx      context.Context
}

func NewAuthenticator(ctx context.Context, port int, oidcConf config.Oidc) (*Authenticator, error) {
	provider, err := oidc.NewProvider(ctx, oidcConf.Provider)
	if err != nil {
		return nil, err
	}
	o2conf := oauth2.Config{
		ClientID:     oidcConf.ClientId,
		ClientSecret: oidcConf.ClientSecret,
		RedirectURL:  fmt.Sprintf("https://localhost:%d/oauth/callback", port),
		Endpoint:     provider.Endpoint(),
		Scopes:       oidcConf.Scopes,
	}

	return &Authenticator{
		Provider: provider,
		Config:   o2conf,
		Ctx:      ctx,
	}, nil
}
