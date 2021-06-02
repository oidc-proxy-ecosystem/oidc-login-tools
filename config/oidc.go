package config

import (
	"golang.org/x/oauth2"
)

type Oidc struct {
	Scopes       []string
	Provider     string `env:"OIDC_PROVIDER"`
	ClientId     string `env:"OIDC_CLIENT_ID"`
	ClientSecret string `env:"OIDC_CLIENT_SECRET"`
	Audiences    []string
}

func (o *Oidc) SetValues() []oauth2.AuthCodeOption {
	var authCodeOptions []oauth2.AuthCodeOption
	var audiences Audiences
	for _, audience := range o.Audiences {
		if audience != "" {
			audiences = append(audiences, Audience(audience))
		}
	}
	authCodeOptions = append(authCodeOptions, audiences.SetValue()...)
	return authCodeOptions
}

func NewOidc(scopes []string, provider, clientId, clientSecret string, audience []string) *Oidc {
	return &Oidc{
		Scopes:       scopes,
		Provider:     provider,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Audiences:    audience,
	}
}
