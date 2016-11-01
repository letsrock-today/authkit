package deezer

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/letsrock-today/authkit/authkit"
)

// Endpoint is Deezer's OAuth 2.0 endpoint.
var Endpoint = oauth2.Endpoint{
	AuthURL:  "https://connect.deezer.com/oauth/auth.php",
	TokenURL: "https://connect.deezer.com/oauth/access_token.php",
}

var _ authkit.OAuth2Config = &Config{}

// Config provides peculiar OAuth2 implementation for Deezer.
type Config struct {
	*oauth2.Config
}

// AuthCodeURL do the same as it's counterpart in golang.org/x/oauth2, but
// with different set of query params.
func (c *Config) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	var buf bytes.Buffer
	buf.WriteString(c.Endpoint.AuthURL)
	v := url.Values{
		"response_type": {"code"},
		"app_id":        {c.ClientID},
		"redirect_uri":  condVal(c.RedirectURL),
		"perms":         condVal(strings.Join(c.Scopes, ",")),
		"state":         condVal(state),
	}
	if strings.Contains(c.Endpoint.AuthURL, "?") {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}
	buf.WriteString(v.Encode())
	return buf.String()
}

// Exchange do the same as it's counterpart in golang.org/x/oauth2, but
// use completely different token exchange request.
func (c *Config) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	v := url.Values{
		"app_id": {c.ClientID},
		"code":   {code},
		"secret": {c.ClientSecret},
	}
	hc := oauth2.NewClient(ctx, nil)
	r, err := hc.Get(fmt.Sprintf("%s?%s", c.Endpoint.TokenURL, v.Encode()))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	if code := r.StatusCode; code < 200 || code > 299 {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v\nResponse: %s", r.Status, body)
	}

	var token *oauth2.Token
	content, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	switch content {
	case "application/x-www-form-urlencoded", "text/plain", "text/html":
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		token = &oauth2.Token{
			AccessToken:  vals.Get("access_token"),
			TokenType:    vals.Get("token_type"),
			RefreshToken: vals.Get("refresh_token"),
		}
		expires, _ := strconv.Atoi(vals.Get("expires"))
		if expires != 0 {
			token.Expiry = time.Now().Add(time.Duration(expires) * time.Second)
		}
	default:
		return nil, fmt.Errorf("Unknown Content-Type: %s", content)
	}
	// Don't overwrite `RefreshToken` with an empty value
	// if this was a token refreshing request.
	if token.RefreshToken == "" {
		token.RefreshToken = v.Get("refresh_token")
	}
	return token, nil
}

func condVal(v string) []string {
	if v == "" {
		return nil
	}
	return []string{v}
}
