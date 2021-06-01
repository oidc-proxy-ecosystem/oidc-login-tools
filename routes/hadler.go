package routes

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/oidc-proxy-ecosystem/oidc-tools/auth"
	"github.com/oidc-proxy-ecosystem/oidc-tools/config"
)

type Process interface {
	Action(certFile, keyFile string) error
}

type handler struct {
	mux           *http.ServeMux
	log           *log.Logger
	state         string
	authenticator *auth.Authenticator
	oidc          *config.Oidc
	server        *http.Server
	port          int
}

func (h *handler) Action(certFile, keyFile string) error {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	state := base64.StdEncoding.EncodeToString(b)
	u := h.authenticator.Config.AuthCodeURL(state, h.oidc.SetValues()...)
	h.log.Printf("open browser: %s", u)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		return err
	}
	h.server = &http.Server{
		Handler: h.mux,
		Addr:    listener.Addr().String(),
	}
	h.state = state
	if err := h.server.ServeTLS(listener, certFile, keyFile); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func New(port int, oidc *config.Oidc, authenticator *auth.Authenticator) Process {
	h := &handler{
		log:           log.Default(),
		oidc:          oidc,
		authenticator: authenticator,
		port:          port,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/callback", h.callback)
	h.mux = mux
	return h
}
