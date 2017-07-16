package stmp

import (
	"encoding/binary"
	"io"
	"errors"
)

var (
	ErrInputTooShort = errors.New("input is too short")
)

type Context struct {
	WithPayloadSize byte
	Encoding        byte
	NextId          uint16
	Codecs          map[byte]Codec
}

func NewContext(withPayloadSize byte, encoding byte, nextId uint16) *Context {
	ctx := &Context{
		WithPayloadSize: withPayloadSize,
		Encoding: encoding,
		NextId: nextId,
		Codecs: map[byte]Codec{},
	}
	ctx.SetCodec(EncodingRaw, &rawCodec{})
	return ctx
}

func (c *Context) SetCodec(encoding byte, codec Codec) {
	c.Codecs[encoding] = codec
}

func (c *Context) NewRequest(action uint32, payload interface{}) *Message {
	id := c.NextId
	c.NextId++ // will be auto chunk
	return &Message{
		Kind: KindRequest,
		WithPayload: 0,
		WithPayloadSize: c.WithPayloadSize,
		Encoding: c.Encoding,
		Id: id,
		Action: action,
		Status: StatusOk,
		PayloadSize: 0,
		Payload: nil,
		Data: payload,
	}
}

func (c *Context) NewResponse(mid uint16, status byte, payload interface{}) *Message {
	return &Message{
		Kind: KindResponse,
		WithPayload: 0,
		WithPayloadSize: c.WithPayloadSize,
		Encoding: c.Encoding,
		Id: mid,
		Action: 0,
		Status: status,
		PayloadSize: 0,
		Payload: nil,
		Data: payload,
	}
}

func (c *Context) NewNotify(action uint32, payload interface{}) {
	return &Message{
		Kind: KindNotify,
		WithPayload: 0,
		WithPayloadSize: c.WithPayloadSize,
		Encoding: c.Encoding,
		Id: 0,
		Action: action,
		Status: StatusOk,
		PayloadSize: 0,
		Payload: nil,
		Data: payload,
	}
}

// read message from a reader
func (*Context) Read(r io.Reader) (msg *Message, err error) {
	byte1 := make([]byte, 1)
	_, err = io.ReadFull(r, byte1)
	if err != nil {
		return
	}
	fixed := byte1[0]
	kind := fixed & FlagKind
	if kind == KindPing {
		// ping message
		return PingMessage, nil
	}
	wp := fixed & FlagWp
	wps := fixed & FlagWps
	// MUST be STMP_FLAG_WPS, else the payload will be ignored,
	// because we cannot get the payload size from a reader
	encoding := fixed & FlagEncoding
	var id uint16
	var action uint32
	var status byte
	var ps uint32
	var payload []byte
	byte4 := make([]byte, 4)
	switch kind {
	case KindRequest:
		if _, err = io.ReadFull(r, byte4[:2]); err != nil {
			// MID
			return
		}
		id = binary.BigEndian.Uint16(byte4)
		if _, err = io.ReadFull(r, byte4); err != nil {
			// ACTION
			return
		}
		action = binary.BigEndian.Uint32(byte4)
		if wp == FlagWp && wps == FlagWps {
			if _, err = io.ReadFull(r, byte4); err != nil {
				// PS
				return
			}
			ps = binary.BigEndian.Uint32(byte4)
			payload = make([]byte, ps)
			if _, err = io.ReadFull(r, payload); err != nil {
				// PAYLOAD
				return
			}
		}
	case KindNotify:
		if _, err = io.ReadFull(r, byte4); err != nil {
			// ACTION
			return
		}
		action = binary.BigEndian.Uint32(byte4)
		if wp == FlagWp && wps == FlagWps {
			if _, err = io.ReadFull(r, byte4); err != nil {
				// PS
				return
			}
			ps = binary.BigEndian.Uint32(byte4)
			payload = make([]byte, ps)
			if _, err = io.ReadFull(r, payload); err != nil {
				// PAYLOAD
				return
			}
		}
	case KindResponse:
		if _, err = io.ReadFull(r, byte4[:2]); err != nil {
			// MID
			return
		}
		id = binary.BigEndian.Uint16(byte4)
		if _, err = io.ReadFull(r, byte1); err != nil {
			// STATUS
			return
		}
		status = byte1[0]
		if wp == FlagWp && wps == FlagWps {
			if _, err = io.ReadFull(r, byte4); err != nil {
				// PS
				return
			}
			ps = binary.BigEndian.Uint32(byte4)
			payload = make([]byte, ps)
			if _, err = io.ReadFull(r, payload); err != nil {
				// PAYLOAD
				return
			}
		}
	}
	msg = &Message{
		Kind: kind,
		WithPayload: wp,
		WithPayloadSize: wps,
		Encoding: encoding,
		Id: id,
		Action: action,
		Status: status,
		PayloadSize: ps,
		Payload: payload,
		Data: nil,
	}
	return
}

