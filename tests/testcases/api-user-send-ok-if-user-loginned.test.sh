:bitbucket
:mongod
:nucleus

tests:clone testcases/create-new-user-after-login.test.sh .
NO_DAEMONS=1 source create-new-user-after-login.test.sh

tests:ensure :request /api/v1/user
tests:assert-stderr '200 OK'
