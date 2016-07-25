:bitbucket
:mongod
:nucleus

tests:ensure :request /
tests:assert-stdout-re "/login/bitbucket-slug/"
tests:assert-stdout-re "bitbucket-name"
