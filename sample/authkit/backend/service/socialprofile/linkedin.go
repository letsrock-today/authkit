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

type lnPhoneNumber struct {
	Type   string `json:"phoneType"`
	Number string `json:"phoneNumber"`
}

type lnPhoneNumbers struct {
	Values []lnPhoneNumber `json:"values"`
}

type lnProfile struct {
	ID            string         `json:"id"`
	Email         string         `json:"emailAddress"`
	FormattedName string         `json:"formattedName"`
	MainAddress   string         `json:"mainAddress"`
	Picture       string         `json:"pictureUrl"`
	PhoneNumbers  lnPhoneNumbers `json:"phoneNumbers"`
}

type linkedin struct{}

const (
	lnProfileURL            = "https://api.linkedin.com%s?oauth2_access_token=%s&format=json"
	lnProfileQueryURLOpaque = "/v1/people/~:(id,formatted-name,main-address,email-address,phone-numbers,picture-url)"
)

func (linkedin) SocialProfile(client *http.Client) (authkit.Profile, error) {
	//TODO: Similar approach in Deezer. Every request has additional query
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
	url := fmt.Sprintf(lnProfileURL, lnProfileQueryURLOpaque, token.AccessToken)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Opaque = lnProfileQueryURLOpaque
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-li-format", "json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var p lnProfile
	err = json.Unmarshal(b, &p)
	if err != nil {
		return nil, err
	}

	phones := []string{}

	for _, v := range p.PhoneNumbers.Values {
		phones = append(phones, v.Number)
	}

	return &Profile{
		Login:         MakeLogin("linkedin", p.ID),
		Email:         p.Email,
		FormattedName: p.FormattedName,
		Location:      p.MainAddress,
		Picture:       p.Picture,
		Birthday:      "", // Requires partner account
		Gender:        "-",
		Phones:        phones,
	}, nil
}

func (linkedin) Friends(client *http.Client) ([]Profile, error) {
	return nil, errors.New("Friends API request is not supported due Linked in policy (partner account required)")
}
