package mp

import (
	"../"
	"github.com/vmihailenco/msgpack"
)

type mc struct{}

func (*mc) Marshal(msg *stmp.Message) ([]byte, error) {
	if msg.Data == nil {
		return nil, nil
	}
	return msgpack.Marshal(msg.Data)
}

func (*mc) Unmarshal(msg *stmp.Message, output interface{}) error {
	if msg.Payload == nil || len(msg.Payload) == 0 {
		return nil
	}
	return msgpack.Unmarshal(msg.Payload, output)
}

func New() stmp.Codec {
	return &mc{}
}
