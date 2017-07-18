package mp

import (
	"../../"
	"github.com/vmihailenco/msgpack"
)

type mc struct{}

func (*mc) Marshal(data interface{}) ([]byte, error) {
	return msgpack.Marshal(data)
}

func (*mc) Unmarshal(payload []byte, data interface{}) error {
	return msgpack.Unmarshal(payload, data)
}

func (*mc) Texture() bool {
	return false
}

func New() stmp.Codec {
	return &mc{}
}
