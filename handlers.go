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
		token, _ = x.Request.Cookie("token")
	)

	if token == nil {
		return false
	}

	user, err := app.getUserByToken(token.Value)
	if err == mgo.ErrNotFound {
		return false
	}

	if err != nil {
		errorln(
			hierr.Errorf(
				err, "can't obtain user with token '%s'",
				token.Value,
			),
		)
		x.AbortWithStatus(http.StatusInternalServerError)
		return false
	}

	x.Set("user", user)

	return true
}

func (app *app) handleIndex(x *gin.Context) {
	if app.checkAccess(x) {
		render(x, layoutToken)
	} else if !x.IsAborted() {
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

	http.SetCookie(x.Writer, &http.Cookie{
		Name:   "token",
		Value:  token,
		Secure: true,
	})
}

func (app *app) handleAccess(x *gin.Context) {
	if app.checkAccess(x) {
		x.AbortWithStatus(http.StatusOK)
	} else {
		x.AbortWithStatus(http.StatusUnauthorized)
	}
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

	user, err := app.getUserByusername(username)
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
		userinfo, err := oauth.GetRequest(
			oauth.UserURL,
			map[string]string{"username": username},
			accessToken,
		)
		if err != nil {
			errorln(
				hierr.Errorf(
					err, "can't execute OAuth request to '%s' (%s)",
					oauth.UserURL, oauth.Server,
				),
			)
			x.AbortWithStatus(http.StatusBadGateway)
			return
		}

		token, err = app.adduser(username, userinfo)
		if err != nil {
			errorln(
				hierr.Errorf(
					err, "can't update userinfo for user '%s'",
					username,
				),
			)
			x.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		token = user.Token
	}

	http.SetCookie(x.Writer, &http.Cookie{
		Name:   "token",
		Value:  token,
		Secure: true,
	})

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
				oauth.SessionURL, oauth.Server,
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
	_, err := app.db.tokens.Upsert(bson.M{
		"username": username,
	}, bson.M{
		"$set": bson.M{
			"token":      token,
			"token_date": time.Now().Unix(),
		},
	})

	return err
}

func (app *app) generateToken() string {
	return fmt.Sprintf(
		"%x",
		md5.Sum([]byte(fmt.Sprint(time.Now().UnixNano(), rand.Int63()))),
	)
}

func (app *app) adduser(
	username string,
	userinfo map[string]interface{},
) (string, error) {
	token := app.generateToken()

	err := app.db.tokens.Insert(bson.M{
		"$set": bson.M{
			"username":    username,
			"userinfo":    userinfo,
			"token":       token,
			"token_date":  time.Now().Unix(),
			"create_date": time.Now().Unix(),
		},
	})

	return token, err
}

func (app *app) getUserByToken(token string) (user, error) {
	var resource user

	err := app.db.tokens.Find(
		bson.M{"token": token},
	).One(&resource)
	if err != nil {
		return user{}, err
	}

	return resource, nil
}

func (app *app) getUserByusername(username string) (user, error) {
	var resource user

	err := app.db.tokens.Find(
		bson.M{"username": username},
	).One(&resource)
	if err != nil {
		return user{}, err
	}

	return resource, nil
}
