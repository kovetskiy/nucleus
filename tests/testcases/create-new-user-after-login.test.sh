if [[ ! "${NO_DAEMONS:-}" ]]; then
    :bitbucket
    :mongod
    :nucleus
fi

:bitbucket-set-response <<X
200 OK

oauth_token=oauth-token&oauth_token_secret=oauth-token-secret
X
tests:ensure :request '/login/bitbucket-slug/'

:bitbucket-set-response <<X
200 OK

oauth_token=oauth-token&oauth_token_secret=oauth-token-secret
X

:bitbucket-set-response 2 <<X
200 OK

{"name":"john"}
X

:bitbucket-set-response 3 <<X
200 OK

{"info-key":"info-value"}
X

tests:ensure :request \
    '/login/bitbucket-slug/?oauth_token=oauth-token&oauth_verifier=oauth-verifier'
tests:assert-stderr-re "Location: /"
tests:assert-stderr-re "Set-Cookie: token=([\w\d]{32});"

tests:ensure mongo $_mongod/nucleus \
    <<< "db.tokens.find({}).pretty();"
tests:assert-stdout-re '"token" : "([\w\d]{32})"'
tests:assert-stdout-re '"username" : "john"'
tests:assert-stdout-re '"info-key" : "info-value"'
tests:assert-stdout-re '"create_date" : NumberLong\(\d+\)'
tests:assert-stdout-re '"token_date" : NumberLong\(\d+\)'
