package contextcacher

import (
	"context"
	"crypto/tls"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"github.com/letsrock-today/hydra-sample/authkit"
	"golang.org/x/oauth2"
)

// creatorCacher implements ContextCreatoir and
// provides context caching, using providerID as a key.
type contextCreator struct {
	config   *Config
	us       authkit.MiddlewareUserService
	contexts *lru.Cache
}

// NewContextCreator returns contextCreator.
func New(us authkit.MiddlewareUserService) (authkit.ContextCreator, error) {
	return NewWithConfig(us, NewConfig(100))
}

func NewWithConfig(
	us authkit.MiddlewareUserService,
	config *Config) (authkit.ContextCreator, error) {
	contexts, err := lru.New(config.cacheContextSize)
	if err != nil {
		return nil, err
	}
	return &contextCreator{
		us:       us,
		config:   config,
		contexts: contexts,
	}, nil
}

// CreateContext returns context from cache by providerID or
// creates new one in case of it absence.
func (c *contextCreator) CreateContext(providerID string) context.Context {
	if ctx, ok := c.contexts.Get(providerID); ok {
		return ctx.(context.Context)
	}
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
}
