package stmp

import (
	"bytes"
	"errors"
)

var (
	ErrIncorrectInputType = errors.New("incorrect input type")
	ErrIncorrectOutputType = errors.New("incorrect output type")
)

// The codec just encode/decode the payload field
// It need to handle the nil pointer of a message.Data
// And if marshal returns a nil, nil, means without
// payload, else payload exists, and if the size of
// []byte is zero, means the payload size is 0
type Codec interface {
	Marshal(msg *Message) ([]byte, error)
	Unmarshal(msg *Message, output interface{}) error
}

type rawCodec struct{}

func (*rawCodec) Marshal(msg *Message) ([]byte, error) {
	if msg.Data == nil {
		return nil, nil
	} else if output, ok := msg.Data.([]byte); ok {
		return output, nil
	} else if buf, ok := msg.Data.(*bytes.Buffer); ok {
		return buf.Bytes(), nil
	} else {
		return nil, ErrIncorrectInputType
	}
}

func (*rawCodec) Unmarshal(msg *Message, output interface{}) error {
	if msg.Payload == nil {
		return nil
	} else if raw, ok := output.(*[]byte); ok {
		*raw = msg.Payload
		return nil
	} else if buf, ok := output.(*bytes.Buffer); ok {
		buf.Write(msg.Payload)
		return nil
	}
	return ErrIncorrectOutputType
}