func (*Context) Parse(data []byte) (msg *Message, err error) {
	size := len(data)
	if size == 0 {
		err = ErrInputTooShort
		return
	}
	fixed := data[0]
	kind := fixed & FlagKind
	if kind == KindPing {
		return PingMessage, nil
	}
	wp := fixed & FlagWp
	wps := fixed & FlagWps
	encoding := fixed & FlagEncoding
	var id uint16
	var action uint32
	var status byte
	var ps uint32
	var payload []byte
	switch kind {
	case KindRequest:
		if size < 7 {
			// FIXED+MID+ACTION(1+2+4)
			return nil, ErrInputTooShort
		}
		if wp == FlagWp && wps == FlagWps {
			if size < 11 {
				// +PS(4)
				return nil, ErrInputTooShort
			}
			ps = binary.BigEndian.Uint32(data[7:])
			if size < 11 + ps {
				return nil, ErrInputTooShort
			}
			payload = data[11:]
		} else if wp == FlagWp {
			// auto get size
			ps = size - 7
			payload = data[7:]
		}
		id = binary.BigEndian.Uint16(data[1:])
		action = binary.BigEndian.Uint32(data[3:])
	case KindNotify:
		if size < 5 {
			return nil, ErrInputTooShort
		}
		if wp == FlagWp && wps == FlagWps {
			if size < 9 {
				return nil, ErrInputTooShort
			}
			ps = binary.BigEndian.Uint32(data[5:])
			if size < 9 + ps {
				return nil, ErrInputTooShort
			}
			payload = data[9:]
		} else if wp == FlagWp {
			ps = size - 5
			payload = data[5:]
		}
		action = binary.BigEndian.Uint32(data[1:])
	case KindResponse:
		if size < 4 {
			return nil, ErrInputTooShort
		}
		if wp == FlagWp && wps == FlagWps {
			if size < 8 {
				return nil, ErrInputTooShort
			}
			ps = binary.BigEndian.Uint32(data[3:])
			if size < 8 + ps {
				return nil, ErrInputTooShort
			}
			payload = data[7:]
		} else if wp == FlagWp {
			ps = size - 4
			payload = data[3:]
		}
		id = binary.BigEndian.Uint16(data[1:])
		status = data[3]
	}
	msg = &Message{
		Kind: kind,
		WithPayload: wp,
		WithPayloadSize: wps,
		Encoding: encoding,
		Id: id,
		Action: action,
		Status: status,
		PayloadSize: ps,
		Payload: payload,
		Data: nil,
	}
	return
}

func (c *Context) Marshal(msg *Message) ([]byte, error) {
	if msg.Kind == KindPing {
		return PingBytes, nil
	}
	var err error
	bufSize := 1
	msg.Payload, err = c.Codecs[msg.Encoding].Marshal(msg)
	if err != nil {
		return nil, err
	}
	if msg.Payload == nil {
		// NO PAYLOAD AND PAYLOAD SIZE
		msg.WithPayloadSize = 0
		msg.WithPayload = 0
	} else {
		// WP
		msg.WithPayload = FlagWp
		msg.PayloadSize = len(msg.Payload)
		bufSize += msg.PayloadSize
		if msg.WithPayloadSize {
			// WPS
			bufSize += 4
		}
	}
	switch msg.Kind {
	case KindRequest:
		bufSize += 6
	case KindNotify:
		bufSize += 4
	case KindResponse:
		bufSize += 3
	}
	buf := make([]byte, bufSize)
	buf[0] = msg.Kind | msg.WithPayload | msg.WithPayloadSize | msg.Encoding
	switch msg.Kind {
	case KindRequest:
		binary.BigEndian.PutUint16(buf[1:], msg.Id)
		binary.BigEndian.PutUint32(buf[3:], msg.Action)
	case KindNotify:
		binary.BigEndian.PutUint32(buf[1:], msg.Action)
	case KindResponse:
		binary.BigEndian.PutUint16(buf[1:], msg.Id)
		buf[3] = msg.Status
	}
	if msg.WithPayloadSize {
		binary.BigEndian.PutUint32(buf[bufSize - msg.PayloadSize - 4:], msg.PayloadSize)
	}
	copy(buf[bufSize - msg.PayloadSize:], msg.Payload)
	return buf, nil
}

func (c *Context) Unmarshal(msg *Message, output interface{}) error {
	return c.Codecs[msg.Encoding].Unmarshal(msg, output)
}
