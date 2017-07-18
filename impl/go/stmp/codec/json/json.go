package json

import (
	"encoding/json"
	"../../"
)

type jc struct{}

func (*jc) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (*jc) Unmarshal(payload []byte, data interface{}) error {
	return json.Unmarshal(payload, data)
}

func (*jc) Texture() bool {
	return true
}

func New() stmp.Codec {
	return &jc{}
}
