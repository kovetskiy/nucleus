:bitbucket
:mongod
:nucleus

tests:clone testcases/create-new-user-after-login.test.sh .

NO_DAEMONS=1 source create-new-user-after-login.test.sh

tests:ensure mongo $_mongod/nucleus --quiet --eval "db.tokens.find().pretty();"

token_doc=$(cat $(tests:get-stdout-file))

NO_DAEMONS=1 source create-new-user-after-login.test.sh

tests:ensure mongo $_mongod/nucleus --quiet --eval "db.tokens.find().pretty();"
tests:assert-no-diff $(tests:get-stdout-file) "$token_doc"
