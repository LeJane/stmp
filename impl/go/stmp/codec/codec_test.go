package codec_test

import (
	"./bson"
	"./json"
	"./pb"
	"./mp"
	".."
	"testing"
)

func TestInterface(t *testing.T) {
	msg := stmp.NewMessage(stmp.KindRequest, stmp.EncodingRaw, 0, 1, []string{"hello", "world"})
	_ = stmp.Marshal(msg, bson.New())
	_ = stmp.Marshal(msg, json.New())
	_ = stmp.Marshal(msg, pb.New())
	_ = stmp.Marshal(msg, mp.New())
}
