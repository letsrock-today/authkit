package hydra

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mendsley/gojwk"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/middleware"
)

// New creates new Hydra-backed auth service.
func New(
	hydraURL string,
	providerID string,
	providerIDTrustedContext string,
	challengeLifespan time.Duration,
	oauth2Config *oauth2.Config,
	clientCredentials *clientcredentials.Config,
	oauth2State authkit.OAuth2State,
	contextCreator authkit.ContextCreator,
	tlsInsecureSkipVerify bool) authkit.AuthService {
	if hydraURL == "" ||
		providerID == "" ||
		providerIDTrustedContext == "" ||
		challengeLifespan == 0 ||
		oauth2Config == nil ||
		clientCredentials == nil ||
		oauth2State == nil {
		panic("invalid argument")
	}
	if contextCreator == nil {
		contextCreator = authkit.DefaultContextCreator{}
	}
	return &hydra{
		hydraURL,
		providerID,
		providerIDTrustedContext,
		challengeLifespan,
		oauth2Config,
		clientCredentials,
		oauth2State,
		contextCreator,
		tlsInsecureSkipVerify,
	}
}

type hydra struct {
	hydraURL                 string
	providerID               string
	providerIDTrustedContext string // used to obtain context from contextCreator for 2-legged flow
	challengeLifespan        time.Duration
	oauth2Config             *oauth2.Config            // for 3-legged flow
	clientCredentials        *clientcredentials.Config // for 2-legged flow
	oauth2State              authkit.OAuth2State
	contextCreator           authkit.ContextCreator
	tlsInsecureSkipVerify    bool
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

func (h hydra) IssueConsentToken(
	clientID string,
	scopes []string) (string, error) {
	key, err := h.getConsentResponsePrivateKey()
	if err != nil {
		return "", errors.WithStack(err)
	}
	claims := challengeClaims{
		jwt.StandardClaims{
			Id:        uuid.New(),
			Audience:  clientID,
			ExpiresAt: time.Now().Add(h.challengeLifespan).Unix(),
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
	conf := h.oauth2Config
	signedTokenString, err := h.IssueConsentToken(conf.ClientID, conf.Scopes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(h.oauth2State.Expiration()).Unix(),
		Issuer:    h.oauth2State.TokenIssuer(),
		Audience:  login,
		Subject:   h.providerID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString(h.oauth2State.TokenSignKey())
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

	httpClient := prepareHTTPClientWithoutRedirects(h)

	resp, err := httpClient.Get(redirectURL)
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
	ctx := h.contextCreator.CreateContext(h.providerID)
	t, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return t, nil
}

func (h hydra) Validate(accessToken string, permissionDescriptor interface{}) error {
	p, ok := permissionDescriptor.(*middleware.DefaultPermission)
	if !ok {
		return errors.WithStack(errors.New("invalid permission object"))
	}
	conf := h.clientCredentials
	ctx := h.contextCreator.CreateContext(h.providerIDTrustedContext)
	client := conf.Client(ctx)

	url := fmt.Sprintf("%s/warden/token/allowed", h.hydraURL)
	b, err := json.Marshal(struct {
		Token    string   `json:"token"`
		Resource string   `json:"resource"`
		Action   string   `json:"action"`
		Scopes   []string `json:"scopes"`
	}{
		Token:    accessToken,
		Resource: p.Resource,
		Action:   p.Action,
		Scopes:   p.Scopes,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return errors.WithStack(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.WithStack(
			errors.Errorf(
				"unexpected status from Hydra: %d, %s",
				resp.StatusCode,
				resp.Status))
	}
	var r map[string]interface{}
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return errors.WithStack(err)
	}
	v, ok := r["allowed"]
	if !ok {
		return errors.WithStack(errors.New("unexpected Hydra response format"))
	}
	if allowed, ok := v.(bool); !ok || !allowed {
		return errors.WithStack(errors.New("Hydra denied access"))
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
	conf := h.clientCredentials
	ctx := h.contextCreator.CreateContext(h.providerIDTrustedContext)
	client := conf.Client(ctx)

	url := fmt.Sprintf("%s/keys/%s/%s", h.hydraURL, set, kid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("failed to retrieve key from Hydra, status: %v", resp.Status)
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	type keyResponse struct {
		Keys []gojwk.Key `json:"keys"`
	}

	var kr keyResponse
	err = json.Unmarshal(b, &kr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(kr.Keys) == 0 {
		err := fmt.Errorf("no keys from Hydra returned")
		return nil, errors.WithStack(err)
	}

	if kid == "public" {
		k, err := kr.Keys[0].DecodePublicKey()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return k, nil
	}
	if kid == "private" {
		k, err := kr.Keys[0].DecodePrivateKey()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return k, nil
	}

	return kr.Keys[0], nil
}

// replaced by test
var prepareHTTPClientWithoutRedirects = func(h hydra) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: h.tlsInsecureSkipVerify,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
