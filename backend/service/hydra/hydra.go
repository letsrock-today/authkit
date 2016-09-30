package hydra

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mendsley/gojwk"
	"github.com/pborman/uuid"
	"golang.org/x/oauth2"

	"github.com/letsrock-today/hydra-sample/backend/config"
)

// context to use for internal requests to Hydra
var (
	once sync.Once
	ctx  context.Context
)

type ChallengeClaims struct {
	jwt.StandardClaims
	Scopes      []string `json:"scp,omitempty"`
	RedirectURL string   `json:"redir,omitempty"`
}

func VerifyConsentChallenge(c string) (*jwt.Token, error) {
	key, err := getKey("hydra.consent.challenge", "public")
	if err != nil {
		return nil, err
	}
	return jwt.ParseWithClaims(
		c,
		&ChallengeClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return key, nil
		})
}

func GenerateConsentToken(
	subj string,
	scopes []string,
	challenge string) (string, error) {
	decodedChallenge, err := VerifyConsentChallenge(challenge)
	if err != nil {
		return "", err
	}
	key, err := getKey("hydra.consent.response", "private")
	if err != nil {
		return "", err
	}
	decodedChallengeClaims := decodedChallenge.Claims.(*ChallengeClaims)
	for _, n := range scopes {
		exist := false
		for _, o := range decodedChallengeClaims.Scopes {
			if n == o {
				exist = true
				break
			}
		}
		if !exist {
			return "", fmt.Errorf("Disallowed to enlarge set of scopes")
		}
	}
	claims := ChallengeClaims{
		jwt.StandardClaims{
			Audience:  decodedChallengeClaims.Audience,
			ExpiresAt: decodedChallengeClaims.ExpiresAt,
			Subject:   subj,
		},
		scopes,
		"",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

func IssueConsentToken(
	client_id string,
	scopes []string) (string, error) {
	key, err := getKey("hydra.consent.response", "private")
	if err != nil {
		return "", err
	}
	claims := ChallengeClaims{
		jwt.StandardClaims{
			Id:        uuid.New(),
			Audience:  client_id,
			ExpiresAt: time.Now().Add(config.Get().ChallengeLifespan).Unix(),
		},
		scopes,
		"",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

func IssueToken(ctx context.Context, login string) (*oauth2.Token, error) {
	cfg := config.Get()
	conf := cfg.HydraOAuth2ConfigInt
	signedTokenString, err := IssueConsentToken(
		conf.ClientID,
		conf.Scopes)
	if err != nil {
		return nil, err
	}

	claims :=
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(cfg.OAuth2State.Expiration).Unix(),
			Issuer:    cfg.OAuth2State.TokenIssuer,
			Audience:  login,
			Subject:   config.PrivPID,
		}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString(cfg.OAuth2State.TokenSignKey)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(conf.Endpoint.AuthURL)
	if err != nil {
		return nil, err
	}
	v := u.Query()
	v.Set("client_id", conf.ClientID)
	v.Set("response_type", "code")
	v.Set("scope", strings.Join(conf.Scopes, " "))
	v.Set("state", state)
	v.Set("consent", signedTokenString)
	u.RawQuery = v.Encode()
	redirectURL := u.String()
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.TLSInsecureSkipVerify,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := httpClient.Get(redirectURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusFound {
		return nil, fmt.Errorf("Unexpected response status: %d, %s", resp.StatusCode, resp.Status)
	}
	location := resp.Header.Get("Location")
	u, err = url.Parse(location)
	if err != nil {
		return nil, err
	}
	v = u.Query()
	code := v.Get("code")
	return conf.Exchange(ctx, code)
}

func ValidateAccessTokenPermissions(
	token, resource, action string,
	scopes []string) error {
	c := config.Get()
	conf := c.HydraClientCredentials
	client := conf.Client(getHttpContext())

	url := fmt.Sprintf("%s/warden/token/allowed", c.HydraAddr)
	b, err := json.Marshal(struct {
		Token    string   `json:"token"`
		Resource string   `json:"resource"`
		Action   string   `json:"action"`
		Scopes   []string `json:"scopes"`
	}{
		Token:    token,
		Resource: resource,
		Action:   action,
		Scopes:   scopes,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("Hydra returned unexpected status:", resp.StatusCode, resp.Status)
		return errors.New("unexpected status from Hydra")
	}
	var r map[string]interface{}
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &r)
	if err != nil {
		return err
	}
	v, ok := r["allowed"]
	if !ok {
		return errors.New("unexpected Hydra response format")
	}
	if allowed, ok := v.(bool); !ok || !allowed {
		return errors.New("access denied")
	}
	return nil
}

func getKey(set, kid string) (interface{}, error) {
	c := config.Get()
	conf := c.HydraClientCredentials
	client := conf.Client(getHttpContext())

	url := fmt.Sprintf("%s/keys/%s/%s", c.HydraAddr, set, kid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve key from Hydra, status: %v", resp.Status)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type keyResponse struct {
		Keys []gojwk.Key `json:"keys"`
	}

	var kr keyResponse
	err = json.Unmarshal(b, &kr)
	if err != nil {
		return nil, err
	}

	if len(kr.Keys) == 0 {
		return nil, fmt.Errorf("no keys from Hydra returned")
	}

	if kid == "public" {
		return kr.Keys[0].DecodePublicKey()
	}
	if kid == "private" {
		return kr.Keys[0].DecodePrivateKey()
	}

	return kr.Keys[0], nil
}

func getHttpContext() context.Context {
	// context to use for internal requests to Hydra
	once.Do(func() {
		ctx = context.WithValue(
			context.Background(),
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: config.Get().TLSInsecureSkipVerify,
					},
				}})
	})
	return ctx
}
