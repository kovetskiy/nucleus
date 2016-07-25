package main

import (
	"time"

	"github.com/mrjones/oauth"
)

type OAuthToken struct {
	Created time.Time
	Token   *oauth.RequestToken
}

type OAuthTokensTable struct {
	tokens map[string]OAuthToken
}

func NewOAuthTokensTable() (*OAuthTokensTable, error) {
	return &OAuthTokensTable{
		tokens: map[string]OAuthToken{},
	}, nil
}

func (table *OAuthTokensTable) Add(token *oauth.RequestToken) {
	table.tokens[token.Token] = OAuthToken{
		Created: time.Now(),
		Token:   token,
	}
}

func (table *OAuthTokensTable) Get(token string) *oauth.RequestToken {
	oauthToken, ok := table.tokens[token]
	if !ok {
		return nil
	}

	return oauthToken.Token
}

func (table *OAuthTokensTable) Init() error {
	go func(table *OAuthTokensTable) {
		for {
			for index, token := range table.tokens {
				if time.Now().Unix()-token.Created.Unix() > 600 {
					delete(table.tokens, index)
				}
			}

			time.Sleep(time.Minute)
		}
	}(table)

	return nil
}
