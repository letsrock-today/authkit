package confirmer

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/letsrock-today/authkit/authkit"
	"github.com/letsrock-today/authkit/authkit/apptoken"
	"github.com/letsrock-today/authkit/sample/authkit/backend/config"
)

type confirmer struct {
	confirmEmailURL    string
	confirmPasswordURL string
}

func New(confirmEmailURL, confirmPasswordURL string) authkit.Confirmer {
	return confirmer{
		confirmEmailURL,
		confirmPasswordURL,
	}
}

func (c confirmer) RequestEmailConfirmation(
	login, email, name string) authkit.UserServiceError {
	err := sendConfirmationEmail(login, email, name, "", c.confirmEmailURL, false)
	if err != nil {
		return authkit.NewRequestConfirmationError(err)
	}
	return nil
}

func (c confirmer) RequestPasswordChangeConfirmation(
	login, email, name, passwordHash string) authkit.UserServiceError {
	err := sendConfirmationEmail(login, email, name, passwordHash, c.confirmPasswordURL, true)
	if err != nil {
		return authkit.NewRequestConfirmationError(err)
	}
	return nil
}

var (
	resetPasswordTmpl = template.Must(template.New("resetPasswordTmpl").Parse(`
Dear {{ .Name }},

We received request for password recovery for service [authkit-sample],
account [{{ .Login }}].

Follow this link to change your password: {{ .URL }}.
`))

	confirmEmailTmpl = template.Must(template.New("confirmEmailTmpl").Parse(`
Dear {{ .Name }},

This email address was provided for password recovery and other communications
regarding account for service [authkit-sample], account [{{ .Login }}].

Follow this link to confirm that you allowed use of this email address: {{ .URL }}.
`))
)

func sendConfirmationEmail(
	login, email, name, passwordhash, urlpath string,
	resetPassword bool) error {
	if login == "" || email == "" {
		return errors.New("Invalid argument")
	}
	if name == "" {
		name = "user"
	}
	c := config.Get()
	oauth2State := c.OAuth2State
	token, err := apptoken.NewEmailTokenString(
		oauth2State.TokenIssuer,
		login,
		email,
		passwordhash,
		c.ConfirmationLinkLifespan,
		oauth2State.TokenSignKey)
	if err != nil {
		return err
	}

	externalURL := c.ExternalBaseURL + urlpath
	link := fmt.Sprintf("%s?token=%s", externalURL, token)
	var (
		tmpl  *template.Template
		topic string
	)
	if resetPassword {
		tmpl = resetPasswordTmpl
		topic = "Confirm password reset"
	} else {
		tmpl = confirmEmailTmpl
		topic = "Confirm account creation"
	}
	text := &bytes.Buffer{}
	if err := tmpl.Execute(text, struct {
		Name  string
		Login string
		URL   string
	}{
		Name:  name,
		Login: login,
		URL:   link,
	}); err != nil {
		return err
	}
	return Send(email, topic, text.String())
}
