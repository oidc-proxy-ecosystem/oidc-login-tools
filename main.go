package main

import (
	"log"
	"os"

	"github.com/oidc-proxy-ecosystem/oidc-tools/cmd"
	"github.com/urfave/cli/v2"
)

var (
	Version  string = "0.0.1"
	Revision string = "beta"
)

func main() {
	app := cli.NewApp()
	app.Name = "openid connect proxy server"
	app.Version = Version + " - " + Revision
	app.Description = "openid connect proxy server"
	app.Commands = []*cli.Command{
		cmd.LoginCommand,
		cmd.KuberntesConfigCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
}
