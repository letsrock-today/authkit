package jwk

import (
	"github.com/go-openapi/strfmt"
	"github.com/mendsley/gojwk"
)

type Key struct {
	gojwk.Key
}

func (*Key) Validate(formats strfmt.Registry) error {
	return nil
}
