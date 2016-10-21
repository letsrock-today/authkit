package socialprofile

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/letsrock-today/hydra-sample/authkit"
)

type Profile struct {
	Email         string   `json:"email" form:"email" valid:"required,email"`
	FormattedName string   `json:"formatted_name" form:"formatted_name" valid:"-"`
	Location      string   `json:"location" form:"location" valid:"-"`
	Picture       string   `json:"picture" form:"picture" valid:"url"`
	Birthday      string   `json:"birthday" form:"birthday" valid:"-"`
	Gender        string   `json:"gender" form:"gender" valid:"matches(male|female|-)"`
	Phones        []string `json:"phones" form:"phones" valid:"numeric,stringlength(333|15)"`
}

func (p Profile) Login() string {
	return p.Email
}

type Service interface {
	authkit.SocialProfileService

	Friends(client *http.Client) ([]Profile, error)
}

func New(providerID string) (Service, error) {
	s, ok := providers[providerID]
	if !ok {
		return nil, fmt.Errorf("Unknown provider: %s", providerID)
	}
	return s, nil
}

var providers = map[string]Service{
	"fb":       facebook{},
	"linkedin": linkedin{},
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
