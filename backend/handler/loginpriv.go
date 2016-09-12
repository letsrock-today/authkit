package handler

import (
	"encoding/json"
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
	privLoginForm struct {
		Login    []string `mapstructure:"login" valid:"email,required"`
		Password []string `mapstructure:"password" valid:"stringlength(3|10),required"`
	}
	privLoginReply struct {
		RedirectURL string `json:"redirUrl"`
	}
)

// Login for "priveleged" client - app's own UI
func LoginPriv(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		writeErrorResponse(w, err)
		return
	}

	// TODO: protect against csrf

	// To simplify validation logic we convert map to structure first

	var lf privLoginForm
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
		"hydra-sample",
		cfg.OAuth2State.Expiration)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	/*
		nonce := make([]byte, 12)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			writeErrorResponse(w, err)
			return
		}
	*/

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
	//v.Set("nonce", base64.URLEncoding.EncodeToString(nonce))
	v.Set("consent", signedTokenString)
	u.RawQuery = v.Encode()

	reply := privLoginReply{
		RedirectURL: u.String(),
	}
	b, err := json.Marshal(reply)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
