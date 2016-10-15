package handler

// ProfileService provides methods to persiste user profiles.
type ProfileService interface {
	EnsureExists(login string) error
}
