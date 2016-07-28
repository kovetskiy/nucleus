package main

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/gin-gonic/gin"
	"github.com/mrjones/oauth"
	"github.com/seletskiy/hierr"
)

var (
	layoutToken = "token.html"
	layoutLogin = "login.html"
)

func (app *app) checkAccess(x *gin.Context) bool {
	var (
		cookie, _       = x.Cookie("token")
		field, _        = x.GetPostForm("token")
		_, basicAuth, _ = x.Request.BasicAuth()
	)

	var token string
	if cookie != "" {
		token = cookie
	}

	if field != "" {
		token = field
	}

	if basicAuth != "" {
		token = basicAuth
	}

	if token == "" {
		tracef("no token specified")
		return false
	}

	user, err := app.getUser("token", token)
	if err == mgo.ErrNotFound {
		tracef("unknown token")
		return false
	}

	if err != nil {
		errorln(
			hierr.Errorf(
				err, "can't obtain user with token '%s'",
				token,
			),
		)
		x.AbortWithStatus(http.StatusInternalServerError)
		return false
	}

	x.Set("user", user)

	tracef("valid token, user found")

	return true
}

func (app *app) handleIndex(x *gin.Context) {
	if app.checkAccess(x) {
		render(x, layoutToken)
	} else if !x.IsAborted() {
		x.Set("oauth", app.config.OAuth)
		render(x, layoutLogin)
	}
}

func (app *app) handleGenerateToken(x *gin.Context) {
	if !app.checkAccess(x) {
		x.AbortWithStatus(http.StatusForbidden)
		return
	}

	assert(
		x.MustGet("user") == nil,
		"user has access but field doesn't exist",
	)

	var (
		token    = app.generateToken()
		username = x.MustGet("user").(user).Name
	)

	err := app.updateToken(username, token)
	if err != nil {
		errorln(
			hierr.Errorf(
				err,
				"can't update token for user '%s'",
				username,
			),
		)
		x.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	tracef("set cookie token=%s", token)

	x.SetCookie("token", token, int(time.Hour), "/", "", false, false)

	x.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}

func (app *app) handleUser(x *gin.Context) {
	if app.checkAccess(x) {
		assert(
			x.MustGet("user") == nil,
			"user has access but field doesn't exist",
		)

		x.IndentedJSON(http.StatusOK, x.MustGet("user"))
		x.Abort()
	} else {
		x.AbortWithStatus(http.StatusUnauthorized)
	}
}

func (app *app) handleLogin(x *gin.Context) {
	oauth, accessToken, username := app.handleAuthentificate(x)
	if x.IsAborted() {
		return
	}

	assert(
		username == "",
		"context is not aborted, but username is empty",
	)

	assert(
		oauth == nil,
		"context is not aborted, but oauth is nil",
	)

	assert(
		accessToken == nil,
		"context is not aborted, but access token is nil",
	)

	user, err := app.getUser("username", username)
	if err != nil && err != mgo.ErrNotFound {
		errorln(
			hierr.Errorf(
				err, "can't obtain user '%s'", username,
			),
		)
		x.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var token string
	if err == mgo.ErrNotFound {
		tracef("creating new user")

		userinfo, err := oauth.GetRequest(
			oauth.UserURL,
			map[string]string{"username": username},
			accessToken,
		)
		if err != nil {
			errorln(
				hierr.Errorf(
					err, "can't execute OAuth request to '%s' (%s)",
					oauth.UserURL, oauth.BasicURL,
				),
			)
			x.AbortWithStatus(http.StatusBadGateway)
			return
		}

		token, err = app.addUser(username, userinfo)
		if err != nil {
			errorln(
				hierr.Errorf(
					err, "can't add user '%s' to database",
					username,
				),
			)
			x.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		tracef("user already created, token is '%s'", user.Token)
		token = user.Token
	}

	tracef("set cookie token=%s", token)

	x.SetCookie(
		"token", token, int(time.Hour), "/", "", false, false,
	)
	x.Redirect(http.StatusTemporaryRedirect, "/")
	x.Abort()

	return
}

func (app *app) handleAuthentificate(
	x *gin.Context,
) (*OAuth, *oauth.AccessToken, string) {
	var (
		verifier = x.Query("oauth_verifier")
		token    = x.Query("oauth_token")
		provider = x.Param("provider")
	)

	oauth, ok := app.oauth[provider]
	if !ok {
		x.AbortWithStatus(http.StatusNotFound)
		return nil, nil, ""
	}

	if verifier == "" {
		requestToken, redirectURL, err := oauth.GetRequestTokenAndURL()
		if err != nil {
			errorln(
				hierr.Errorf(
					err,
					"can't get request token and url for provider %s",
					provider,
				),
			)
			x.AbortWithStatus(http.StatusInternalServerError)
			return nil, nil, ""
		}

		app.tokens.Add(requestToken)

		if redirectURL != "" {
			x.Redirect(
				http.StatusTemporaryRedirect, redirectURL,
			)
			x.Abort()
			return nil, nil, ""
		}
	}

	requestToken := app.tokens.Get(token)
	if requestToken == nil {
		x.AbortWithStatus(http.StatusBadRequest)
		return nil, nil, ""
	}

	accessToken, err := oauth.GetAccessToken(requestToken, verifier)
	if err != nil {
		errorln(
			hierr.Errorf(
				err,
				"can't get access token for request token "+
					"'%s' and oauth verifier '%s'",
				requestToken, verifier,
			),
		)
		x.AbortWithStatus(http.StatusInternalServerError)
		return nil, nil, ""
	}

	response, err := oauth.GetRequest(
		oauth.SessionURL, map[string]string{}, accessToken,
	)
	if err != nil {
		errorln(
			hierr.Errorf(
				err, "can't execute OAuth request to '%s' (%s)",
				oauth.SessionURL, oauth.BasicURL,
			),
		)
		x.AbortWithStatus(http.StatusBadGateway)
		return nil, nil, ""
	}

	username, ok := response["name"].(string)
	if !ok {
		errorln(
			hierr.Errorf(
				err,
				"ambigious oauth server response, key 'name' doesn't exists",
			),
		)
		x.AbortWithStatus(http.StatusBadGateway)
		return nil, nil, ""
	}

	return oauth, accessToken, username
}

func (app *app) updateToken(username string, token string) error {
	payload := bson.M{
		"$set": bson.M{
			"token":      token,
			"token_date": time.Now().Unix(),
		},
	}

	tracef("update token for user '%s': %#v", payload)

	err := app.db.tokens.Update(bson.M{"username": username}, payload)
	return err
}

func (app *app) generateToken() string {
	return fmt.Sprintf(
		"%x",
		md5.Sum([]byte(fmt.Sprint(time.Now().UnixNano(), rand.Int63()))),
	)
}

func (app *app) addUser(
	username string,
	userinfo map[string]interface{},
) (string, error) {
	token := app.generateToken()
	payload := bson.M{
		"username":    username,
		"userinfo":    userinfo,
		"token":       token,
		"token_date":  time.Now().Unix(),
		"create_date": time.Now().Unix(),
	}

	tracef("adding user %#v", payload)

	err := app.db.tokens.Insert(payload)
	return token, err
}

func (app *app) getUser(key, value string) (user, error) {
	tracef("get user by %s '%s'", key, value)

	var resource user
	err := app.db.tokens.Find(bson.M{key: value}).One(&resource)

	tracef("resource: %#v", resource)

	return resource, err
}
