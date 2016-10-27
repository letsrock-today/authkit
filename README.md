[![Build Status](https://travis-ci.org/letsrock-today/hydra-sample.svg?branch=master)](https://travis-ci.org/letsrock-today/hydra-sample)

# authkit

"authkit" is a set of http handlers and middleware to implement auth2
authorization and SSO in the web application (from the resource owner point of view).

This project is not aimed to create yet another OAuth2 provider or client
library, it is rather aimed to glue existing implementations, to fill gaps
between them and to represent demo (blueprint?) solution(s) for particular
authorization scenario(s). Also, any piece of code (helper, handler, middleware)
should be customizable and reusable.

If you simply need a social login via one provider from the list, you may use
one of this libraries:
- https://github.com/golang/oauth2;
- https://github.com/markbates/goth.

Our project uses golang/oauth2 and simplifies configuration a bit
(in our opinion). Though, this task is may be already straightforward enough.

markbates/goth is good for the task of using many different auth providers from
predefined list, it provides ability to configure providers and to implement
new ones. It even goes as far as to retrieve user-related data from provider
(like user name, location or avatar URL from social network). But it leaves
couple of gaps unfilled:
- it seems that it has no ready-to-use integration with form-based login;
- it seems that it has no ready-to-use token-based auth middleware;
- it seems that it has no ready-to-use integration with on-promise
  oauth2 provider (so that app would make it's API accessible to 3rd party
  with login via OAuth2);

This gaps may be filled with other projects. Echo, Gorilla, Negrony, Iris, etc
can be used to provide necessary handlers and middleware. And our project is an
attempt to provide some reusable pieces to fill this gap.

We started this project from porting and tailoring
[Hydra usage example](https://github.com/ory-am/hydra-idp-react). It may be seen
as more elaborated usage example of Hydra, oauth2, Echo and related stuff.

Initially, we are focused on the following task:

- we have custom http API;
- we want this API or part of it be available only to authorized users;
- we want that users be able to authorize using their existing social network
  accounts (Facebook, Google+, LinkedIn, etc);
- we want to be able to provide username/password login as well;
- we want that no matter which type of login user choose, API would be protected
  using single approach with access token, issued by our side;
- we want to be able to make our API or parts of it available to third parties
  via OAuth2 code flow;
- we want to expose single IP address to our clients (or several equivalent
  addresses), so, for example, we want to proxy requests to Hydra via Nginx or
  our own application.

So, our side in this context controls resource owner, OAuth2 client and provider.

This project uses:

- https://github.com/labstack/echo, as a server framework (to implement routes,
  http handlers, middleware, etc); "echo" designed to be compatible with
  gorilla, negroni, standard lib, so, it should be possible to use our lib
  with them either;
- https://github.com/golang/oauth2, as a client OAuth2 library;
- https://github.com/ory-am/hydra, as a OAuth2 provider;
- https://github.com/dgrijalva/jwt-go, to create tokens, when we have to;
- https://github.com/asaskevich/govalidator, for validation logic;
- https://github.com/mitchellh/mapstructure, to convert maps into structures;
- https://github.com/hashicorp/golang-lru, for caching;
- https://github.com/stretchr/testify, for testing and mocking;
- https://github.com/h2non/gock, to mock HTTP server;
- ...

Our aim is to provide:

- [x] a set of server-side primitives (http handlers, routes, middleware, helpers)
      usable to implement different oauth2 scenarios;
- [ ] binary auth module for Nginx (see http://nginx.org/en/docs/http/ngx_http_auth_request_module.html);
- [ ] examples of our library usage in both modes: directly and behind Nginx;
- [x] sample web app implementation using [Riot.js](http://riotjs.com/);
- [ ] sample Android app implementation;

TODO:
- how simple it would be to reuse providers' implementations from markbates/goth?
- if we use swagger or another lib to generate SDK from provider's API, or
  ready-to-use SDK from provider in our app, is it still convenient to use
  something like markbates/goth (we assume, that application creator still
  needs provider's API for other features)?
- should we try to implement example with Iris to learn is it simple/possible?
- if we use swagger or another lib to generate server-side stubs from app's API
  description, how it would affect usage of the lib? could we provide custom
  templates to generate code, which is used our lib and Echo?

Finally, when the lib is ready, we want to migrate our own application
(https://letsrock.today/) to it and to enjoy benefits of going open-source.

Currently almost all of relevant parts are implemented in our closed-source
application or in samples, but code is a bit messy and scattered, so we are not
ready to open it as it is.

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

Ubuntu users may use following scriptlet to install necessary tools:


```
sudo add-apt-repository ppa:masterminds/glide && sudo apt-get update
sudo apt-get install make ubuntu-make docker.io glide

umake go
umake nodejs

go get github.com/vektra/mockery/.../

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
git clone https://github.com/letsrock-today/hydra-sample.git

# update go dependencies

cd $WRK_DIR/hydra-sample
glide up

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
