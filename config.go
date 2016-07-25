package main

import "github.com/kovetskiy/ko"

type config struct {
	Web struct {
		Listen         string `toml:"listen" required:"true"`
		URL            string `toml:"url" required:"true"`
		TLSKey         string `toml:"tls_key" required:"true"`
		TLSCertificate string `toml:"tls_certificate" required:"true"`
	} `toml:"web" required:"true"`

	OAuth []struct {
		Name       string `toml:"name" required:"true"`
		Server     string `toml:"server" required:"true"`
		Consumer   string `toml:"consumer" required:"true"`
		Key        string `toml:"key" required:"true"`
		SessionURL string `toml:"session_url" required:"true"`
		UserURL    string `toml:"user_url" required:"true"`
	} `toml:"oauth" required:"true"`

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
