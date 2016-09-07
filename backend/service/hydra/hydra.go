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

func VerifyConsentChallenge(c string) (*jwt.Token, error) {
	key, err := getKey("consent.challenge", "public")
	if err != nil {
		return nil, err
	}
	return jwt.Parse(c, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})
}

/*
generateConsentToken(subject, scopes, challenge) {
        warn()
        return new Promise((resolve, reject) => {
            this.getKey('consent.endpoint', 'private').then((key) => {
                const {payload: {aud, exp}}  = jwt.decode(challenge, {complete: true})
                jwt.sign({
                    aud,
                    exp,
                    scp: scopes,
                    sub: subject
                }, jwkToPem({
                    ...key,
                    // the following keys are optional in the spec but for some reason required by the library.
                    dp: '', dq: '', qi: ''
                }, {private: true}), {algorithm: 'RS256'}, (error, token) => {
                    if (error) {
                        return reject({error: 'Could not verify consent challenge: ' + error})
                    }
                    resolve({consent: token})
                })
            }).catch(reject)
        })
    }
*/
