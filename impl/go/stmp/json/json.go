package json

import (
	j "encoding/json"
	"../"
)

type jc struct{}

func (*jc) Marshal(msg *stmp.Message) {
	if msg.Data == nil {
		return nil, nil
	}
	return j.Marshal(msg.Data)
}

func (*jc) Unmarshal(msg *stmp.Message, output interface{}) error {
	if msg.Payload == nil || len(msg.Payload) == 0 {
		return nil
	}
	return j.Unmarshal(msg.Payload, output)
}

func New() stmp.Codec {
	return &jc{}
}
