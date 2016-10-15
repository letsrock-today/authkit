package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
	"github.com/letsrock-today/hydra-sample/authkit/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

type testConfig struct {
	oauth2State           testOAuth2State
	oauth2Configs         map[string]oauth2.Config
	oauth2Providers       []testOAuth2Provider
	privateOAuth2Provider testOAuth2Provider
	privateOAuth2Config   oauth2.Config
	modTime               time.Time
}

func (c testConfig) OAuth2State() config.OAuth2State {
	return c.oauth2State
}

func (c testConfig) OAuth2Configs() map[string]oauth2.Config {
	return c.oauth2Configs
}

func (c testConfig) OAuth2Providers() []config.OAuth2Provider {
	r := []config.OAuth2Provider{}
	for _, p := range c.oauth2Providers {
		r = append(r, p)
	}
	return r
}

func (c testConfig) PrivateOAuth2Provider() config.OAuth2Provider {
	return c.privateOAuth2Provider
}

func (c testConfig) PrivateOAuth2Config() oauth2.Config {
	return c.privateOAuth2Config
}

func (c testConfig) ModTime() time.Time {
	return c.modTime
}

type testOAuth2State struct {
	tokenIssuer  string
	tokenSignKey []byte
	expiration   time.Duration
}

func (s testOAuth2State) TokenIssuer() string {
	return s.tokenIssuer
}

func (s testOAuth2State) TokenSignKey() []byte {
	return s.tokenSignKey
}

func (s testOAuth2State) Expiration() time.Duration {
	return s.expiration
}

type testOAuth2Provider struct {
	id           string
	name         string
	clientID     string
	clientSecret string
	publicKey    string
	scopes       []string
	iconURL      string
	tokenURL     string
	authURL      string
}

func (p testOAuth2Provider) ID() string {
	return p.id
}

func (p testOAuth2Provider) Name() string {
	return p.name
}

func (p testOAuth2Provider) ClientID() string {
	return p.clientID
}

func (p testOAuth2Provider) ClientSecret() string {
	return p.clientSecret
}

func (p testOAuth2Provider) PublicKey() string {
	return p.publicKey
}

func (p testOAuth2Provider) Scopes() []string {
	return p.scopes
}

func (p testOAuth2Provider) IconURL() string {
	return p.iconURL
}

func (p testOAuth2Provider) TokenURL() string {
	return p.tokenURL
}

func (p testOAuth2Provider) AuthURL() string {
	return p.authURL
}

type bodyEncoderFunc func(v url.Values) io.Reader

var bodyEncoders = []struct {
	name        string
	contentType string
	invalid     bool
	encoder     bodyEncoderFunc
}{
	{
		name:        "invalid payload",
		contentType: "application/json",
		invalid:     true,
		encoder: func(v url.Values) io.Reader {
			return strings.NewReader("zzz")
		},
	},
	{
		name:        "form payload",
		contentType: "application/x-www-form-urlencoded",
		encoder: func(v url.Values) io.Reader {
			return strings.NewReader(v.Encode())
		},
	},
	{
		name:        "multipart payload",
		contentType: "multipart/form-data; boundary=---",
		encoder: func(v url.Values) io.Reader {
			var b bytes.Buffer
			w := multipart.NewWriter(&b)
			w.SetBoundary("---")
			for n, v := range v {
				for _, v := range v {
					fw, err := w.CreateFormField(n)
					if err != nil {
						panic(err)
					}
					_, err = fw.Write([]byte(v))
					if err != nil {
						panic(err)
					}
				}
			}
			w.Close()
			return &b
		},
	},
}

type testErrorCustomizer struct{}

func (testErrorCustomizer) InvalidRequestParameterError(error) interface{} {
	return struct {
		Code string
	}{
		"invalid req param",
	}
}

func (testErrorCustomizer) UserCreationError(e error) interface{} {
	var msg string
	err := e.(testUserServiceError)
	switch {
	case err.IsDuplicateUser():
		msg = "dup user"
	case err.IsAccountDisabled():
		msg = "acc disabled"
	default:
		msg = "general error"
	}
	return struct {
		Code string
	}{
		msg,
	}
}

