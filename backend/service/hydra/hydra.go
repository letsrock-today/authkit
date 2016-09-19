package hydra

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mendsley/gojwk"
	"github.com/pborman/uuid"
	"golang.org/x/oauth2"

	"github.com/letsrock-today/hydra-sample/backend/config"
)

// context to use for internal requests to Hydra
//TODO: use real certeficates in PROD and remove transport replacement
var ctx = context.WithValue(
	context.Background(),
	oauth2.HTTPClient,
	&http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}})

type ChallengeClaims struct {
	jwt.StandardClaims
	Scopes      []string `json:"scp,omitempty"`
	RedirectURL string   `json:"redir,omitempty"`
}

func VerifyConsentChallenge(c string) (*jwt.Token, error) {
	key, err := getKey("consent.challenge", "public")
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
	key, err := getKey("consent.endpoint", "private")
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
	key, err := getKey("consent.endpoint", "private")
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

func getKey(set, kid string) (interface{}, error) {
	c := config.Get()
	conf := c.HydraClientCredentials
	client := conf.Client(ctx)

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
