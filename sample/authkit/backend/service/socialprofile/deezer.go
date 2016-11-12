package socialprofile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/letsrock-today/authkit/authkit"
)

type deezerProfileResponse struct {
	deezerProfile
	Error *deezerErrorResponse `json:"error"`
}

type deezerProfile struct {
	ID      int64  `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type deezerErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type deezer struct{}

const (
	deezerProfileQueryURL = "https://api.deezer.com/user/me?access_token=%s"
)

func (deezer) SocialProfile(client *http.Client) (authkit.Profile, error) {
	//TODO: Similar approach in LinkedIn. Every request has additional query
	// param with access token. May be to introduce custom http.Client for
	// every such provider and move this code there?
	transport, ok := client.Transport.(*oauth2.Transport)
	if !ok {
		return nil, errors.New("Cannot retrieve token from http.Client")
	}
	token, err := transport.Source.Token()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(deezerProfileQueryURL, token.AccessToken)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var p deezerProfileResponse
	if err = json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if p.Error != nil {
		return nil, fmt.Errorf(p.Error.Message)
	}
	return &Profile{
		Login:         MakeLogin("deezer", fmt.Sprintf("%x", p.ID)),
		Email:         p.Email,
		FormattedName: p.Name,
		Picture:       p.Picture,
	}, nil
}

func (deezer) Friends(client *http.Client) ([]Profile, error) {
	return nil, errors.New("Friends API request is not implemented for Deezer")
}
