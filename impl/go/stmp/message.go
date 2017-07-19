package stmp

import (
	"encoding/binary"
	"io"
	"errors"
)

const (
	OffsetEncoding byte = 3
	OffsetKind byte = 6
)

const (
	// 0b00111000
	FlagEncoding byte = 7 << OffsetEncoding
	// 0b11000000
	FlagKind byte = 3 << OffsetKind
)

const (
	// 0x00, 0
	EncodingNone byte = iota << OffsetEncoding
	// 0x08, 8
	EncodingProtocolBuffers
	// 0x10, 16
	EncodingJson
	// 0x18, 24
	EncodingMessagePack
	// 0x20, 32
	EncodingBson
	// 0x28, 40
	EncodingRaw
)

const (
	// 0x00, 0
	KindPing byte = iota << OffsetKind
	// 0x40, 64
	KindRequest
	// 0x80, 128
	KindNotify
	// 0xC0, 192
	KindResponse
)

const (
	StatusOk byte = 0x00

	// 16
	StatusMovedPermanently = 0x10
	// 17
	StatusFound = 0x11
	// 18
	StatusNotModified = 0x12

	// 32
	StatusBadRequest = 0x20
	// 33
	StatusUnauthorized = 0x21
	// 34
	StatusPaymentRequired = 0x22
	// 35
	StatusForbidden = 0x23
	// 36
	StatusNotFound = 0x24
	// 37
	StatusRequestTimeout = 0x25
	// 38
	StatusRequestEntityTooLarge = 0x26
	// 39
	StatusTooManyRequests = 0x27

	// 48
	StatusInternalServerError = 0x30
	// 49
	StatusNotImplemented = 0x31
	// 50
	StatusBadGateway = 0x32
	// 51
	StatusServiceUnavailable = 0x33
	// 52
	StatusGatewayTimeout = 0x34
	// 53
	VersionNotSupported = 0x35
)

type Message struct {
	Kind        byte
	Encoding    byte
	Id          uint16
	Action      uint32
	Status      byte
	PayloadSize uint32
	Payload     []byte
	Data        interface{}
}

var (
	PingBinary = []byte{0}
	PingTexture = []byte{'0'}
	PingString = "0"
	PingMessage = NewRawMessage(KindPing, EncodingNone, 0, 0, 0, nil)
)

var (
	ErrInputTooShort = errors.New("input is too short")
	ErrIncorrectDataType = errors.New("data type is incorrect to reflect marshal/unmarshal")
	ErrIncorrectPayloadType = errors.New("payload type is incorrect to marshal/unmarshal")
	ErrNotImplemented = errors.New("not implemented")
	ErrCodecIsNil = errors.New("codec is nil")
)

// serialize/unserialize payload
type Codec interface {
	// The data must not be nil pointer
	// If the err is not nil, will be emit
	// else if payload is nil or len(payload) == 0, the WP and WPS field will be set as 0
	Marshal(data interface{}) (payload []byte, err error)

	// The payload must not be nil pointer and len(payload) > 0
	Unmarshal(payload []byte, data interface{}) error

	// check the codec marshal result is text or binary
	Texture() bool
}

type ProtocolVersion struct {
	Major byte
	Minor byte
}

var StmpVersion = &ProtocolVersion{Major: 0, Minor: 1}

func (v *ProtocolVersion) Binary() []byte {
	return []byte{(v.Major << 4) + v.Minor}
}

func (v *ProtocolVersion) Texture() []byte {
	out := []byte{v.Major + 0x30, v.Minor + 0x30}
	if out[0] > 0x39 {
		out[0] += 0x27
	}
	if out[1] > 0x39 {
		out[1] += 0x27
	}
	return out
}

func (v *ProtocolVersion) String() string {
	return string([]byte{v.Major + 0x30, '.', v.Minor + 0x30})
}

func ParseBinaryVersions(input []byte) []*ProtocolVersion {
	output := make([]*ProtocolVersion, len(input))
	for index, version := range (input) {
		output[index] = &ProtocolVersion{
			Major: version >> 4,
			Minor: version & 0xF,
		}
	}
	return output
}

