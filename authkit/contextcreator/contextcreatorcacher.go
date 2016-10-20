package contextcreator

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/letsrock-today/hydra-sample/authkit"
	"golang.org/x/oauth2"
)

const (
	OAuth2TokenKeyName = "oauth2token" // TODO key name
)

// ContextCreatorCacher implements ContextCreatoir and
// provides context caching, using providerID as a key.
type ContextCreatorCacher struct {
	config *Config
	us     authkit.MiddlewareUserService

	mu       sync.Mutex
	contexts *lru.Cache

	pmu             sync.Mutex
	privateContexts *lru.Cache
}

// NewContextCreator returns ContextCreator.
func NewContextCreator(us authkit.MiddlewareUserService) (authkit.ContextCreator, error) {
	return NewContextCreatorWithConfig(us, NewConfig(10, 42))
}

func NewContextCreatorWithConfig(us authkit.MiddlewareUserService, config *Config) (authkit.ContextCreator, error) {
	contexts, err := lru.New(config.cacheContextSize)
	if err != nil {
		return nil, err
	}
	privateContexts, err := lru.New(config.cachePrivateContextSize)
	if err != nil {
		return nil, err
	}
	return &ContextCreatorCacher{
		us:              us,
		config:          config,
		contexts:        contexts,
		privateContexts: privateContexts,
	}, nil
}

// CreateContext returns context from cache by providerID or
// creates new one in case of it absence.
func (c *ContextCreatorCacher) CreateContext(providerID string) (context.Context, error) {
	// TODO add server token into context
	if ctx, ok := c.contexts.Get(providerID); !ok {
		c.mu.Lock()
		ctx := context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: c.config.Get(providerID).TLSSkipVerify,
					},
				}})
		c.mu.Unlock()
		c.contexts.Add(providerID, ctx)
		return ctx, nil
	} else {
		return ctx.(context.Context), nil
	}
}

// CreatePrivateContext returns context from cache by login, providerID or
// creates new one in case of it absence.
func (c *ContextCreatorCacher) CreatePrivateContext(login, providerID string) (context.Context, error) {
	// TODO get token by login from store via c.us
	// TODO add token into context
	if ctx, ok := c.privateContexts.Get(providerID); !ok {
		c.pmu.Lock()
		ctx := context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: c.config.Get(providerID).TLSSkipVerify,
					},
				}})
		token, err := c.us.OAuth2Token(login, providerID)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, OAuth2TokenKeyName, token)
		c.pmu.Unlock()
		c.privateContexts.Add(providerID, ctx)
		return ctx, nil
	} else {
		return ctx.(context.Context), nil
	}
}
