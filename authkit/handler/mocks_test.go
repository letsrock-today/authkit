package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	context2 "golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/letsrock-today/hydra-sample/authkit"
	"github.com/letsrock-today/hydra-sample/authkit/apptoken"
)

type testConfig struct {
	oauth2State           testOAuth2State
	oauth2Providers       []testOAuth2Provider
	privateOAuth2Provider testOAuth2Provider
	modTime               time.Time
	tlsInsecureSkipVerify bool
}

func (c testConfig) OAuth2Providers() chan authkit.OAuth2Provider {
	ch := make(chan authkit.OAuth2Provider)
	go func() {
		for _, p := range c.oauth2Providers {
			ch <- p
		}
		close(ch)
	}()
	return ch
}

func (c testConfig) OAuth2ProviderByID(id string) authkit.OAuth2Provider {
	for _, p := range c.oauth2Providers {
		if id == p.ID() {
			return p
		}
	}
	return nil
}

func (c testConfig) PrivateOAuth2Provider() authkit.OAuth2Provider {
	return c.privateOAuth2Provider
}

func (c testConfig) OAuth2State() authkit.OAuth2State {
	return c.oauth2State
}

func (c testConfig) AuthCookieName() string {
	return "xxx-auth-cookie"
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
	id               string
	name             string
	iconURL          string
	oauth2Config     authkit.OAuth2Config
	privOAuth2Config authkit.OAuth2Config
}

func (p testOAuth2Provider) ID() string {
	return p.id
}

func (p testOAuth2Provider) Name() string {
	return p.name
}

func (p testOAuth2Provider) IconURL() string {
	return p.iconURL
}

func (p testOAuth2Provider) OAuth2Config() authkit.OAuth2Config {
	return p.oauth2Config
}

func (p testOAuth2Provider) PrivateOAuth2Config() authkit.OAuth2Config {
	return p.privOAuth2Config
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

func (m *testAuthService) IssueToken(
	c context.Context,
	login string) (*oauth2.Token, error) {
	args := m.Called(c, login)
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

type testUserService struct {
	mock.Mock
}

func (m *testUserService) OAuth2Token(
	login, providerID string) (*oauth2.Token, authkit.UserServiceError) {
	args := m.Called(login, providerID)
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *testUserService) UpdateOAuth2Token(
	login, providerID string, token *oauth2.Token) authkit.UserServiceError {
	args := m.Called(login, providerID, token)
	return args.Error(0)
}

func (m *testUserService) Create(
	login, password string) authkit.UserServiceError {
	args := m.Called(login, password)
	err := args.Get(0)
	if err != nil {
		return err.(authkit.UserServiceError)
	}
	return nil
}

func (m *testUserService) CreateEnabled(
	login, password string) authkit.UserServiceError {
	args := m.Called(login, password)
	err := args.Get(0)
	if err != nil {
		return err.(authkit.UserServiceError)
	}
	return nil
}

func (m *testUserService) Enable(login string) authkit.UserServiceError {
	args := m.Called(login)
	return args.Error(0)
}

func (m *testUserService) Authenticate(
	login, password string) authkit.UserServiceError {
	args := m.Called(login, password)
	err := args.Get(0)
	if err != nil {
		return err.(authkit.UserServiceError)
	}
	return nil
}

func (m *testUserService) User(login string) (
	authkit.User, authkit.UserServiceError) {
	args := m.Called(login)
	err := args.Get(1)
	if err != nil {
		return nil, err.(authkit.UserServiceError)
	}
	return args.Get(0).(authkit.User), nil
}

func (m *testUserService) UpdatePassword(
	login, oldPasswordHash, newPassword string) authkit.UserServiceError {
	args := m.Called(login, oldPasswordHash, newPassword)
	err := args.Get(0)
	if err != nil {
		return err.(authkit.UserServiceError)
	}
	return nil
}

func (m *testUserService) RequestEmailConfirmation(
	login string) authkit.UserServiceError {
	args := m.Called(login)
	err := args.Get(0)
	if err != nil {
		return err.(authkit.UserServiceError)
	}
	return nil
}

func (m *testUserService) RequestPasswordChangeConfirmation(
	login, passwordHash string) authkit.UserServiceError {
	args := m.Called(login, passwordHash)
	err := args.Get(0)
	if err != nil {
		return err.(authkit.UserServiceError)
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

func (e testUserServiceError) IsDuplicateUser() bool {
	return e.isDuplicateUser
}

func (e testUserServiceError) IsUserNotFound() bool {
	return e.isUserNotFound
}

func (e testUserServiceError) IsAccountDisabled() bool {
	return e.isAccountDisabled
}

func testNewTestDuplicateUserError() authkit.UserServiceError {
	return testUserServiceError{
		isDuplicateUser: true,
	}
}

func testNewTestUserNotFoundError() authkit.UserServiceError {
	return testUserServiceError{
		isUserNotFound: true,
	}
}

func testNewTestAccountDisabledError() authkit.UserServiceError {
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

func (m *testProfileService) Save(p authkit.Profile) error {
	args := m.Called(p)
	return args.Error(0)
}

type testOAuth2Config struct {
	mock.Mock
	oauth2.Config
}

func (m *testOAuth2Config) Exchange(c context2.Context, code string) (*oauth2.Token, error) {
	args := m.Called(c, code)
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

type testSocialProfileServices struct {
	mock.Mock
}

func (m *testSocialProfileServices) SocialProfileService(
	providerID string) (authkit.SocialProfileService, error) {
	args := m.Called(providerID)
	return args.Get(0).(authkit.SocialProfileService), args.Error(1)
}

type testSocialProfileService struct {
	mock.Mock
}

func (m *testSocialProfileService) SocialProfile(
	client *http.Client) (authkit.Profile, error) {
	args := m.Called(client)
	return args.Get(0).(authkit.Profile), args.Error(1)
}

type testProfile struct {
	login string
}

func (p testProfile) Login() string {
	return p.login
}

type testBodyEncoderFunc func(v url.Values) io.Reader

var testBodyEncoders = []struct {
	name        string
	contentType string
	invalid     bool
	encoder     testBodyEncoderFunc
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

func testNewEmailTokenString(
	t *testing.T,
	config authkit.Config,
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

func testNewStateTokenString(
	t *testing.T,
	config authkit.Config,
	pid, login string,
	expired ...bool) []string {
	exp := 1 * time.Hour
	if len(expired) > 0 && expired[0] {
		exp = -1 * time.Hour
	}
	if login == "" {
		s, err := apptoken.NewStateTokenString(
			config.OAuth2State().TokenIssuer(),
			pid,
			exp,
			config.OAuth2State().TokenSignKey())
		assert.NoError(t, err)
		return []string{s}
	}
	s, err := apptoken.NewStateWithLoginTokenString(
		config.OAuth2State().TokenIssuer(),
		pid,
		login,
		exp,
		config.OAuth2State().TokenSignKey())
	assert.NoError(t, err)
	return []string{s}
}
