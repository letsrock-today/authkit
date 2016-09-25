package socialprofile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type fbProfileResponse struct {
	fbProfile
	Error *fbErrorResponse `json:"error"`
}

type fbProfile struct {
	Id       string     `json:"id"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	Picture  fbPicture  `json:"picture"`
	Birthday string     `json:"birthday"`
	Gender   string     `json:"gender"`
	Location fbLocation `json:"location"`
}

type fbPicture struct {
	fbPictureData `json:"data"`
}

type fbPictureData struct {
	Url string `json:"url"`
}

type fbErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type fbLocation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type fbFriendsResponse struct {
	Profiles []fbProfile      `json:"data"`
	Error    *fbErrorResponse `json:"error"`
}

type facebook struct{}

const (
	fbProfileQueryURL = "https://graph.facebook.com/me?fields=id,email,name,picture,birthday,gender,location"
	fbFriendsQueryURL = "https://graph.facebook.com/me/friends?fields=id,email,name,picture,birthday,gender,location"
)

func (facebook) Profile(client *http.Client) (*Profile, error) {
	resp, err := client.Get(fbProfileQueryURL)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var p fbProfileResponse
	if err = json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if p.Error != nil {
		return nil, fmt.Errorf(p.Error.Message)
	}
	return &Profile{
		Email:         p.Email,
		FormattedName: p.Name,
		Location:      p.Location.Name,
		Picture:       p.Picture.Url,
		Birthday:      p.Birthday,
		Gender:        p.Gender,
		Phones:        []string{},
	}, nil
}

func (facebook) Friends(client *http.Client) ([]Profile, error) {
	resp, err := client.Get(fbFriendsQueryURL)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var friends fbFriendsResponse
	err = json.Unmarshal(b, &friends)
	if err != nil {
		return nil, err
	}
	if friends.Error != nil {
		return nil, fmt.Errorf(friends.Error.Message)
	}

	res := make([]Profile, 0)
	for _, p := range friends.Profiles {
		res = append(res, Profile{
			Email:         p.Email,
			FormattedName: p.Name,
			Location:      p.Location.Name,
			Picture:       p.Picture.Url,
			Birthday:      p.Birthday,
			Gender:        p.Gender,
			Phones:        []string{},
		})
	}
	return res, nil
}
