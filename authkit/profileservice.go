package authkit

import "net/http"

type (

	// Profile is an interface representing user's profile (social and local).
	// Actually, it's assumed that application maps social profiles to local ones
	// in the SocialProfileService.
	Profile interface {
		GetLogin() string
		SetLogin(string)
		GetEmail() string
		IsEmailConfirmed() bool
		GetFormattedName() string
	}

	// ProfileService provides methods to persist user profiles (locally).
	ProfileService interface {

		// EnsureExists creates new empty profile if it is not exists already.
		EnsureExists(login, email string) error

		// Save saves profile.
		Save(Profile) error

		// SetEmailConfirmed sets email confirmed flag.
		SetEmailConfirmed(login, email string, confirmed bool) error

		// Email returns (confirmed or not) email address (and user name) by login.
		Email(login string) (email, name string, err error)

		// ConfirmedEmail returns confirmed email address (and user name) by login.
		ConfirmedEmail(login string) (email, name string, err error)
	}

	// SocialProfileServices allows to discover SocialProfileService by provider ID.
	SocialProfileServices interface {
		SocialProfileService(providerID string) (SocialProfileService, error)
	}

	// SocialProfileService allows to retrieve social profile from the social
	// network or other OAuth2 provider and map it to the local user's profile.
	SocialProfileService interface {
		SocialProfile(client *http.Client) (Profile, error)
	}
)

//go:generate mockery -name Profile
//go:generate mockery -name ProfileService
//go:generate mockery -name SocialProfileService
//go:generate mockery -name SocialProfileServices
