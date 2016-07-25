:bitbucket
:mongod
:nucleus

:bitbucket-set-response <<X
200 OK

oauth_token=oauth-token&oauth_token_secret=oauth-token-secret
X
tests:ensure :request /login/bitbucket-slug/
tests:assert-stderr \
    "Location: http://$_bitbucket/authorize-token?oauth_token=oauth-token"
