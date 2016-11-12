package apptoken

import "github.com/pkg/errors"

// ErrInvalidToken error returned when token cannot be parsed.
var ErrInvalidToken = errors.New("invalid token")
