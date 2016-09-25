package socialprofile

import (
	"fmt"
	"net/http"
)

type Profile struct {
	Email         string   `json:"email"`
	FormattedName string   `json:"formatted_name"`
	Location      string   `json:"location"`
	Picture       string   `json:"picture"`
	Birthday      string   `json:"birthday"`
	Gender        string   `json:"gender"`
	Phones        []string `json:"phones"`
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
