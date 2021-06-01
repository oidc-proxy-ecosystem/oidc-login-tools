package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
	"github.com/oidc-proxy-ecosystem/oidc-tools/config"
)

func (h *handler) callback(w http.ResponseWriter, r *http.Request) {
	defer h.server.Shutdown(context.Background())
	ctx := r.Context()
	if r.URL.Query().Get("state") != h.state {
		log.Fatalln("stateが一致しません。")
	}

	token, err := h.authenticator.Config.Exchange(ctx, r.URL.Query().Get("code"), h.oidc.SetValues()...)
	if err != nil {
		log.Fatalln(fmt.Sprintf("no token found: %v", err))
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Fatalln("No id_token field in oauth2 token.")
	}

	oidcConfig := &oidc.Config{
		ClientID: h.oidc.ClientId,
	}

	_, err = h.authenticator.Provider.Verifier(oidcConfig).Verify(ctx, rawIDToken)

	if err != nil {
		log.Fatalln("Failed to verify ID Token: " + err.Error())
	}
	tokens := &config.Token{
		IdToken:      rawIDToken,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	file, err := os.Create("./auth.json")
	if err != nil {
		log.Fatalln("auth.jsonの生成に失敗しました。")
	}
	defer file.Close()
	buf, _ := json.Marshal(tokens)
	file.Write(buf)
	w.WriteHeader(http.StatusOK)
}
