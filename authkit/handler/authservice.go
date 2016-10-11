package handler

// AuthService provides a low-level auth implemetation
type AuthService interface {
	GenerateConsentToken(
		subj string,
		scopes []string,
		challenge string) (string, error)
}
