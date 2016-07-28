tests:clone ../nucleus.test bin/
tests:clone bitbucket.mock bin/
tests:clone ../templates/ .

tests:ensure chmod +x bin/bitbucket.mock

_mongod="127.0.0.1:64999"
_nucleus="127.0.0.1:64888"
_bitbucket="127.0.0.1:64777"

:mongod() {
    tests:make-tmp-dir db
    tests:run-background mongod_background \
        mongod --quiet --dbpath $(tests:get-tmp-dir)/db --port 64999
}

:bitbucket() {
    tests:value _blankd \
        $(which blankd) \
        -l "$_bitbucket" \
        -o $(tests:get-tmp-dir)/blankd.log \
        -d $(tests:get-tmp-dir)/ \
        -e $(tests:get-tmp-dir)/bin/bitbucket.mock
    tests:put-string _blankd_process "$_blankd"
}

:bitbucket-set-response() {
    tests:put bitbucket_response${1:+_$1}
}

:nucleus() {
    tests:put nucleus.conf <<CONF
[web]
    listen          = "$_nucleus"
    url             = "https://$_nucleus/"
    tls_key         = "$(tests:get-tmp-dir)/nucleus.key"
    tls_certificate = "$(tests:get-tmp-dir)/nucleus.cert"
[[oauth]]
    slug                = "bitbucket-slug"
    name                = "bitbucket-name"
    basic_url           = "http://$_bitbucket"
    consumer            = "bitbucket-consumer"
    key_file            = "$(tests:get-tmp-dir)/bitbucket.key"
    session_url         = "/session"
    user_url            = "/user"
    request_token_url   = "/request-token"
    authorize_token_url = "/authorize-token"
    access_token_url    = "/access-token"
[database]
    address = "mongodb://$_mongod/nucleus"
CONF

    tests:ensure openssl req \
        -batch \
        -new \
        -newkey rsa:1024 \
        -days 365 \
        -nodes \
        -x509 \
        -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=localhost" \
        -keyout $(tests:get-tmp-dir)/nucleus.key \
        -out  $(tests:get-tmp-dir)/nucleus.cert

    tests:ensure \
        ssh-keygen -q -t rsa -b 1024 -f $(tests:get-tmp-dir)/bitbucket.key

    tests:run-background nucleus_background nucleus.test \
        --config $(tests:get-tmp-dir)/nucleus.conf
    sleep 1
}

:request() {
    tests:eval curl -s -v --insecure -b cookies -c cookies \
        "https://$_nucleus$1" "${@:2}"

    cat $(tests:get-stdout-file)
    cat $(tests:get-stderr-file) >&2
    return $(tests:get-exitcode)
}
