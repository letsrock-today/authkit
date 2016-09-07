package userapi

type UserAPI interface {
	Create(login, password string) error
	Authenticate(login, password string) error
}
