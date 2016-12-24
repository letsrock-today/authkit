package hydra

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/dgrijalva/jwt-go"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/pkg/errors"

	"github.com/letsrock-today/authkit/authkit"
	hydraclient "github.com/letsrock-today/authkit/authkit/hydra/client"
	"github.com/letsrock-today/authkit/authkit/hydra/client/jwk"
	hydraoauth2 "github.com/letsrock-today/authkit/authkit/hydra/client/oauth2"
	"github.com/letsrock-today/authkit/authkit/hydra/client/warden"
	"github.com/letsrock-today/authkit/authkit/hydra/models"
	"github.com/letsrock-today/authkit/authkit/middleware"
)

//go:generate swagger generate client -f hydra.authkit.yaml

// New creates new Hydra-backed auth service.
func New(c Config) authkit.AuthService {
	if !c.Valid() {
		panic("invalid argument")
	}
	if c.ContextCreator == nil {
		c.ContextCreator = authkit.DefaultContextCreator{}
	}
	u, err := url.Parse(c.HydraURL)
	if err != nil {
		panic(err)
	}
	return &hydra{c, u}
}

// Config represents configuration for hydra.New().
type Config struct {
	HydraURL                 string
	ProviderID               string
	ProviderIDTrustedContext string // used to obtain context from contextCreator for 2-legged flow
	ChallengeLifespan        time.Duration
	OAuth2Config             *oauth2.Config            // for 3-legged flow
	ClientCredentials        *clientcredentials.Config // for 2-legged flow
	OAuth2State              authkit.OAuth2State
	ContextCreator           authkit.ContextCreator
	TLSInsecureSkipVerify    bool
}

// Valid validates configuration.
func (c Config) Valid() bool {
	return c.HydraURL != "" &&
		c.ProviderID != "" &&
		c.ProviderIDTrustedContext != "" &&
		c.ChallengeLifespan != 0 &&
		c.OAuth2Config != nil &&
		c.ClientCredentials != nil
}

type hydra struct {
	Config
	HydraParsedURL *url.URL
}

