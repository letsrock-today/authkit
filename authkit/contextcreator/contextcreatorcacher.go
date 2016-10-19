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
	config   Config
	mu       sync.Mutex
	contexts map[string]context.Context
	us       authkit.MiddlewareUserService

	pmu             sync.Mutex
	privateContexts map[string]context.Context
	tlsSkipVerify   bool
}

// NewContextCreator returns ContextCreator.
func NewContextCreator(us authkit.MiddlewareUserService) authkit.ContextCreator {
	return &ContextCreatorCacher{
		us: us,
	}
}

func NewContextCreatorWithConfig(us authkit.MiddlewareUserService, config Config) authkit.ContextCreator {
	return &ContextCreatorCacher{
		us:     us,
		config: config,
	}
}

// CreateContext returns context from cache by providerID or
// creates new one in case of it absence.
func (c *ContextCreatorCacher) CreateContext(providerID string) context.Context {
	// TODO add token into context
	if ctx, ok := c.contexts[providerID]; !ok {
		c.mu.Lock()
		ctx = context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: c.config.Get(providerID).TLSSkipVerify,
					},
				}})
		c.contexts[providerID] = ctx
		c.mu.Unlock()
		return ctx
	} else {
		return ctx
	}
}

// CreatePrivateContext returns context from cache by login, providerID or
// creates new one in case of it absence.
func (c *ContextCreatorCacher) CreatePrivateContext(login, providerID string) context.Context {
	// TODO get token by login from store via c.us
	// TODO add token into context
	if ctx, ok := c.privateContexts[providerID]; !ok {
		c.pmu.Lock()
		ctx = context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: c.config.Get(providerID).TLSSkipVerify,
					},
				}})
		c.privateContexts[providerID] = ctx
		c.pmu.Unlock()
		return ctx
	} else {
		return ctx
	}
}
