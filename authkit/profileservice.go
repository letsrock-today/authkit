package authkit

import "net/http"

type (

	// Profile is an interface representing user's profile (social and local).
	// Actually, it's assumed that application maps social profiles to local ones
	// in the SocialProfileService.
	Profile interface {
		Login() string
	}

	// ProfileService provides methods to persist user profiles (locally).
	ProfileService interface {
		EnsureExists(login string) error
		Save(Profile) error
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
