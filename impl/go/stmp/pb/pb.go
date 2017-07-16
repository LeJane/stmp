package pb

import (
	"../"
	"github.com/golang/protobuf/proto"
)

type pc struct{}

func (*pc) Marshal(msg *stmp.Message) ([]byte, error) {
	if msg.Data == nil {
		return nil, nil
	}
	if m, ok := msg.Data.(proto.Message); ok {
		return proto.Marshal(m)
	}
	return nil, stmp.ErrIncorrectInputType
}

func (*pc) Unmarshal(msg *stmp.Message, output interface{}) error {
	if msg.Payload == nil || len(msg.Payload) == 0 {
		return nil
	}
	if m, ok := output.(proto.Message); ok {
		return proto.Unmarshal(msg.Payload, m)
	}
	return stmp.ErrIncorrectOutputType
}

func New() stmp.Codec {
	return &pc{}
}
