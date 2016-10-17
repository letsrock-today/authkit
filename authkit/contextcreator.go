package authkit

import "context"

// ContextCreator used by the Handler to prepare context for http calls.
type ContextCreator interface {

	// CreateContext returns context to be used in http requests to
	// OAuth2 providers. When external is true, application should return
	// context suitable to access external OAuth2 providers, otherwise it
	// should return context suitable to access private (on-promise) OAuth2
	// provider. For example, for private context app may switch off TLS
	// for debugging.
	CreateContext(external bool) context.Context
}

// DefaultContextCreator returns context.Background().
type DefaultContextCreator struct{}

// CreateContext returns context.Background().
func (c DefaultContextCreator) CreateContext(external bool) context.Context {
	return context.Background()
}
