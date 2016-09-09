package hydra

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/mendsley/gojwk"
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
	decodedChallengeClaims := decodedChallenge.Claims.(*ChallengeClaims)
	claims := ChallengeClaims{
		jwt.StandardClaims{
			Audience:  decodedChallengeClaims.Audience,
			ExpiresAt: decodedChallengeClaims.ExpiresAt,
			Subject:   subj,
		},
		scopes,
		"",
	}
	key, err := getKey("consent.endpoint", "private")
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

func getKey(set, kid string) (interface{}, error) {
	c := config.GetConfig()
	conf := c.HydraOAuth2Config
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
