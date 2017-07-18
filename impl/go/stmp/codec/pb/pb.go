package pb

import (
	"../../"
	"github.com/golang/protobuf/proto"
)

type pc struct{}

func (*pc) Marshal(data interface{}) ([]byte, error) {
	if m, ok := data.(proto.Message); ok {
		return proto.Marshal(m)
	}
	return nil, stmp.ErrIncorrectDataType
}

func (*pc) Unmarshal(payload []byte, data interface{}) error {
	if m, ok := data.(proto.Message); ok {
		return proto.Unmarshal(payload, m)
	}
	return stmp.ErrIncorrectDataType
}

func (*pc) Texture() bool {
	return false
}

func New() stmp.Codec {
	return &pc{}
}
