package config

import (
	"strings"

	"github.com/caarlos0/env"
	"github.com/mcuadros/go-defaults"
	"golang.org/x/oauth2"
)

type Oidc struct {
	Scopes       []string `default:"[email,openid,offline_access,profile]"`
	Provider     string   `env:"OIDC_PROVIDER"`
	ClientId     string   `env:"OIDC_CLIENT_ID"`
	ClientSecret string   `env:"OIDC_CLIENT_SECRET"`
	Audience     string   `env:"OIDC_AUDIENCE"`
	audiences    []string
}

func (o *Oidc) SetValues() []oauth2.AuthCodeOption {
	var authCodeOptions []oauth2.AuthCodeOption
	var audiences Audiences
	for _, audience := range o.audiences {
		if audience != "" {
			audiences = append(audiences, Audience(audience))
		}
	}
	authCodeOptions = append(authCodeOptions, audiences.SetValue()...)
	return authCodeOptions
}

func NewOidc() *Oidc {
	oidc := Oidc{}
	if err := env.Parse(&oidc); err != nil {
		return nil
	}
	defaults.SetDefaults(&oidc)
	oidc.audiences = strings.Split(oidc.Audience, ",")
	return &oidc
}
