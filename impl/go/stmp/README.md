# STMP

The simplest message protocol.

## Install

```bash
go get github.com/acrazing/stmp/impl/go/stmp
```

## Usage

### Use message directly

```go
package main

import (
  "github.com/acrazing/stmp/impl/go/stmp"
  "github.com/acrazing/stmp/impl/go/stmp/codec/json"
  "bytes"
  "github.com/stretchr/testify/assert"
)

func main()  {
  // init a new message with some payload
  msg := stmp.NewMessage(stmp.KindRequest, stmp.EncodingJson, 0x1234, 0x56789ABC, []string{"hello", "world"})
  
  // marshal a message according to your codec, TODO use context to auto load codecs
  _ = stmp.Marshal(msg, json.New())
  
  // serialize a message use binary protocol, and with PS field
  output := stmp.SerializeBinary(msg, true)
  // as a result, output content should be as follow:
  expected := []byte{
    0x68, // the fixed header, 0b01101000 in binary
    0x12, 0x34, // the ID field
    0x56, 0x78, 0x9A, 0xBC, // the ACTION field
    0, 0, 0, 0x11, // the PS field, 17 bytes: '["hello","world"]'
    '[', '"', 'h', 'e', 'l', 'l', 'o', '"', ',',  // the json payload 'hello'
    '"', 'w', 'o', 'r', 'l', 'd', '"', ']', // the json payload 'world'
  }
  assert.Equal(nil, expected, output, "the marshal result")
  
  // read from a reader
  buf := bytes.NewBuffer(output)
  m, _ := stmp.ReadBinary(buf)
  assert.Equal(nil, m, msg, "read binary with payload size")
  
  // parse from a bytes slice
  m1, _ := stmp.ParseBinary(output, true)
  assert.Equal(nil, m1, msg, "parse binary with payload size")
}
```

## License

MIT

## TODO

- [ ] texture implements
- [x] test
- [ ] context
- [ ] server
- [ ] client
