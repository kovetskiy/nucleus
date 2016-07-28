:bitbucket
:mongod
:nucleus

tests:clone testcases/create-new-user-after-login.test.sh .
NO_DAEMONS=1 source create-new-user-after-login.test.sh

tests:value token \
    mongo --quiet --eval  "db.tokens.find({}, {token:1, _id:0});" \
        $_mongod/nucleus '|' jq -r '.token'

tests:ensure :request /token -X POST
tests:assert-stderr '200 OK'
tests:assert-stdout-re '{"token":"([\w\d]{32})"}'

tests:value new_token \
    mongo --quiet --eval  "db.tokens.find({}, {token:1, _id:0});" \
        $_mongod/nucleus '|' jq -r '.token'

tests:assert-test "$token" '!=' "$new_token"
