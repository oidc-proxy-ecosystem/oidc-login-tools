package config

import "errors"

type Token struct {
	IdToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (t *Token) Valid() error {
	errFn := func(name string) error {
		return errors.New("%sが空文字です。")
	}
	if t.AccessToken == "" {
		return errFn("access token")
	}
	if t.IdToken == "" {
		return errFn("id token")
	}
	if t.RefreshToken == "" {
		return errFn("refresh token")
	}
	return nil
}
