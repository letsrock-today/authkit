package hydra

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/dgrijalva/jwt-go"

	"github.com/letsrock-today/hydra-sample/backend/config"
)

func getKey(set, kid string) (string, error) {
	c := config.GetConfig()
	conf := c.HydraOAuth2Config
	//TODO: use real certeficates in PROD and remove this
	ctx := context.WithValue(
		context.Background(),
		oauth2.HTTPClient,
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}})
	client := conf.Client(ctx)

	url := fmt.Sprintf("%s/keys/%s/%s", c.HydraAddr, set, kid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	//TODO

	return string(b), nil
}

func VerifyConsentChallenge(c string) (*jwt.Token, error) {
	key, err := getKey("consent.challenge", "public")
	if err != nil {
		return nil, err
	}
	log.Println("Key from hydra", key)
	return jwt.Parse(c, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})
}
