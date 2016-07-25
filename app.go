package main

import "github.com/seletskiy/hierr"

type app struct {
	config *config
	oauth  map[string]*OAuth
	db     *database
	tokens *OAuthTokensTable
}

func newApp(config *config) (*app, error) {
	oauth := map[string]*OAuth{}
	for _, oauthData := range config.OAuth {
		server, err := NewOAuth(
			oauthData,
			config.Web.URL,
		)
		if err != nil {
			return nil, hierr.Errorf(
				err,
				"can't obtain oauth instance for server %s",
				oauthData.BasicURL,
			)
		}

		oauth[oauthData.Slug] = server
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

	infof("tokens table initialized")

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

	infof("database connection established")

	app := &app{
		config: config,
		oauth:  oauth,
		tokens: tokens,
		db:     db,
	}

	return app, nil
}
