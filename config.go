package main

import "github.com/kovetskiy/ko"

type OAuthConfig struct {
	Name              string `toml:"name" required:"true"`
	Slug              string `toml:"slug" required:"true"`
	BasicURL          string `toml:"basic_url" required:"true"`
	Consumer          string `toml:"consumer" required:"true"`
	KeyFile           string `toml:"key_file" required:"true"`
	SessionURL        string `toml:"session_url" required:"true"`
	UserURL           string `toml:"user_url" required:"true"`
	RequestTokenURL   string `toml:"request_token_url" required:"true"`
	AuthorizeTokenURL string `toml:"authorize_token_url" required:"true"`
	AccessTokenURL    string `toml:"access_token_url" required:"true"`
}

type config struct {
	Web struct {
		Listen         string `toml:"listen" required:"true"`
		URL            string `toml:"url" required:"true"`
		TLSKey         string `toml:"tls_key" required:"true"`
		TLSCertificate string `toml:"tls_certificate" required:"true"`
	} `toml:"web" required:"true"`

	OAuth []OAuthConfig `toml:"oauth" required:"true"`

	Database struct {
		Address string `toml:"address" required:"true"`
	} `toml:"database" required:"true"`
}

func getConfig(path string) (*config, error) {
	config := &config{}
	err := ko.Load(path, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
