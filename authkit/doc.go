// Package authkit is an attempt to extract oauth2 logic from application so
// that it would be possible to reuse it in different applications.
//
// Package provides interfaces for app to be implemented to supply them into
// handler.NewHandler() and middleware.AccessToken(), which return
// authkit.Handler and echo.MiddlewareFunc, respectively.
// These two objects are suitable to implement authorization logic in the
// application. Application will need to setup Echo routes and middleware with
// methods of returned objects to use them.
//
// Another option, is to reuse any helper from subpackages of package authkit
// directly in the application's code.
package authkit
