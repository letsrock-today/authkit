package handler

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConsentLogin(t *testing.T) {

	as := new(testAuthService)
	as.On(
		"GenerateConsentToken",
		"valid@login.ok",
		[]string{"valid_scope"},
		"unknown_challenge").Return("", errors.New("invalid challenge"))
	as.On(
		"GenerateConsentToken",
		"valid@login.ok",
		[]string{"valid_scope"},
		"valid_challenge").Return("valid_token", nil)
	as.On(
		"GenerateConsentToken",
		"new.valid@login.ok",
		[]string{"valid_scope"},
		"valid_challenge").Return("valid_token", nil)
	as.On(
		"GenerateConsentToken",
		"old.valid@login.ok",
		[]string{"valid_scope"},
		"valid_challenge").Return("valid_token", nil)
	as.On(
		"GenerateConsentToken",
		"broken.valid@login.ok",
		[]string{"valid_scope"},
		"valid_challenge").Return("valid_token", nil)

	us := new(testUserService)
	us.On(
		"Authenticate",
		"valid@login.ok",
		"invalid_password").Return(testUserServiceError{isUserNotFound: true})
	us.On(
		"Authenticate",
		"valid@login.ok",
		"valid_password").Return(nil)
	us.On(
		"Create",
		"new.valid@login.ok",
		"valid_password").Return(testUserServiceError{isAccountDisabled: true})
	us.On(
		"Create",
		"broken.valid@login.ok",
		"valid_password").Return(testUserServiceError{isAccountDisabled: true})
	us.On(
		"Create",
		"old.valid@login.ok",
		"valid_password").Return(testUserServiceError{isDuplicateUser: true})
	us.On(
		"RequestEmailConfirmation",
		"new.valid@login.ok").Return(nil)
	us.On(
		"RequestEmailConfirmation",
		"old.valid@login.ok").Return(nil)
	us.On(
		"RequestEmailConfirmation",
		"broken.valid@login.ok").Return(
		testUserServiceError{errors.New("cannot send email"), false, false, false})

	h := handler{
		errorCustomizer: testErrorCustomizer{},
		auth:            as,
		users:           us,
	}

	cases := []struct {
		name          string
		params        url.Values
		expStatusCode int
		expBody       string
		internalError bool
	}{
		{
			name:          "No params",
			params:        make(url.Values),
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"invalid req param"}`,
		},
		{
			name: "Login: Unknown challenge",
			params: url.Values{
				"action":    []string{"login"},
				"challenge": []string{"unknown_challenge"},
				"login":     []string{"valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "Login: invalid password",
			params: url.Values{
				"action":    []string{"login"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"valid@login.ok"},
				"password":  []string{"invalid_password"},
				"scopes":    []string{"valid_scope"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"user auth err"}`,
		},
		{
			name: "Login: OK",
			params: url.Values{
				"action":    []string{"login"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			expStatusCode: http.StatusOK,
			expBody:       `{"consent":"valid_token"}`,
		},
		{
			name: "Signup: OK",
			params: url.Values{
				"action":    []string{"signup"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"new.valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			// account disabled until user confirms it
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"acc disabled"}`,
		},
		{
			name: "Signup: duplicate",
			params: url.Values{
				"action":    []string{"signup"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"old.valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			expStatusCode: http.StatusUnauthorized,
			expBody:       `{"Code":"dup user"}`,
		},
		{
			name: "Signup: cannot send email",
			params: url.Values{
				"action":    []string{"signup"},
				"challenge": []string{"valid_challenge"},
				"login":     []string{"broken.valid@login.ok"},
				"password":  []string{"valid_password"},
				"scopes":    []string{"valid_scope"},
			},
			internalError: true,
		},
		//TODO:
	}

	for _, c := range cases {
		c := c
		for _, enc := range bodyEncoders {
			enc := enc
			t.Run(c.name+", "+enc.name, func(st *testing.T) {
				st.Parallel()
				assert := assert.New(st)
				e := echo.New()

				req, err := http.NewRequest(echo.POST, "", enc.encoder(c.params))
				assert.NoError(err)
				rec := httptest.NewRecorder()
				ctx := e.NewContext(
					standard.NewRequest(req, e.Logger()),
					standard.NewResponse(rec, e.Logger()))
				ctx.Request().Header().Set("Content-Type", enc.contentType)

				err = h.ConsentLogin(ctx)
				if enc.invalid || c.internalError {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(c.expStatusCode, rec.Code)
					assert.Equal(c.expBody, string(rec.Body.Bytes()))
				}
			})
		}
	}

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

type testUserServiceError struct {
	error
	isDuplicateUser   bool
	isUserNotFound    bool
	isAccountDisabled bool
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

type testErrorCustomizer struct{}

func (testErrorCustomizer) InvalidRequestParameterError(error) interface{} {
	return struct {
		Code string
	}{
		"invalid req param",
	}
}

func (testErrorCustomizer) UserCreationError(e error) interface{} {
	msg := "user creation err"
	if err, ok := e.(UserServiceError); ok {
		if err.IsAccountDisabled() {
			msg = "acc disabled"
		} else if err.IsDuplicateUser() {
			msg = "dup user"
		}
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

type testUserService struct {
	mock.Mock
}

func (m *testUserService) Create(login, password string) UserServiceError {
	args := m.Called(login, password)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(UserServiceError)
}

func (m *testUserService) Authenticate(login, password string) UserServiceError {
	args := m.Called(login, password)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(UserServiceError)
}

func (m *testUserService) RequestEmailConfirmation(login string) UserServiceError {
	args := m.Called(login)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(UserServiceError)
}
