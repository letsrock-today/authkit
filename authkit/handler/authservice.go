package handler

import (
	"context"

	"golang.org/x/oauth2"
)

// AuthService provides a low-level auth implemetation
type AuthService interface {
	GenerateConsentToken(
		subj string,
		scopes []string,
		challenge string) (string, error)
	IssueConsentToken(
		clientID string,
		scopes []string) (string, error)
	IssueToken(c context.Context, login string) (*oauth2.Token, error)
}
