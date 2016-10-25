package authkit

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

// ContextCreator used by the Handler to prepare context for http calls.
type ContextCreator interface {

	// CreateContext returns context to be used in http requests to
	// OAuth2 providers. It allows to provide specific settings for different
	// providers. For example, application may switch off TLS for debugging.
	// Application may provide context caching, using providerID as a key.
	// Application may populate context with OAuth2 token from persistent storage.
	CreateContext(providerID string) context.Context
}

// DefaultContextCreator returns context.Background().
type DefaultContextCreator struct{}

// CreateContext returns context.Background().
func (c DefaultContextCreator) CreateContext(string) context.Context {
	return context.Background()
}

// NewCustomHTTPClientContextCreator returns new CustomHTTPClientContextCreator.
func NewCustomHTTPClientContextCreator(
	clients map[string]*http.Client) ContextCreator {
	return CustomHTTPClientContextCreator{clients}
}

// CustomHTTPClientContextCreator used to provide custom http.Client for
// some of providers. It holds a map of http.Clients for providerIDs.
// If there is no providerID in the map, then context.Background() returned for
// this providerID.
type CustomHTTPClientContextCreator struct {
	clients map[string]*http.Client
}

// CreateContext returns context with custom http.Client, if it is provided in
// the map during CustomHTTPClientContextCreator creation. It returns
// context.Background() otherwise.
func (c CustomHTTPClientContextCreator) CreateContext(
	providerID string) context.Context {
	client, ok := c.clients[providerID]
	if !ok {
		return context.Background()
	}
	return context.WithValue(
		context.Background(),
		oauth2.HTTPClient,
		client)
}
