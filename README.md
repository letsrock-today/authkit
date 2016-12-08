[![Build Status](https://travis-ci.org/letsrock-today/authkit.svg?branch=master)](https://travis-ci.org/letsrock-today/authkit)

# authkit

TL;DR: may be you will be happy with one of following:

- https://auth0.com/
- http://www.janrain.com/
- https://github.com/golang/oauth2
- https://github.com/markbates/goth
- https://github.com/ory-am/hydra

__Status__: unstable, experimental, probably unsecure.

"authkit" is a set of reusable peaces of code (HTTP handlers, middleware, helpers
and other primitives), mostly on Go programming language, useful to implement
OAuth 2 authorization and SSO at the server-side, accompanied with relatively
complete demo implementation.

Package is not a framework, yet another OAuth2 provider or OAuth2 client library.
Instead of trying to create more general library, we are striving to focus
on the scenarios we currently need and use in our own projects, prototype them
here and extract reusable code into library package. Project can be seen as an
example, blueprint and related glue code for particular authorization scenario(s).

Before we started this project, we already implemented some relevant parts in
several different projects. We wanted to add Hydra support and gather all
scattered pieces in one place, somewhat brush and simplify messy parts, etc.

So, we started this project from porting and tailoring
[Hydra usage example](https://github.com/ory-am/hydra-idp-react). It may be seen
as more elaborated usage example of Hydra, oauth2, Echo and related stuff. And 
later we start to add our other existing code here, experimenting, rethinking
and refactoring it on the way.

Mainly, we are focused on the following scenario:

- we have custom http API;
- we want this API or part of it be available only to authorized users;
- we want that users be able to authorize using their existing social network
  accounts (Facebook, Google+, LinkedIn, etc);
- we want to be able to provide username/password login as well;
- we want that no matter which type of login user choose, API would be protected
  using single approach with access token, issued by our side;
- we want to be able to make our API or parts of it accessible to third parties
  via OAuth2;
- we want to expose single IP address to our clients (or several equivalent
  addresses), so, for example, we want to proxy requests to Hydra via Nginx or
  our own application.

So, our side in this context controls resource owner, OAuth2 client and provider.

This project uses:

- https://github.com/labstack/echo, as a server framework (to implement routes,
  http handlers, middleware, etc); "echo" designed to be compatible with
  gorilla, negroni, standard lib, so, it should be possible to use our lib
  with them either (have not tested yet, though);
- https://github.com/golang/oauth2, as a client OAuth2 library;
- https://github.com/ory-am/hydra, as a OAuth2 provider;
- https://github.com/dgrijalva/jwt-go, to create tokens, when we have to;
- https://github.com/asaskevich/govalidator, for validation logic;
- https://github.com/mitchellh/mapstructure, to convert maps into structures;
- https://github.com/stretchr/testify, for testing and mocking;
- https://github.com/h2non/gock, to mock HTTP server;
- ...

Our aim is to provide:

- [x] a set of server-side primitives (http handlers, routes, middleware, helpers)
      usable to implement different oauth2 scenarios;
- [x] sample web app implementation using [Riot.js](http://riotjs.com/);
- [ ] binary auth module for Nginx (see http://nginx.org/en/docs/http/ngx_http_auth_request_module.html);
- [ ] example of Nginx module usage;
- [ ] sample Android app implementation;

Finally, when the lib is ready, we want to migrate our own application
(https://letsrock.today/) to it and to enjoy benefits of going open-source.

At the time of writing this text we are using Go 1.7.
We use Glide to manage dependencies.

See ./sample/authkit folder for the first demo app. Instructions to run samples
are below.


# Getting started

## Prerequisites

See ./.travis.yml for concise test/dev environment description.

Below are listed versions of tools we used in dev equivalent:

- Ubuntu 16.04.1 LTS
- go 1.7
- docker 1.11
- docker-compose 1.8.1
- glide 0.12.2
- nodejs 6.7.0
- npm 3.10.3
- webpack 1.13.2
- GNU Make 4.1
- go-swagger dev

Ubuntu users may use following scriptlet to install necessary tools:


```
sudo add-apt-repository ppa:masterminds/glide && sudo apt-get update
sudo apt-get install make ubuntu-make docker.io glide

umake go
umake nodejs

# current mockery version is broken, use vendored instead
#go get github.com/vektra/mockery/.../

go get -u github.com/go-swagger/go-swagger/cmd/swagger

npm install webpack -g

sudo -i
curl -L https://github.com/docker/compose/releases/download/1.8.1/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
exit
```


## Build project from sources

```
# clone repo

WRK_DIR=$GOPATH/src/github.com/letsrock-today
mkdir -p $WRK_DIR
cd $WRK_DIR
git clone https://github.com/letsrock-today/authkit.git

# update go dependencies

cd $WRK_DIR/authkit
glide install

# use vendored mockery version, till it be fixed

go install github.com/letsrock-today/authkit/vendor/github.com/vektra/mockery/.../

# update npm dependencies for every sample

pushd $WRK_DIR/sample/authkit/ui-web
npm install
popd

# you should be able to build samples with "make"
# but to run samples (using "make up") you have to adjust sample apps' configurations
# see README.md in the correspondent sample's directory

pushd $WRK_DIR/sample/authkit

make up

# App should be running at this point.
# If browser won't appear automatically, see command output in console for links.
# https://localhost:8080 - for the sample app with login dialog
# https://localhost:8080/oauth2/auth?... - for login as a third party to use app's API

make down

popd

#TODO: other samples

```

## Note

We will appreciate any interest and contribution to our project. We are not
security experts. We hope, that project will be useful enough to others to
have it well tested and reviewed and to have a benefits from its
"open-sourceness". Though, if you decide to use this (or any other
security-related open-source project) in your own application, be warned that
it's your charge to review it thoroughly, take in account all possible
implications, etc. It's **you** how are in charge of security of your
solution after all, don't blame us.
