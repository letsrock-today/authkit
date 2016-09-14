package querybinder

import (
	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

type binder struct {
}

var b binder

func New() echo.Binder {
	return &b
}

func (b *binder) Bind(i interface{}, c echo.Context) (err error) {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   i,
		TagName:  "form",
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(c.Request().FormParams())
}
