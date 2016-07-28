:bitbucket
:mongod
:nucleus

tests:ensure :request /api/v1/user
tests:assert-stderr-re '401 Unauthorized'
