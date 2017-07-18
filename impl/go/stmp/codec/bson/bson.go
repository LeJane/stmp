package bson

import (
	"../../"
	"gopkg.in/mgo.v2/bson"
)

type bc struct{}

func (*bc) Marshal(data interface{}) ([]byte, error) {
	return bson.Marshal(data)
}

func (*bc) Unmarshal(payload []byte, data interface{}) error {
	return bson.Unmarshal(payload, data)
}

func (*bc) Texture() bool {
	return false
}

func New() stmp.Codec {
	return &bc{}
}
