package socialprofile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/letsrock-today/authkit/authkit"
)

type googleConnectionsResponse struct {
	Connections []googleProfile `json:"connections"`
	Error       *googleError    `json:"error"`
}

type googleProfileResponse struct {
	googleProfile
	Error *googleError `json:"error"`
}

type googleError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type googleProfile struct {
	Emails       []googleEmail       `json:"emailAddresses"`
	Names        []googleName        `json:"names"`
	Addresses    []googleAddress     `json:"addresses"`
	Photos       []googlePhoto       `json:"photos"`
	Birthdays    []googleBirthday    `json:"birthdays"`
	Genders      []googleGender      `json:"genders"`
	PhoneNumbers []googlePhoneNumber `json:"phoneNumbers"`
}

type googleMetadata struct {
	Primary bool `json:"primary"`
}

type googleEmail struct {
	Metadata googleMetadata `json:"metadata"`
	Value    string         `json:"value"`
}

type googleName struct {
	Metadata    googleMetadata `json:"metadata"`
	DisplayName string         `json:"displayName"`
}

type googleAddress struct {
	Metadata       googleMetadata `json:"metadata"`
	FormattedValue string         `json:"formattedValue"`
}

type googlePhoto struct {
	Metadata googleMetadata `json:"metadata"`
	URL      string         `json:"url"`
}

type googleBirthday struct {
	Metadata googleMetadata `json:"metadata"`
	Text     string         `json:"text"`
}

type googleGender struct {
	Metadata googleMetadata `json:"metadata"`
	Value    string         `json:"value"`
}

type googlePhoneNumber struct {
	Metadata      googleMetadata `json:"metadata"`
	CanonicalForm string         `json:"canonicalForm"`
}

type google struct{}

const (
	googleProfileQueryURL     = "https://people.googleapis.com/v1/people/me?fields=addresses(formattedValue%2Cmetadata%2Fprimary)%2Cbirthdays%2Ftext%2CemailAddresses(metadata%2Fprimary%2Cvalue)%2Cgenders%2FformattedValue%2Cnames(displayName%2Cmetadata%2Fprimary)%2CphoneNumbers(canonicalForm%2Cvalue)%2Cphotos(metadata%2Fprimary%2Curl)"
	googleConnectionsQueryURL = "https://people.googleapis.com/v1/people/me/connections?fields=connections(names(displayName%2Cmetadata%2Fprimary)%2Cphotos(metadata%2Fprimary%2Curl)%2Crelations)"
)

func (google) SocialProfile(client *http.Client) (authkit.Profile, error) {
	resp, err := client.Get(googleProfileQueryURL)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var p googleProfileResponse
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if p.Error != nil {
		return nil, fmt.Errorf(p.Error.Message)
	}
	return convertProfile(p.googleProfile), nil
}

func (google) Friends(client *http.Client) ([]Profile, error) {
	resp, err := client.Get(googleConnectionsQueryURL)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cr googleConnectionsResponse
	if err := json.Unmarshal(b, &cr); err != nil {
		return nil, err
	}
	if cr.Error != nil {
		return nil, fmt.Errorf(cr.Error.Message)
	}

	var res []Profile
	for _, p := range cr.Connections {
		res = append(res, convertProfile(p))
	}
	return res, nil
}

func convertProfile(p googleProfile) Profile {
	email := ""
	if len(p.Emails) > 0 {
		email = p.Emails[0].Value
	}
	for _, e := range p.Emails {
		if e.Metadata.Primary {
			email = e.Value
		}
	}
	name := ""
	if len(p.Names) > 0 {
		name = p.Names[0].DisplayName
	}
	for _, n := range p.Names {
		if n.Metadata.Primary {
			name = n.DisplayName
		}
	}
	address := ""
	if len(p.Addresses) > 0 {
		address = p.Addresses[0].FormattedValue
	}
	for _, a := range p.Addresses {
		if a.Metadata.Primary {
			address = a.FormattedValue
		}
	}
	picture := ""
	if len(p.Photos) > 0 {
		picture = p.Photos[0].URL
	}
	for _, p := range p.Photos {
		if p.Metadata.Primary {
			picture = p.URL
		}
	}
	birthday := ""
	if len(p.Birthdays) > 0 {
		birthday = p.Birthdays[0].Text
	}
	for _, b := range p.Birthdays {
		if b.Metadata.Primary {
			birthday = b.Text
		}
	}
	gender := ""
	if len(p.Genders) > 0 {
		gender = p.Genders[0].Value
	}
	for _, g := range p.Genders {
		if g.Metadata.Primary {
			gender = g.Value
		}
	}

	return Profile{
		Email:         email,
		FormattedName: name,
		Location:      address,
		Picture:       picture,
		Birthday:      birthday,
		Gender:        normalizeGender(gender),
		Phones:        []string{},
	}
}
