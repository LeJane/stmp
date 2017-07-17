package stmp

const (
	PosEncoding byte = 1
	PosWps = 4
	PosWp = 5
	PosKind = 6
)

const (
	// golang does not support binary integer literal currently
	FlagEncoding byte = 7 << PosEncoding
	FlagWps = 1 << PosWps
	FlagWp = 1 << PosWp
	FlagKind = 3 << PosKind
)

const (
	EncodingRaw byte = 0 << PosEncoding
	EncodingProtocolBuffers = 1 << PosEncoding
	EncodingJson = 2 << PosEncoding
	EncodingMessagePack = 3 << PosEncoding
	EncodingBson = 4 << PosEncoding
)

const (
	KindPing byte = 0 << PosKind
	KindRequest = 1 << PosKind
	KindNotify = 2 << PosKind
	KindResponse = 3 << PosKind
)

const (
	StatusOk byte = 0x00

	StatusMovedPermanently = 0x10
	StatusFound = 0x11
	StatusNotModified = 0x12

	StatusBadRequest = 0x20
	StatusUnauthorized = 0x21
	StatusPaymentRequired = 0x22
	StatusForbidden = 0x23
	StatusNotFound = 0x24
	StatusRequestTimeout = 0x25
	StatusRequestEntityTooLarge = 0x26
	StatusTooManyRequests = 0x27

	StatusInternalServerError = 0x30
	StatusNotImplemented = 0x31
	StatusBadGateway = 0x32
	StatusServiceUnavailable = 0x33
	StatusGatewayTimeout = 0x34
)

type Message struct {
	Kind            byte
	WithPayload     byte
	WithPayloadSize byte
	Encoding        byte
	Id              uint16
	Action          uint32
	Status          byte
	PayloadSize     uint32
	Payload         []byte
	Data            interface{}
}

const PingMessage = &Message{
	Kind: KindPing,
	WithPayload: 0,
	WithPayloadSize: 0,
	Encoding: EncodingRaw,
	PayloadSize: 0,
	Payload: nil,
	Data: nil,
}

const PingBytes = []byte{0}
