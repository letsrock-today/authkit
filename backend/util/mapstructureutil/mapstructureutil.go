package mapstructureutil

import (
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func JoinStringsFunc(sep string) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if t != reflect.String || f != reflect.Slice {
			return data, nil
		}

		raw := data.([]string)

		return strings.Join(raw, sep), nil
	}
}

func DecodeWithHook(
	hook mapstructure.DecodeHookFunc,
	input, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:   nil,
		Result:     output,
		DecodeHook: hook,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}
