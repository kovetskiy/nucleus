package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mrjones/oauth"
	"github.com/seletskiy/hierr"
)

type OAuth struct {
	OAuthConfig
	homeURL  string
	consumer *oauth.Consumer
}

func NewOAuth(config OAuthConfig, homeURL string) (*OAuth, error) {
	keyData, err := ioutil.ReadFile(config.KeyFile)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't read key file",
		)
	}

	// second return value is not error
	decodedKey, _ := pem.Decode(keyData)

	privateKey, err := x509.ParsePKCS1PrivateKey(decodedKey.Bytes)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't parse private key",
		)
	}

	basicURL := strings.TrimRight(config.BasicURL, "/")

	consumer := oauth.NewRSAConsumer(
		config.Consumer,
		privateKey,
		oauth.ServiceProvider{
			RequestTokenUrl:   basicURL + config.RequestTokenURL,
			AuthorizeTokenUrl: basicURL + config.AuthorizeTokenURL,
			AccessTokenUrl:    basicURL + config.AccessTokenURL,
			HttpMethod:        "POST",
		},
	)

	consumer.Debug(false)

	consumer.HttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	return &OAuth{
		OAuthConfig: config,
		homeURL:     homeURL,
		consumer:    consumer,
	}, nil
}

func (oauth *OAuth) GetRequestTokenAndURL() (
	*oauth.RequestToken, string, error,
) {
	loginURL := strings.TrimRight(oauth.homeURL, "/") +
		"/login/" + oauth.Slug + "/"

	return oauth.consumer.GetRequestTokenAndUrl(loginURL)
}

func (oauth *OAuth) GetAccessToken(
	requestToken *oauth.RequestToken, verifier string,
) (*oauth.AccessToken, error) {
	return oauth.consumer.AuthorizeToken(requestToken, verifier)
}

func (oauth *OAuth) GetRequest(
	url string, params map[string]string, accessToken *oauth.AccessToken,
) (map[string]interface{}, error) {
	rawResponse, err := oauth.consumer.Get(
		strings.TrimRight(oauth.BasicURL, "/")+url, params, accessToken,
	)
	if err != nil {
		return nil, err
	}

	defer rawResponse.Body.Close()

	response := map[string]interface{}{}
	err = json.NewDecoder(rawResponse.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
