package bson

import (
	"../"
	b "gopkg.in/mgo.v2/bson"
)

type bc struct{}

func (*bc) Marshal(msg *stmp.Message) ([]byte, error) {
	if msg.Data == nil {
		return nil, nil
	}
	return b.Marshal(msg.Data)
}

func (*bc) Unmarshal(msg *stmp.Message, output interface{}) error {
	if msg.Payload == nil || len(msg.Payload) == 0 {
		return nil
	}
	return b.Unmarshal(msg.Payload, output)
}

func New() stmp.Codec {
	return &bc{}
}
