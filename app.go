package main

import (
	"io/ioutil"

	"github.com/seletskiy/hierr"
)

type app struct {
	config *config
	oauth  map[string]*OAuth
	db     *database
	tokens *OAuthTokensTable
}

func newApp(config *config) (*app, error) {
	oauth := map[string]*OAuth{}
	for _, oauthData := range config.OAuth {
		keyData, err := ioutil.ReadFile(oauthData.Key)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				"can't read oauth key file for server %s",
				oauthData.Server,
			)
		}

		server, err := NewOAuth(
			oauthData.Name,
			oauthData.Server,
			oauthData.SessionURL,
			oauthData.UserURL,
			oauthData.Consumer,
			keyData,
			config.Web.URL,
		)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				"can't obtain oauth instance for server %s",
				oauthData.Server,
			)
		}

		oauth[oauthData.Name] = server
	}

	tokens, err := NewOAuthTokensTable()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't obtain oauth tokens table instance",
		)
	}

	err = tokens.Init()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't initialize oauth tokens table",
		)
	}

	db := &database{
		dsn: config.Database.Address,
	}

	err = db.connect()
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't establish database connection",
		)
	}

	app := &app{
		config: config,
		oauth:  oauth,
		tokens: tokens,
		db:     db,
	}

	return app, nil
}
