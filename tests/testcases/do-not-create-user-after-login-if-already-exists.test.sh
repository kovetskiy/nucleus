:bitbucket
:mongod
:nucleus

tests:clone testcases/create-new-user-after-login.test.sh .

NO_DAEMONS=1 source create-new-user-after-login.test.sh

tests:ensure mongo $_mongod/nucleus --quiet --eval "db.tokens.find().pretty();"

NO_DAEMONS=1 original=$(cat $(tests:get-stdout-file))

source create-new-user-after-login.test.sh

tests:ensure mongo $_mongod/nucleus --quiet --eval "db.tokens.find().pretty();"
tests:assert-no-diff $(tests:get-stdout-file) "$original"