func ParseTextureVersions(input []byte) []*ProtocolVersion {
	size := len(input) / 2
	out := make([]*ProtocolVersion, size)
	for i := 0; i < size; i += 2 {
		out[i] = &ProtocolVersion{
			Major: input[i] - 0x30,
			Minor: input[i + 1] - 0x30,
		}
		if out[i].Major > 9 {
			out[i].Major -= 0x27
		}
		if out[i].Minor > 9 {
			out[i].Minor -= 0x27
		}
	}
	return out
}

// A original message that need to marshal payload according to the encoding
func NewRawMessage(kind byte, encoding byte, id uint16, action uint32, status byte, data interface{}) *Message {
	return &Message{kind, encoding, id, action, status, 0, nil, data}
}

// A message that payload is set, and the data need to wait for unmarshal
func NewBinMessage(kind byte, encoding byte, id uint16, action uint32, status byte, payload []byte) *Message {
	msg := &Message{kind, encoding, id, action, status, uint32(len(payload)), payload, nil}
	if msg.PayloadSize == 0 {
		msg.Encoding = EncodingNone
		msg.Payload = nil
	}
	return msg
}

// read binary message from a reader
// this must with payload size, because we cannot
// split multi-message automatically
func readBinaryWithHeader(r io.Reader, header byte) (msg *Message, err error) {
	kind := header & FlagKind
	if kind == KindPing {
		// ping message
		return PingMessage, nil
	}
	// MUST be STMP_FLAG_WPS, else the payload will be ignored,
	// because we cannot get the payload size from a reader
	encoding := header & FlagEncoding
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
		if encoding != EncodingNone {
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
		if encoding != EncodingNone {
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
		if _, err = io.ReadFull(r, byte4[0:1]); err != nil {
			// STATUS
			return
		}
		status = byte4[0]
		if encoding != EncodingNone {
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
	msg = NewBinMessage(kind, encoding, id, action, status, payload)
	return
}

// read message from a reader
func ReadBinary(r io.Reader) (*Message, error) {
	byte1 := make([]byte, 1)
	_, err := io.ReadFull(r, byte1)
	if err != nil {
		return nil, err
	}
	return readBinaryWithHeader(r, byte1[0])
}

// read a message with texture protocol
func readTextureWithHeader(r io.Reader, header byte) (msg *Message, err error) {
	kind := (header - 0x30) << 6
	if kind == KindPing {
		return PingMessage, nil
	}
	// TODO implements this
	// this may not be implemented forever, because the bufio.Reader
	// will read data more than delimiter, so the connection will be
	// consumed more than one message. This maybe refactored by change
	// input parameter, just like wrap a connection to a bufio.Reader
	return nil, ErrNotImplemented
}

// auto distinguish binary and texture
// read message from a message must contains PS field
func Read(r io.Reader) (*Message, error) {
	byte1 := make([]byte, 1)
	_, err := io.ReadFull(r, byte1)
	if err != nil {
		return nil, err
	}
	header := byte1[0]
	if header < 0x40 && header != 0x00 {
		return readTextureWithHeader(r, header)
	} else {
		return readBinaryWithHeader(r, header)
	}
}

// parse from a message
func ParseBinary(data []byte, wps bool) (msg *Message, err error) {
	size := uint32(len(data))
	if size == 0 {
		err = ErrInputTooShort
		return
	}
	fixed := data[0]
	kind := fixed & FlagKind
	if kind == KindPing {
		return PingMessage, nil
	}
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
		if encoding != EncodingNone && wps {
			if size < 11 {
				// +PS(4)
				return nil, ErrInputTooShort
			}
			ps = binary.BigEndian.Uint32(data[7:])
			if size < 11 + ps {
				return nil, ErrInputTooShort
			}
			payload = data[11:]
		} else if encoding != EncodingNone {
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
		if encoding != EncodingNone && wps {
			if size < 9 {
				return nil, ErrInputTooShort
			}
			ps = binary.BigEndian.Uint32(data[5:])
			if size < 9 + ps {
				return nil, ErrInputTooShort
			}
			payload = data[9:]
		} else if encoding != EncodingNone {
			ps = size - 5
			payload = data[5:]
		}
		action = binary.BigEndian.Uint32(data[1:])
	case KindResponse:
		if size < 4 {
			return nil, ErrInputTooShort
		}
		if encoding != EncodingNone && wps {
			if size < 8 {
				return nil, ErrInputTooShort
			}
			ps = binary.BigEndian.Uint32(data[4:])
			if size < 8 + ps {
				return nil, ErrInputTooShort
			}
			payload = data[8:]
		} else if encoding != EncodingNone {
			ps = size - 4
			payload = data[4:]
		}
		id = binary.BigEndian.Uint16(data[1:])
		status = data[3]
	}
	msg = NewBinMessage(kind, encoding, id, action, status, payload)
	return
}

// parse a message by texture protocol
func ParseTexture(data []byte, wps bool) (msg *Message, err error) {
	// TODO
	return
}

// wps: with PS field or not
func Parse(data []byte, wps bool) (*Message, error) {
	if data == nil || len(data) < 1 {
		return nil, ErrInputTooShort
	}
	if data[0] < 0x40 && data[0] != 0x00 {
		return ParseTexture(data, wps)
	} else {
		return ParseBinary(data, wps)
	}
}

// serialize a message to binary
// the PAYLOAD field is determined by the msg.WP, and the length is the msg.PS
// the PS field is determined by the param wps, and the value is msg.PayloadSize
func SerializeBinary(msg *Message, wps bool) []byte {
	if msg.Kind == KindPing {
		return PingBinary
	}
	bufSize := 1 + msg.PayloadSize
	if msg.Encoding != EncodingNone && wps {
		bufSize += 4
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
	buf[0] = msg.Kind | msg.Encoding
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
	if msg.Encoding != EncodingNone && wps {
		binary.BigEndian.PutUint32(buf[bufSize - msg.PayloadSize - 4:], msg.PayloadSize)
	}
	if msg.Encoding != EncodingNone {
		copy(buf[bufSize - msg.PayloadSize:], msg.Payload)
	}
	return buf
}

// serialize a message by texture protocol, the handle logic
// is same to serialize binary
func SerializeTexture(msg *Message) (output []byte, err error) {
	// TODO
	return
}

// marshal a message payload, and bind it to msg.Payload
// and auto update WP, PS field, you need to init encoding
// field before use this, so, a full progress to serialize
// a message is follow:
//
// 1. init a message: msg := NewMessage(Kind, Encoding, Id, Action, Status, Data)
// 2. marshal payload: Marshal(msg, json.New())
// 3. serialize it: reader.Write(SerializeBinary(msg, wps))
func Marshal(msg *Message, codec Codec) error {
	var err error
	var ok bool
	if msg.Data != nil {
		if msg.Encoding == EncodingRaw {
			msg.Payload, ok = msg.Data.([]byte)
			if !ok {
				err = ErrIncorrectDataType
			}
		} else {
			msg.Payload, err = codec.Marshal(msg.Data)
		}
	}
	if err == nil {
		if msg.Payload == nil || len(msg.Payload) == 0 {
			msg.Encoding = EncodingNone
			msg.PayloadSize = 0
			msg.Payload = nil
		} else {
			msg.PayloadSize = uint32(len(msg.Payload))
		}
	}
	return err
}

// This is use for unmarshal a message payload
// full stack:
//
// 1. read/parse a message from a reader/buffer
//		msg := Parse(buffer)
// 2. unmarshal it: Unmarshal(msg, binding, json.New())
//
// This will not update the Data field, maybe need to do this.
func Unmarshal(msg *Message, data interface{}, codec Codec) error {
	if msg.Payload == nil || len(msg.Payload) == 0 {
		return nil
	} else if msg.Encoding == EncodingRaw {
		// should never to here
		return nil
	} else if codec == nil {
		return ErrCodecIsNil
	}
	return codec.Unmarshal(msg.Payload, data)
}
