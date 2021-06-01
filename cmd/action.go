package cmd

import (
	"github.com/oidc-proxy-ecosystem/oidc-tools/auth"
	"github.com/oidc-proxy-ecosystem/oidc-tools/config"
	"github.com/oidc-proxy-ecosystem/oidc-tools/routes"
	"github.com/urfave/cli/v2"
)

var LoginCommand = &cli.Command{
	Name:  "login",
	Usage: "openid connect login",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "cert",
			Aliases: []string{"c", "C"},
			Usage:   "ssl cert file",
			Value:   "ssl/server.crt",
		},
		&cli.StringFlag{
			Name:    "key",
			Aliases: []string{"k", "K"},
			Usage:   "ssl key file",
			Value:   "ssl/server.key",
		},
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p", "P"},
			Usage:   "http port number",
			Value:   3000,
		},
	},
	Action: func(c *cli.Context) error {
		return runLogin(c)
	},
}

func runLogin(c *cli.Context) error {
	certFile := c.String("cert")
	keyFile := c.String("key")
	port := c.Int("port")
	oidc := config.NewOidc()
	authenticator, err := auth.NewAuthenticator(c.Context, port, *oidc)
	if err != nil {
		return err
	}
	handler := routes.New(port, oidc, authenticator)
	return handler.Action(certFile, keyFile)
}
