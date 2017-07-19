package stmp_test

import (
	"testing"
	"."
	"github.com/stretchr/testify/assert"
	"bytes"
)

type Inst struct {
	msg *stmp.Message
	wps []byte
	nps []byte
}

var (
	okCases = map[string]Inst{
		"ping": {
			stmp.NewBinMessage(stmp.KindPing, 0, 0, 0, 0, nil),
			[]byte{0},
			[]byte{0},
		},
		"request:!PAYLOAD": {
			stmp.NewBinMessage(stmp.KindRequest, 0, 0x01DA, 0x00F10CD2, 0, nil),
			[]byte{stmp.KindRequest, 0x01, 0xDA, 0x00, 0xF1, 0x0C, 0xD2},
			[]byte{stmp.KindRequest, 0x01, 0xDA, 0x00, 0xF1, 0x0C, 0xD2},
		},
		"request": {
			stmp.NewBinMessage(stmp.KindRequest, stmp.EncodingJson, 0x1234, 0x12345678, 0, []byte("1.2345")),
			[]byte{stmp.KindRequest | stmp.EncodingJson, 0x12, 0x34, 0x12, 0x34, 0x56, 0x78, 0, 0, 0, 6, '1', '.', '2', '3', '4', '5'},
			[]byte{stmp.KindRequest | stmp.EncodingJson, 0x12, 0x34, 0x12, 0x34, 0x56, 0x78, '1', '.', '2', '3', '4', '5'},
		},
		"notify:!PAYLOAD": {
			stmp.NewBinMessage(stmp.KindNotify, 0, 0, 0x12345678, 0, nil),
			[]byte{stmp.KindNotify, 0x12, 0x34, 0x56, 0x78},
			[]byte{stmp.KindNotify, 0x12, 0x34, 0x56, 0x78},
		},
		"notify": {
			stmp.NewBinMessage(stmp.KindNotify, stmp.EncodingJson, 0, 0x12345678, 0, []byte("1.2345")),
			[]byte{stmp.KindNotify | stmp.EncodingJson, 0x12, 0x34, 0x56, 0x78, 0, 0, 0, 6, '1', '.', '2', '3', '4', '5'},
			[]byte{stmp.KindNotify | stmp.EncodingJson, 0x12, 0x34, 0x56, 0x78, '1', '.', '2', '3', '4', '5'},
		},
		"response:!PAYLOAD": {
			stmp.NewBinMessage(stmp.KindResponse, 0, 0x1234, 0, stmp.StatusOk, nil),
			[]byte{stmp.KindResponse, 0x12, 0x34, stmp.StatusOk},
			[]byte{stmp.KindResponse, 0x12, 0x34, stmp.StatusOk},
		},
		"response": {
			stmp.NewBinMessage(stmp.KindResponse, stmp.EncodingJson, 0x1234, 0, stmp.StatusOk, []byte("1.2345")),
			[]byte{stmp.KindResponse | stmp.EncodingJson, 0x12, 0x34, stmp.StatusOk, 0, 0, 0, 6, '1', '.', '2', '3', '4', '5'},
			[]byte{stmp.KindResponse | stmp.EncodingJson, 0x12, 0x34, stmp.StatusOk, '1', '.', '2', '3', '4', '5'},
		},
	}
)

func TestProtocolVersion(t *testing.T) {
	assert.Equal(t, byte(0), stmp.StmpVersion.Major, "major version")
	assert.Equal(t, byte(1), stmp.StmpVersion.Minor, "minor version")
	assert.Equal(t, []byte{(0 << 4) + 1}, stmp.StmpVersion.Binary(), "binary marshal")
	assert.Equal(t, []byte{'0', '1'}, stmp.StmpVersion.Texture(), "texture marshal")
	assert.Equal(t, "0.1", stmp.StmpVersion.String(), "stringify")
}

func TestMessage(t *testing.T) {
	for name, inst := range okCases {
		msg, err := stmp.ReadBinary(bytes.NewBuffer(inst.wps))
		assert.Nil(t, err, name + ": read wps error")
		assert.Equal(t, inst.msg, msg, name + ": read wps result")
		msg, err = stmp.ParseBinary(inst.wps, true)
		assert.Nil(t, err, name + ": parse wps error")
		assert.Equal(t, inst.msg, msg, name + ": parse wps result")
		msg, err = stmp.ParseBinary(inst.nps, false)
		assert.Nil(t, err, name + ": parse nps error")
		assert.Equal(t, inst.msg, msg, name + ": parse nps result")
		buf := stmp.SerializeBinary(msg, true)
		assert.Equal(t, inst.wps, buf, name + ": serialize wps")
		buf = stmp.SerializeBinary(inst.msg, false)
		assert.Equal(t, inst.nps, buf, name + ": serialize nps")
	}
}
