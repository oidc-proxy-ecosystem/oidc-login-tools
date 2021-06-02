package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/oidc-proxy-ecosystem/oidc-tools/auth"
	"github.com/oidc-proxy-ecosystem/oidc-tools/config"
	"github.com/urfave/cli/v2"
)

var LoginCommand = &cli.Command{
	Name:  "login",
	Usage: "openid connect login",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "idp-issuer-url",
			Aliases:  []string{"idp"},
			Usage:    "identity provider isser url",
			Required: true,
			EnvVars:  []string{"OIDC_PROVIDER"},
		},
		&cli.StringFlag{
			Name:     "client-id",
			Aliases:  []string{"id"},
			Usage:    "identity provider client_id",
			Required: true,
			EnvVars:  []string{"OIDC_CLIENT_ID"},
		},
		&cli.StringFlag{
			Name:     "client-secret",
			Aliases:  []string{"sec"},
			Usage:    "identity provider client_secret",
			Required: true,
			EnvVars:  []string{"OIDC_CLIENT_SECRET"},
		},
		&cli.StringFlag{
			Name:  "callback-url",
			Usage: "idp callback url",
			Value: "https://localhost/oauth/callback",
		},
		&cli.StringSliceFlag{
			Name:  "scopes",
			Usage: "openid connect scope",
			Value: cli.NewStringSlice("email", "openid", "offline_access", "profile"),
		},
		&cli.StringSliceFlag{
			Name:  "audiences",
			Usage: "openid connect audience",
			Value: cli.NewStringSlice(),
		},
	},
	Action: func(c *cli.Context) error {
		return runLogin(c)
	},
}

func runLogin(c *cli.Context) error {
	idpIssuerUrl := c.String("idp-issuer-url")
	clientId := c.String("client-id")
	clientSecret := c.String("client-secret")
	scopes := c.StringSlice("scopes")
	audiences := c.StringSlice("audiences")

	callbackUrl := c.String("callback-url")
	oidc := config.NewOidc(scopes, idpIssuerUrl, clientId, clientSecret, audiences)
	authenticator, err := auth.NewAuthenticator(c.Context, *oidc, callbackUrl)
	if err != nil {
		return err
	}
	return login(c.Context, oidc, authenticator)
}

func login(ctx context.Context, oidcConfig *config.Oidc, authenticator *auth.Authenticator) error {
	u := authenticator.Config.AuthCodeURL("", oidcConfig.SetValues()...)
	log.Printf("open browser: %s\n", u)
	fmt.Print("code: >")
	scanner := bufio.NewScanner(os.Stdin)
	var code string
	scanner.Scan()
	code = scanner.Text()
	if code == "" {
		return errors.New("codeを入力して下さい。")
	}
	token, err := authenticator.Config.Exchange(ctx, code, oidcConfig.SetValues()...)
	if err != nil {
		return fmt.Errorf("no token found: %v", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return errors.New("No id_token field in oauth2 token.")
	}

	verifyOidcConfig := &oidc.Config{
		ClientID: oidcConfig.ClientId,
	}

	_, err = authenticator.Provider.Verifier(verifyOidcConfig).Verify(ctx, rawIDToken)

	if err != nil {
		return errors.New("Failed to verify ID Token: " + err.Error())
	}
	tokens := &config.Token{
		IdToken:      rawIDToken,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	file, err := os.Create("./auth.json")
	if err != nil {
		return errors.New("auth.jsonの生成に失敗しました。")
	}
	defer file.Close()
	buf, _ := json.Marshal(tokens)
	_, err = file.Write(buf)
	return err
}
