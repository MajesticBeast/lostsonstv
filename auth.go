package main

import (
	disgoauth "github.com/realTristan/disgoauth"
)

func NewDiscordClient(id, secret, url string) *disgoauth.Client {
	dc := disgoauth.Init(&disgoauth.Client{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURI:  url,
		Scopes:       []string{disgoauth.ScopeIdentify},
	})

	return dc
}
