package contextcreator

import (
	"context"
	"crypto/tls"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"github.com/letsrock-today/hydra-sample/authkit"
	"golang.org/x/oauth2"
)

// ContextCreatorCacher implements ContextCreatoir and
// provides context caching, using providerID as a key.
type ContextCreatorCacher struct {
	config   *Config
	us       authkit.MiddlewareUserService
	contexts *lru.Cache
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
	return &ContextCreatorCacher{
		us:       us,
		config:   config,
		contexts: contexts,
	}, nil
}

// CreateContext returns context from cache by providerID or
// creates new one in case of it absence.
func (c *ContextCreatorCacher) CreateContext(providerID string) context.Context {
	if ctx, ok := c.contexts.Get(providerID); !ok {
		ctx := context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: c.config.Get(providerID).TLSSkipVerify,
					},
				}})
		c.contexts.Add(providerID, ctx)
		return ctx
	} else {
		return ctx.(context.Context)
	}
}
