package stmp_test

import (
	"testing"
)

func TestUint16(t *testing.T) {
	var value uint16 = 0xFFFF
	t.Logf("origin: %d, added: %d", value, value + 1)
}

func covert(data interface{}) bool {
	if _, ok := data.([]byte); ok {
		return ok
	}
	return false
}

func TestInterface(t *testing.T) {
	data := []byte{0}
	t.Log(covert(data))
}

func appendSlice(input *[]byte) {
	*input = append(*input, 1, 2, 3)
}

func TestSlice(t *testing.T) {
	input := []byte{1, 2, 3}
	appendSlice(&input)
	t.Log(input)
}
