nucleus(1) -- user authentification tokens distribution service
============

## DESCRIPTION

*nucleus* is service for distribution user authentification tokens basing on
third-party OAuth services.

*nucleus* don't know how user will be authentificated, *nucleus* know only one -
user can be authentificated and when new user visits index page, *nucleus* will
give to user a choose of how he will authentificate basing on OAuth providers
that described in configuration file. 

When user click at a login button 'Authorize using %service%', *nucleus* will
redirect user to service's authorization gateway and after successfull/failed
authorization user will be redirected to *nucleus* callback URL.

At that moment *nucleus* will create unique access token for specified user (if
not created already), that token will be used for next authentifications
between user and *nucleus*.

If another external service supports authentification using *nucleus* than user
will pass secret token to that service as password, and that service will go to
the nucleus and send provided token, nucleus will token and send information
about user, if specified token really exists in database.

This method provides way for fast authentification, because *nucleus* doesn't
communicates with external services for checking user access. User authorizes
only once and uses authentification token everywhere.

## SYNOPSIS

    nucleus [--config <path>]
    nucleus -h | --help
    nucleus --version

## OPTIONS

**-c, --config <path>**  
Use specified configuration file instead of default
`/etc/nucleus/nucleus.conf`.

**-h, --help**  
Show program help.

**--version**  
Show version of program. This value can be equal to `[manual build]` if
nucleus has been installed using `go get` or `go install` instead of using
system-wide package manager. Don't be a stupid, use packages, it will save
human resources in future.

## CONFIGURATION

*nucleus* must be configured before starting, configuration file should be
located in `/etc/nucleus/nucleus.conf`, but this path can be changed using
`--config` flag, configuration file must be written in TOML format using
following template:

    [web]
      listen = ":80"
    [database]
      url = "mongodb://localhost/dbname"
    [[oauth]]
      name                = "Enterprise Service"
      slug                = "office"
      basic_url           = "http://intranet.local/"
      consumer            = "consumer"
      key_file            = "/etc/nucleus/intranet.key"
      session_url         = "/api/session"
      user_url            = "/api/user"
      request_token_url   = "/api/token/request"
      authorize_token_url = "/api/token/authorize"
      access_token_url    = "/api/token/access"


As you can see, you must pass data source name for mongodb database in
specified format:

    mongodb://[user@]hostname[:port][,[user2@]hostname2[:port2]]/[database]

Of course, **nucleus** will not crash if any of database host dies down,
**nucleus** will ping database, properly catch database error if any, and try
to reestablish database connection.

## REPRESENTATIONAL STATE TRANSFER APPLICATION INTERFACE

### GET /api/v1/user

Authentificate user and retrieve information about user.

Authentification token should be passed as Basic Authorization password or
using Cookie `token` header.

**RESPONSE STATUSES**

* **401 Unathorized**  
    User with specified token doesn't exists or token has been revoked by user.

* **200 OK**  
    User authentificated.

## MAINTAINER

Egor Kovetskiy <e.kovetskiy@office.ngs.ru>
