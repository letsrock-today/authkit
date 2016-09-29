package socialprofile

import (
	"fmt"
	"net/http"
)

//TODO: validation rules
type Profile struct {
	Email         string   `json:"email" form:"email" valid:"required, email"`
	FormattedName string   `json:"formatted_name" form:"formatted_name" valid:"-"`
	Location      string   `json:"location" form:"location" valid:"-"`
	Picture       string   `json:"picture" form:"-" valid:"-"`
	Birthday      string   `json:"birthday" form:"birthday" valid:"-"`
	Gender        string   `json:"gender" form:"gender" valid:"-"`
	Phones        []string `json:"phones" form:"phones" valid:"-"`
}

type ProfileAPI interface {

	// context should be created by oauth2 and contain token
	Profile(client *http.Client) (*Profile, error)

	Friends(client *http.Client) ([]Profile, error)
}

func New(pid string) (ProfileAPI, error) {
	api, ok := providers[pid]
	if !ok {
		return nil, fmt.Errorf("Unknown provider: %s", pid)
	}
	return api, nil
}

var providers = map[string]ProfileAPI{
	"fb":       facebook{},
	"linkedin": linkedin{},
}
