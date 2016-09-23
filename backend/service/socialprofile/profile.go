package socialprofile

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Profile struct {
	Id            string
	Email         string
	FormattedName string
	Location      string
	Picture       string
	Birthday      string
	Gender        string
	Phones        []string
}

type ProfileAPI interface {

	// close storage
	io.Closer

	// context should be created by oauth2 and contain token
	Profile(client *http.Client) (*Profile, error)

	Save(client *http.Client, profile *Profile) error

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
	"fb":           facebook{},
	"linkedin":     linkedin{},
	"hydra-sample": hydrasample{},
}

var ErrorNotImplemented = errors.New("Method not implemented")
