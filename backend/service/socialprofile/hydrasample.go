package socialprofile

import "net/http"

type hydrasample struct{}

func (hydrasample) Profile(client *http.Client) (*Profile, error) {
	//TODO: use DB
	return nil, ErrorNotImplemented
}

func (hydrasample) Save(client *http.Client, profile *Profile) error {
	//TODO: use DB
	return ErrorNotImplemented
}

func (hydrasample) Friends(client *http.Client) ([]Profile, error) {
	return nil, ErrorNotImplemented
}

func (hydrasample) Close() error {
	//hs.dbsession.Close()
	return nil
}
