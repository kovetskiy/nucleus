package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"strings"

	"github.com/mrjones/oauth"
	"github.com/seletskiy/hierr"
)

type OAuth struct {
	Name        string
	HomeURL     string
	Server      string
	SessionURL  string
	UserURL     string
	RedirectURL string
	Consumer    *oauth.Consumer
}

func NewOAuth(
	name string,
	server string,
	sessionURL string,
	userURL string,
	consumerKey string,
	privatePEMKey []byte,
	homeURL string,
) (*OAuth, error) {
	// second return value is not error
	decodedKey, _ := pem.Decode(privatePEMKey)

	privateKey, err := x509.ParsePKCS1PrivateKey(decodedKey.Bytes)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't parse private key",
		)
	}

	server = strings.TrimRight(server, "/")

	consumer := oauth.NewRSAConsumer(
		consumerKey,
		privateKey,
		oauth.ServiceProvider{
			RequestTokenUrl:   server + "/plugins/servlet/oauth/request-token",
			AuthorizeTokenUrl: server + "/plugins/servlet/oauth/authorize",
			AccessTokenUrl:    server + "/plugins/servlet/oauth/access-token",
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
		Name:       name,
		SessionURL: sessionURL,
		UserURL:    userURL,
		HomeURL:    homeURL,
		Server:     server,
		Consumer:   consumer,
	}, nil
}

func (oauth *OAuth) GetRequestTokenAndURL() (
	*oauth.RequestToken, string, error,
) {
	loginURL := oauth.HomeURL + "/login/" + oauth.Name

	return oauth.Consumer.GetRequestTokenAndUrl(loginURL)
}

func (oauth *OAuth) GetAccessToken(
	requestToken *oauth.RequestToken, verifier string,
) (*oauth.AccessToken, error) {
	return oauth.Consumer.AuthorizeToken(requestToken, verifier)
}

func (oauth *OAuth) GetRequest(
	url string, params map[string]string, accessToken *oauth.AccessToken,
) (map[string]interface{}, error) {
	rawResponse, err := oauth.Consumer.Get(
		oauth.Server+url, params, accessToken,
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
