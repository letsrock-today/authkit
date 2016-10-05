# authkit

"authkit" is a set of http handlers and middleware to implement auth2
authorization and SSO in the web application (from the resource owner point of view).

This project is not aimed to create yet another OAuth2 provider or client
library, it is rather aimed to glue existing implementations, to fill gaps
between them and to represent demo (blueprint?) solution(s) for particular
authorization scenario(s). Also, any piece of code (helper, handler, middleware)
should be customizable and reusable.

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
  our own application;

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
- ...

Our aim is to provide:

- [ ] a set of server-side primitives (http handlers, routes, middleware, helpers)
      usable to implement different oauth2 scenarios;
- [ ] binary auth module for Nginx (see http://nginx.org/en/docs/http/ngx_http_auth_request_module.html);
- [ ] examples of our library usage in both modes: directly and behind Nginx;
- [ ] sample web app implementation using Riot.js;
- [ ] sample Android app implementation;

Finally, when the lib is ready, we want to migrate our own application
(https://letsrock.today/) to it and to enjoy benefits of going open-source.

Currently almost all of relevant parts are implemented in our closed-source
application or in samples, but code is a bit messy and scattered, so we are not
ready to open it as it is.

At the time of writing this text we are using Go 1.7.
We use Glide to manage dependencies.

See ./sample/authkit folder for demo app, README.md in that folder contains instructions how to run it.
