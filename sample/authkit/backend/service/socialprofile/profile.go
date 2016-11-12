package socialprofile

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/letsrock-today/authkit/authkit"
)

type Profile struct {
	Login          string   `json:"-" form:"-" valid:"-"`
	Email          string   `json:"email" form:"email" valid:"required~email-required,email~email-format"`
	EmailConfirmed bool     `json:"email_confirmed" form:"-" valid:"-"`
	FormattedName  string   `json:"formatted_name" form:"formatted_name" valid:"-"`
	Location       string   `json:"location" form:"location" valid:"-"`
	Picture        string   `json:"picture" form:"picture" valid:"url"`
	Birthday       string   `json:"birthday" form:"birthday" valid:"-"`
	Gender         string   `json:"gender" form:"gender" valid:"matches(male|female|-)"`
	Phones         []string `json:"phones" form:"phones" valid:"numeric,stringlength(3|15)"`
}

func (p Profile) GetLogin() string {
	return p.Login
}

func (p *Profile) SetLogin(login string) {
	p.Login = login
}

func (p Profile) GetEmail() string {
	return p.Email
}

func (p Profile) IsEmailConfirmed() bool {
	return p.EmailConfirmed
}

func (p Profile) GetFormattedName() string {
	return p.FormattedName
}

func MakeLogin(providerID, profileID string) string {
	// It should be enough, we are not going to allow user to use this login manually.
	// If we want to allow user to sign up using social login and form-based one
	// into the same account, we'd better merge accounts.
	return fmt.Sprintf("%x", []byte(providerID+profileID))
}

type Service interface {
	authkit.SocialProfileService

	Friends(client *http.Client) ([]Profile, error)
}

func New(providerID string) (Service, error) {
	s, err := Providers().SocialProfileService(providerID)
	if err != nil {
		return nil, err
	}
	return s.(Service), nil
}

func Providers() authkit.SocialProfileServices {
	return _providers
}

type providers map[string]Service

func (p providers) SocialProfileService(providerID string) (
	authkit.SocialProfileService, error) {
	s, ok := p[providerID]
	if !ok {
		return nil, fmt.Errorf("Unknown provider: %s", providerID)
	}
	return s, nil
}

var _providers = providers{
	"fb":       facebook{},
	"linkedin": linkedin{},
	"google":   google{},
	"deezer":   deezer{},
}

func normalizeGender(g string) string {
	if strings.EqualFold(g, "male") {
		return "male"
	}
	if strings.EqualFold(g, "female") {
		return "female"
	}
	return "-"
}
