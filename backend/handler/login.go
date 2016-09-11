package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/mitchellh/mapstructure"

	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/service/hydra"
	"github.com/letsrock-today/hydra-sample/backend/util/jwtutil"
)

type (
	basicLoginForm struct {
		Login    []string `mapstructure:"login" valid:"email,required"`
		Password []string `mapstructure:"password" valid:"stringlength(3|10),required"`
	}
	loginForm struct {
		Challenge []string `mapstructure:"challenge" valid:"required"`
		Login     []string `mapstructure:"login" valid:"email,required"`
		Password  []string `mapstructure:"password" valid:"stringlength(3|10),required"`
		Scopes    []string `mapstructure:"scopes" valid:"stringlength(1|500),required"`
	}
	loginReply struct {
		Consent string `json:"consent"`
	}
)

func Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		writeErrorResponse(w, err)
		return
	}

	// To simplify validation logic we convert map to structure first

	var lf loginForm
	if err := mapstructure.Decode(r.Form, &lf); err != nil {
		writeErrorResponse(w, err)
		return
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		writeErrorResponse(w, err)
		return
	}

	if err := UserService.Authenticate(
		lf.Login[0],
		lf.Password[0]); err != nil {
		writeErrorResponse(w, err)
		return
	}

	signedTokenString, err := hydra.GenerateConsentToken(
		lf.Login[0],
		lf.Scopes,
		lf.Challenge[0])
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	reply := loginReply{
		Consent: signedTokenString,
	}
	b, err := json.Marshal(reply)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func LoginHydra(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		writeErrorResponse(w, err)
		return
	}

	// To simplify validation logic we convert map to structure first

	var lf basicLoginForm
	if err := mapstructure.Decode(r.Form, &lf); err != nil {
		writeErrorResponse(w, err)
		return
	}

	if _, err := govalidator.ValidateStruct(lf); err != nil {
		writeErrorResponse(w, err)
		return
	}

	cfg := config.GetConfig()
	signedTokenString, err := hydra.IssueConsentToken(
		cfg.HydraOAuth2Config.ClientID,
		cfg.HydraOAuth2Config.Scopes)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	state, err := jwtutil.NewJWTSignedString(
		cfg.OAuth2State.TokenSignKey,
		cfg.OAuth2State.TokenIssuer,
		"hydra", // TODO
		cfg.OAuth2State.Expiration)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		writeErrorResponse(w, err)
		return
	}

	u, err := url.Parse(cfg.HydraOAuth2Config.Endpoint.AuthURL)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	v := u.Query()
	v.Set("client_id", cfg.HydraOAuth2Config.ClientID)
	v.Set("response_type", "code")
	v.Set("scope", strings.Join(cfg.HydraOAuth2Config.Scopes, "+"))
	v.Set("state", state)
	v.Set("nonce", base64.URLEncoding.EncodeToString(nonce))
	v.Set("consent", signedTokenString)
	u.RawQuery = v.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}
