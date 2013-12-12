package db

import (
	"github.com/mitchellh/mapstructure"
)

func (h Hstore) ToStruct(i interface{}) error {
	return MapToStruct(h, i)
}

func MapToStruct(h map[string]interface{}, i interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		Result:           i,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(h)
}
