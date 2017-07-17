package stmp_test

import (
	"testing"
	"."
	"github.com/stretchr/testify/assert"
)

func TestProtocolVersion(t *testing.T) {
	assert.Equal(t, byte(0), stmp.StmpVersion.Major, "major version")
	assert.Equal(t, byte(1), stmp.StmpVersion.Minor, "minor version")
	assert.Equal(t, []byte{(0 << 4) + 1}, stmp.StmpVersion.Binary(), "binary marshal")
	assert.Equal(t, []byte{'0', '1'}, stmp.StmpVersion.Texture(), "texture marshal")
	assert.Equal(t, "0.1", stmp.StmpVersion.String(), "stringify")
}

func TestParseBinary(t *testing.T) {
	msg, err := stmp.ParseBinary([]byte{0})
	assert.Equal(t, nil, err, "parse ping ok")
	assert.Equal(t, stmp.KindPing, msg.Kind, "PING KIND")
	msg, err = stmp.ParseBinary([]byte{
		stmp.KindRequest | stmp.FlagWp | stmp.FlagWps | stmp.EncodingRaw, // Flags
		0xFC, 0xD1, // ID
		0x10, 0xFF, 0xDC, 0xF0, // ACTION
		0, 0, 0, 1, // PS
		0x30, // PAYLOAD
	})
	assert.Equal(t, nil, err, "parse REQ|WP|WPS|RAW ok")
	assert.Equal(t, stmp.KindRequest, msg.Kind, "REQ KIND")
	assert.Equal(t, uint16(0xFCD1), msg.Id, "REQ ID")
	assert.Equal(t, uint32(0x10FFDCF0), msg.Action, "REQ ACTION")
	assert.Equal(t, uint32(1), msg.PayloadSize, "REQ PS")
	assert.Equal(t, []byte{'0'}, msg.Payload, "REQ PAYLOAD")
}
