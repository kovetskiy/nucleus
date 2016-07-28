:bitbucket
:mongod
:nucleus

tests:clone testcases/create-new-user-after-login.test.sh .
NO_DAEMONS=1 source create-new-user-after-login.test.sh

tests:ensure rm cookies

tests:value token \
    mongo --quiet --eval  "db.tokens.find({}, {token:1, _id:0});" \
        $_mongod/nucleus '|' jq -r '.token'

tests:ensure :request /api/v1/user --basic --user "nomatters:$token"
tests:assert-stderr '200 OK'
