package querybinder

import (
	"reflect"

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
		Metadata:   nil,
		Result:     i,
		TagName:    "form",
		DecodeHook: flattenSlice,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(c.Request().FormParams())
}

func flattenSlice(f, t reflect.Kind, data interface{}) (interface{}, error) {
	if t != reflect.String || f != reflect.Slice {
		return data, nil
	}
	r := data.([]string)
	if len(r) != 1 {
		return data, nil
	}
	return r[0], nil
}