func (h hydra) GenerateConsentToken(
	subj string,
	scopes []string,
	challenge string) (string, error) {
	decodedChallenge, err := h.verifyConsentChallenge(challenge)
	if err != nil {
		return "", errors.WithStack(err)
	}
	key, err := h.getConsentResponsePrivateKey()
	if err != nil {
		return "", errors.WithStack(err)
	}
	decodedChallengeClaims := decodedChallenge.Claims.(*challengeClaims)
	for _, n := range scopes {
		exist := false
		for _, o := range decodedChallengeClaims.Scopes {
			if n == o {
				exist = true
				break
			}
		}
		if !exist {
			return "", errors.WithStack(errors.New("disallowed to enlarge set of scopes"))
		}
	}
	claims := challengeClaims{
		jwt.StandardClaims{
			Audience:  decodedChallengeClaims.Audience,
			ExpiresAt: decodedChallengeClaims.ExpiresAt,
			Subject:   subj,
		},
		scopes,
		"",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	s, err := token.SignedString(key)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return s, nil
}

func (h hydra) GenerateConsentTokenPriv(
	subj string,
	scopes []string,
	clientID string) (string, error) {
	key, err := h.getConsentResponsePrivateKey()
	if err != nil {
		return "", errors.WithStack(err)
	}
	claims := challengeClaims{
		jwt.StandardClaims{
			Audience:  clientID,
			ExpiresAt: time.Now().Add(h.ChallengeLifespan).Unix(),
			Subject:   subj,
		},
		scopes,
		"",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	s, err := token.SignedString(key)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return s, nil
}

func (h hydra) IssueToken(login string) (*oauth2.Token, error) {
	// This method used to retrieve access token for our own web app in
	// case of form-based login within app (not consent page).
	// Not sure, whether it's safe to expose token,
	// obtained through 2-legged flow, to the external world. So, we just
	// emulate 3-legged flow on behave of user. We already authorized him
	// and want to give him similar token, as for third-party user,
	// so that tokens in both cases could be validated with same middleware.
	conf := h.OAuth2Config
	signedTokenString, err := h.GenerateConsentTokenPriv(
		login,
		conf.Scopes,
		conf.ClientID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(h.OAuth2State.Expiration).Unix(),
		Issuer:    h.OAuth2State.TokenIssuer,
		Audience:  login,
		Subject:   h.ProviderID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString(h.OAuth2State.TokenSignKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	u, err := url.Parse(conf.Endpoint.AuthURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	v := u.Query()
	v.Set("client_id", conf.ClientID)
	v.Set("response_type", "code")
	v.Set("scope", strings.Join(conf.Scopes, " "))
	v.Set("state", state)
	v.Set("consent", signedTokenString)
	u.RawQuery = v.Encode()
	redirectURL := u.String()

	client := prepareHTTPClientWithoutRedirects(h)

	resp, err := client.Get(redirectURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode != http.StatusFound {
		return nil, errors.WithStack(errors.Errorf(
			"Unexpected response status: %d, %s",
			resp.StatusCode,
			resp.Status))
	}
	location := resp.Header.Get("Location")
	u, err = url.Parse(location)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	v = u.Query()
	code := v.Get("code")
	// Here we use 3-legged flow, hence correspondent context.
	ctx := h.ContextCreator.CreateContext(h.ProviderID)
	t, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return t, nil
}

func (h hydra) Validate(
	accessToken string,
	permissionDescriptor interface{}) (string, error) {
	p, ok := permissionDescriptor.(*middleware.DefaultPermission)
	if !ok {
		return "", errors.WithStack(errors.New("invalid permission object"))
	}
	conf := h.ClientCredentials
	ctx := h.ContextCreator.CreateContext(h.ProviderIDTrustedContext)

	client := hydraclient.New(
		httptransport.NewWithClient(
			h.HydraParsedURL.Host,
			h.HydraParsedURL.Path,
			[]string{h.HydraParsedURL.Scheme},
			conf.Client(ctx),
		),
		strfmt.Default)

	r, err := client.Warden.IsAllowed(
		warden.NewIsAllowedParams().WithBody(&models.WardenIsAllowedRequest{
			Token:    &accessToken,
			Resource: &p.Resource,
			Action:   &p.Action,
			Scopes:   p.Scopes,
		}))
	if err != nil {
		return "", errors.WithStack(err)
	}

	if !*r.Payload.Allowed {
		return "", errors.WithStack(errors.New("Hydra denied access"))
	}
	return *r.Payload.Sub, nil
}

func (h hydra) RevokeAccessToken(accessToken string) error {
	conf := h.ClientCredentials
	ctx := h.ContextCreator.CreateContext(h.ProviderIDTrustedContext)
	httpclient := http.DefaultClient
	if hc, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		httpclient.Transport = hc.Transport
	}
	client := hydraclient.New(
		httptransport.NewWithClient(
			h.HydraParsedURL.Host,
			h.HydraParsedURL.Path,
			[]string{h.HydraParsedURL.Scheme},
			httpclient,
		),
		strfmt.Default)
	_, err := client.Oauth2.Revoke(
		hydraoauth2.NewRevokeParams().WithToken(accessToken),
		httptransport.BasicAuth(conf.ClientID, conf.ClientSecret),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type challengeClaims struct {
	jwt.StandardClaims
	Scopes      []string `json:"scp,omitempty"`
	RedirectURL string   `json:"redir,omitempty"`
}

func (h hydra) verifyConsentChallenge(c string) (*jwt.Token, error) {
	key, err := h.getConsentChallengePublicKey()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	token, err := jwt.ParseWithClaims(
		c,
		&challengeClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return key, nil
		})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return token, nil
}

var (
	consentChallengePublicKey   interface{}
	consentChallengePublicKeyTS time.Time
	consentResponsePrivateKey   interface{}
	consentResponsePrivateKeyTS time.Time
	ttl                         = 10 * time.Minute
)

func (h hydra) getConsentChallengePublicKey() (interface{}, error) {
	if consentChallengePublicKey == nil ||
		time.Since(consentChallengePublicKeyTS) > ttl {
		var err error
		consentChallengePublicKey, err = h.getKey("hydra.consent.challenge", "public")
		if err != nil {
			consentChallengePublicKey = nil
			return nil, err
		}
		consentChallengePublicKeyTS = time.Now()
	}
	return consentChallengePublicKey, nil
}

func (h hydra) getConsentResponsePrivateKey() (interface{}, error) {
	if consentResponsePrivateKey == nil ||
		time.Since(consentResponsePrivateKeyTS) > ttl {
		var err error
		consentResponsePrivateKey, err = h.getKey("hydra.consent.response", "private")
		if err != nil {
			consentResponsePrivateKey = nil
			return nil, err
		}
		consentResponsePrivateKeyTS = time.Now()
	}
	return consentResponsePrivateKey, nil
}

func (h hydra) getKey(set, kid string) (interface{}, error) {
	conf := h.ClientCredentials
	ctx := h.ContextCreator.CreateContext(h.ProviderIDTrustedContext)

	client := hydraclient.New(
		httptransport.NewWithClient(
			h.HydraParsedURL.Host,
			h.HydraParsedURL.Path,
			[]string{h.HydraParsedURL.Scheme},
			conf.Client(ctx),
		),
		strfmt.Default)

	r, err := client.Jwk.GetJWK(
		jwk.NewGetJWKParams().WithSet(set).WithKid(kid))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	keys := r.Payload.Keys
	if len(keys) == 0 {
		err := fmt.Errorf("no keys from Hydra returned")
		return nil, errors.WithStack(err)
	}
	key := keys[0]
	var k interface{}
	switch kid {
	default:
		return key, nil
	case "public":
		k, err = key.DecodePublicKey()
	case "private":
		k, err = key.DecodePrivateKey()
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return k, nil
}

// replaced by test
var prepareHTTPClientWithoutRedirects = func(h hydra) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: h.TLSInsecureSkipVerify,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
