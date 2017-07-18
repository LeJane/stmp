package stmp

import (
	"encoding/binary"
	"io"
	"errors"
	"strconv"
)

const (
	PosEncoding byte = 2
	PosWp byte = 5
	PosKind byte = 6
)

const (
	// golang does not support binary integer literal currently
	FlagEncoding byte = 7 << PosEncoding
	FlagWp byte = 1 << PosWp
	FlagKind byte = 3 << PosKind
)

const (
	EncodingRaw byte = 0 << PosEncoding
	EncodingProtocolBuffers byte = 1 << PosEncoding
	EncodingJson byte = 2 << PosEncoding
	EncodingMessagePack byte = 3 << PosEncoding
	EncodingBson byte = 4 << PosEncoding
)

const (
	KindPing byte = 0 << PosKind
	KindRequest byte = 1 << PosKind
	KindNotify byte = 2 << PosKind
	KindResponse byte = 3 << PosKind
)

const (
	StatusOk byte = 0x00

	StatusMovedPermanently byte = 0x10
	StatusFound byte = 0x11
	StatusNotModified byte = 0x12

	StatusBadRequest byte = 0x20
	StatusUnauthorized byte = 0x21
	StatusPaymentRequired byte = 0x22
	StatusForbidden byte = 0x23
	StatusNotFound byte = 0x24
	StatusRequestTimeout byte = 0x25
	StatusRequestEntityTooLarge byte = 0x26
	StatusTooManyRequests byte = 0x27

	StatusInternalServerError byte = 0x30
	StatusNotImplemented byte = 0x31
	StatusBadGateway byte = 0x32
	StatusServiceUnavailable byte = 0x33
	StatusGatewayTimeout byte = 0x34
	VersionNotSupported byte = 0x35
)

type Message struct {
	Kind        byte
	WithPayload byte
	Encoding    byte
	Id          uint16
	Action      uint32
	Status      byte
	PayloadSize uint32
	Payload     []byte
	Data        interface{}
}

var (
	PingMessage = &Message{
		Kind: KindPing,
		WithPayload: 0,
		Encoding: EncodingRaw,
		Id: 0,
		Action: 0,
		Status: StatusOk,
		PayloadSize: 0,
		Payload: nil,
		Data: nil,
	}
	PingBytes = []byte{0}
	PingTexture = []byte{'0'}
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
	return []byte{v.Major + 0x30, v.Minor + 0x30}
}

func (v *ProtocolVersion) String() string {
	return strconv.Itoa(int(v.Major)) + "." + strconv.Itoa(int(v.Minor))
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
	output := make([]*ProtocolVersion, size)
	for i := 0; i < size; i += 2 {
		output[i] = &ProtocolVersion{
			Major: input[i] - 0x30,
			Minor: input[i + 1] - 0x30,
		}
	}
	return output
}

func NewMessage(kind byte, encoding byte, id uint16, action uint32, data interface{}) *Message {
	return &Message{
		Kind: kind,
		WithPayload: 0, // This will be determined by the serialize result
		Encoding: encoding,
		Id: id,
		Action: action,
		Status: StatusOk,
		PayloadSize: 0,
		Payload: nil,
		Data: data,
	}
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
	wp := header & FlagWp
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
		if wp == FlagWp {
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
		if wp == FlagWp {
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
		if wp == FlagWp {
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
	wp := fixed & FlagWp
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
		if wp == FlagWp && wps {
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
		if wp == FlagWp && wps {
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
		if wp == FlagWp && wps {
			if size < 8 {
				return nil, ErrInputTooShort
			}
			ps = binary.BigEndian.Uint32(data[4:])
			if size < 8 + ps {
				return nil, ErrInputTooShort
			}
			payload = data[8:]
		} else if wp == FlagWp {
			ps = size - 4
			payload = data[4:]
		}
		id = binary.BigEndian.Uint16(data[1:])
		status = data[3]
	}
	msg = &Message{
		Kind: kind,
		WithPayload: wp,
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
		return PingBytes
	}
	bufSize := 1 + msg.PayloadSize
	if msg.WithPayload == FlagWp && wps {
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
	buf[0] = msg.Kind | msg.WithPayload | msg.Encoding
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
	if msg.WithPayload == FlagWp && wps {
		binary.BigEndian.PutUint32(buf[bufSize - msg.PayloadSize - 4:], msg.PayloadSize)
	}
	if msg.WithPayload == FlagWp {
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
// and auto update WP, PS field
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
			msg.WithPayload = 0
			msg.PayloadSize = 0
			msg.Payload = nil
		} else {
			msg.WithPayload = FlagWp
			msg.PayloadSize = uint32(len(msg.Payload))
		}
	}
	return err
}

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
