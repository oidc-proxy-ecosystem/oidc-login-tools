package cmd

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/oidc-proxy-ecosystem/oidc-tools/config"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var KuberntesConfigCommand = &cli.Command{
	Name:  "config",
	Usage: "modify kubernetes config",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "cluster-name",
			Aliases: []string{"cln"},
			Usage:   "cluster name",
			Value:   "default-cluster",
			EnvVars: []string{""},
		},
		&cli.StringFlag{
			Name:     "cluster-url",
			Aliases:  []string{"clu"},
			Usage:    "cluster server url",
			Required: true,
		},
		&cli.BoolFlag{
			Name:    "tls-verify",
			Aliases: []string{"tv"},
			Usage:   "insecure skip tls verify",
			Value:   true,
		},
		&cli.StringFlag{
			Name:    "user-name",
			Aliases: []string{"un"},
			Usage:   "user name",
			Value:   "default-user",
		},
		&cli.StringFlag{
			Name:    "provider-name",
			Aliases: []string{"pn"},
			Usage:   "provider name",
			Value:   "oidc",
		},
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
			Name:    "token-file",
			Aliases: []string{"f"},
			Usage:   "token file path",
		},
		&cli.StringFlag{
			Name:  "id-token",
			Usage: "idp issuer id token",
		},
		&cli.StringFlag{
			Name:  "refresh-token",
			Usage: "idp issuer refresh token",
		},
		&cli.StringFlag{
			Name:    "context-name",
			Aliases: []string{"ctxn"},
			Usage:   "context name",
			Value:   "default-context",
		},
		&cli.StringFlag{
			Name:     "namespace",
			Aliases:  []string{"n", "ns"},
			Usage:    "kubernetes namespace",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		return setK8sConfig(c)
	},
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

type arguments struct {
	clusterName  string
	clusterUrl   string
	tlsVerify    bool
	userName     string
	providerName string
	idpIssuerUrl string
	clientId     string
	clientSecret string
	tokenFile    string
	idToken      string
	refreshToken string
	contextName  string
	namespace    string
}

func (args *arguments) getProviderMap() map[string]string {
	return map[string]string{
		"idp-issuer-url": args.idpIssuerUrl,
		"client-id":      args.clientId,
		"client-secret":  args.clientSecret,
		"refresh-token":  args.refreshToken,
		"id-token":       args.idToken,
	}
}

func newArguments(c *cli.Context) (*arguments, error) {
	arg := arguments{
		clusterName:  c.String("cluster-name"),
		clusterUrl:   c.String("cluster-url"),
		tlsVerify:    c.Bool("tls-verify"),
		userName:     c.String("user-name"),
		providerName: c.String("provider-name"),
		idpIssuerUrl: c.String("idp-issuer-url"),
		clientId:     c.String("client-id"),
		clientSecret: c.String("client-secret"),
		tokenFile:    c.String("token-file"),
		idToken:      c.String("id-token"),
		refreshToken: c.String("refresh-token"),
		contextName:  c.String("context-name"),
		namespace:    c.String("namespace"),
	}
	readFile := false
	if !fileExists("./auth.json") {
		if arg.tokenFile == "" {
			if arg.idToken == "" || arg.refreshToken != "" {
				return nil, errors.New("tokenファイルパスを設定しない場合は`idToken`, `refreshToken`を指定してください。")
			}
		} else {
			readFile = true
		}
	} else {
		readFile = true
	}
	if readFile {
		buf, err := os.ReadFile("./auth.json")
		if err != nil {
			return nil, err
		}
		token := config.Token{}
		json.Unmarshal(buf, &token)
		if err := token.Valid(); err != nil {
			return nil, err
		}
		arg.idToken = token.IdToken
		arg.refreshToken = token.RefreshToken
	}
	if arg.providerName == "" {
		arg.providerName = "oidc"
	}
	return &arg, nil
}

func setK8sConfig(c *cli.Context) error {
	args, err := newArguments(c)
	if err != nil {
		return err
	}
	var pathOption clientcmd.ConfigAccess = clientcmd.NewDefaultPathOptions()
	conf, err := pathOption.GetStartingConfig()
	if err != nil {
		return err
	}
	startingCluster, exists := conf.Clusters[args.clusterName]
	if !exists {
		startingCluster = clientcmdapi.NewCluster()
	}
	cluster := modifyCluster(*args, *startingCluster)
	conf.Clusters[args.clusterName] = &cluster
	if err := clientcmd.ModifyConfig(pathOption, *conf, true); err != nil {
		return err
	}

	startingAuthInfo, exists := conf.AuthInfos[args.userName]
	if !exists {
		startingAuthInfo = clientcmdapi.NewAuthInfo()
	}
	authInfo := modifyAuthInfo(*args, *startingAuthInfo)
	conf.AuthInfos[args.userName] = &authInfo
	if err := clientcmd.ModifyConfig(pathOption, *conf, true); err != nil {
		return err
	}

	startingContext, exists := conf.Contexts[args.contextName]
	if !exists {
		startingContext = clientcmdapi.NewContext()
	}
	k8sContext := modifyContext(*args, *startingContext)
	conf.Contexts[args.contextName] = &k8sContext
	conf.CurrentContext = args.contextName
	if err := clientcmd.ModifyConfig(pathOption, *conf, true); err != nil {
		return err
	}
	return nil
}

func modifyCluster(args arguments, existingCluster clientcmdapi.Cluster) clientcmdapi.Cluster {
	modifiedCluster := existingCluster
	modifiedCluster.Server = args.clusterUrl
	modifiedCluster.InsecureSkipTLSVerify = args.tlsVerify
	if modifiedCluster.InsecureSkipTLSVerify {
		modifiedCluster.CertificateAuthority = ""
		modifiedCluster.CertificateAuthorityData = nil
	}

	return modifiedCluster
}

func modifyAuthInfo(args arguments, existingAuthInfo clientcmdapi.AuthInfo) clientcmdapi.AuthInfo {
	modifiedAuthInfo := existingAuthInfo
	if modifiedAuthInfo.AuthProvider == nil || modifiedAuthInfo.AuthProvider.Name != args.providerName {
		modifiedAuthInfo.AuthProvider = &clientcmdapi.AuthProviderConfig{
			Name: args.providerName,
		}
	}
	if modifiedAuthInfo.AuthProvider != nil {
		if modifiedAuthInfo.AuthProvider.Config == nil {
			modifiedAuthInfo.AuthProvider.Config = make(map[string]string)
		}

		for key, value := range args.getProviderMap() {
			modifiedAuthInfo.AuthProvider.Config[key] = value
		}
	}
	return modifiedAuthInfo
}

func modifyContext(args arguments, existingContext clientcmdapi.Context) clientcmdapi.Context {
	modifiedContext := existingContext
	modifiedContext.Cluster = args.clusterName
	modifiedContext.AuthInfo = args.userName
	modifiedContext.Namespace = args.namespace
	return modifiedContext
}