func (testErrorCustomizer) UserAuthenticationError(error) interface{} {
	return struct {
		Code string
	}{
		"user auth err",
	}
}

type testAuthService struct {
	mock.Mock
}

func (m *testAuthService) GenerateConsentToken(
	subj string,
	scopes []string,
	challenge string) (string, error) {
	args := m.Called(subj, scopes, challenge)
	return args.String(0), args.Error(1)
}

func (m *testAuthService) IssueConsentToken(
	clientID string,
	scopes []string) (string, error) {
	args := m.Called(clientID, scopes)
	return args.String(0), args.Error(1)
}

type testUserService struct {
	mock.Mock
}

func (m *testUserService) Create(login, password string) UserServiceError {
	args := m.Called(login, password)
	err := args.Get(0)
	if err != nil {
		return err.(UserServiceError)
	}
	return nil
}

func (m *testUserService) Authenticate(login, password string) UserServiceError {
	args := m.Called(login, password)
	err := args.Get(0)
	if err != nil {
		return err.(UserServiceError)
	}
	return nil
}

func (m *testUserService) User(login string) (User, UserServiceError) {
	args := m.Called(login)
	err := args.Get(1)
	if err != nil {
		return nil, err.(UserServiceError)
	}
	return args.Get(0).(User), nil
}

func (m *testUserService) UpdatePassword(login, oldPasswordHash, newPassword string) UserServiceError {
	args := m.Called(login, oldPasswordHash, newPassword)
	err := args.Get(0)
	if err != nil {
		return err.(UserServiceError)
	}
	return nil
}
func (m *testUserService) Enable(login string) UserServiceError {
	args := m.Called(login)
	return args.Error(0)
}

func (m *testUserService) RequestEmailConfirmation(login string) UserServiceError {
	args := m.Called(login)
	err := args.Get(0)
	if err != nil {
		return err.(UserServiceError)
	}
	return nil
}

func (m *testUserService) RequestPasswordChangeConfirmation(login, passwordHash string) UserServiceError {
	args := m.Called(login, passwordHash)
	err := args.Get(0)
	if err != nil {
		return err.(UserServiceError)
	}
	return nil
}

type testUserServiceError struct {
	isUserNotFound    bool
	isDuplicateUser   bool
	isAccountDisabled bool
}

func (testUserServiceError) Error() string {
	return "user service error"
}

func (e testUserServiceError) IsUserNotFound() bool {
	return e.isUserNotFound
}

func (e testUserServiceError) IsDuplicateUser() bool {
	return e.isDuplicateUser
}

func (e testUserServiceError) IsAccountDisabled() bool {
	return e.isAccountDisabled
}

func newTestUserNotFoundError() UserServiceError {
	return testUserServiceError{
		isUserNotFound: true,
	}
}

func newTestDuplicateUserError() UserServiceError {
	return testUserServiceError{
		isDuplicateUser: true,
	}
}

func newTestAccountDisabledError() UserServiceError {
	return testUserServiceError{
		isAccountDisabled: true,
	}
}

type testUser struct {
	login        string
	email        string
	passwordHash string
}

func (u testUser) Login() string {
	return u.login
}

func (u testUser) Email() string {
	return u.email
}

func (u testUser) PasswordHash() string {
	return u.passwordHash
}

type testProfileService struct {
	mock.Mock
}

func (m *testProfileService) EnsureExists(login string) error {
	args := m.Called(login)
	return args.Error(0)
}

func newEmailTokenString(
	t *testing.T,
	config config.Config,
	email, passwordHash string,
	expired ...bool) []string {
	exp := 1 * time.Hour
	if len(expired) > 0 && expired[0] {
		exp = -1 * time.Hour
	}
	s, err := apptoken.NewEmailTokenString(
		config.OAuth2State().TokenIssuer(),
		email,
		passwordHash,
		exp,
		config.OAuth2State().TokenSignKey())
	assert.NoError(t, err)
	return []string{s}
}
