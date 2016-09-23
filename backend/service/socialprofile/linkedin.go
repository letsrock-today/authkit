package socialprofile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
)

type lnPhoneNumber struct {
	Type   string `json:"phoneType"`
	Number string `json:"phoneNumber"`
}

type lnPhoneNumbers struct {
	Values []lnPhoneNumber `json:"values"`
}

type lnProfile struct {
	Email         string         `json:"emailAddress"`
	Id            string         `json:"id"`
	FormattedName string         `json:"formattedName"`
	MainAddress   string         `json:"mainAddress"`
	PhoneNumbers  lnPhoneNumbers `json:"phoneNumbers"`
}

type linkedin struct{}

const (
	lnProfileURL            = "https://api.linkedin.com%s?oauth2_access_token=%s&format=json"
	lnProfileQueryURLOpaque = "/v1/people/~:(id,formatted-name,main-address,email-address,phone-numbers)"
)

func (linkedin) Profile(client *http.Client) (*Profile, error) {
	transport, ok := client.Transport.(*oauth2.Transport)
	if !ok {
		return nil, errors.New("Cannot retrieve token from http.Client")
	}
	token, err := transport.Source.Token()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(
		lnProfileURL,
		lnProfileQueryURLOpaque,
		token.AccessToken)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Opaque = lnProfileQueryURLOpaque
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-li-format", "json")
	resp, err := client.Do(req)
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
		Id:            p.Id,
		Email:         p.Email,
		FormattedName: p.FormattedName,
		Location:      p.MainAddress,
		Picture:       "", //TODO
		Birthday:      "",
		Gender:        "",
		Phones:        phones,
	}, nil
}

func (linkedin) Save(client *http.Client, profile *Profile) error {
	return ErrorNotImplemented
}

func (linkedin) Friends(client *http.Client) ([]Profile, error) {
	return nil, ErrorNotImplemented
}

func (linkedin) Close() error {
	return nil
}
