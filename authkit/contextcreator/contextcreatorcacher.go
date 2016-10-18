package contextcreator

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"

	"github.com/letsrock-today/hydra-sample/authkit"
	"golang.org/x/oauth2"
)

// ContextCreatorCacher implements ContextCreatoir and
// provides context caching, using providerID as a key.
type ContextCreatorCacher struct {
	mu       sync.Mutex
	contexts map[string]context.Context
	//us            authkit.MiddlewareUserService
	tlsSkipVerify bool
}

// NewContextCreator returns ContextCreator.
func NewContextCreator( /*us authkit.MiddlewareUserService,*/ tlsSkipVerify bool) authkit.ContextCreator {
	return &ContextCreatorCacher{
		//us:            us,
		tlsSkipVerify: tlsSkipVerify,
	}
}

// CreateContext returns context from cache by providerID or
// creates new one in case of his absence.
func (c *ContextCreatorCacher) CreateContext(providerID string) context.Context {
	// TODO
	if ctx, ok := c.contexts[providerID]; !ok {
		c.mu.Lock()
		ctx = context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: c.tlsSkipVerify,
					},
				}})
		c.contexts[providerID] = ctx
		c.mu.Unlock()
		return ctx
	} else {
		return ctx
	}
}
