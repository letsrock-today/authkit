package authkit

import "context"

//TODO: rename interface and method (ContextProvider.Context(...) - ?)
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
