# STMP

The simplest message protocol.

## Install

```bash
go get github.com/acrazing/stmp/impl/go/stmp...
```

## Usage

At first, you need to init a context for your application:

```go
package main

import (
 "github.com/acrazing/stmp/impl/go/stmp"
)

var ctx = stmp.NewContext(stmp.FlagWps, stmp.EncodingJson, 0)
```

And then, you need to mount your codec to encode/decode marshal for specified encoding type. For example: mount
default json codec:

```go
package main

import (
  "github.com/acrazing/stmp/impl/go/stmp"
  "github.com/acrazing/stmp/impl/go/stmp/json"
)

var ctx = stmp.NewContext(stmp.FlagWps, stmp.EncodingJson, 0)

func init()  {
  ctx.SetCodec(stmp.EncodingJson, json.New())
}
```

And then, you can use it as follow:

```go
package main

import (
  "github.com/acrazing/stmp/impl/go/stmp"
  "github.com/acrazing/stmp/impl/go/stmp/json"
)

var ctx = stmp.NewContext(stmp.FlagWps, stmp.EncodingJson, 0)

func init()  {
  ctx.SetCodec(stmp.EncodingJson, json.New())
}

func main()  {
  // create a request message
  msg := ctx.NewRequest(0, nil)
  
  // serialize it
  bytes, err := ctx.Marshal(msg)
  if err != nil {
    return 
  }
  
  // send to a conn
  conn.Write(bytes)
  
  // read a message from conn
  msg, err = ctx.Read(conn)
  if err != nil {
    return 
  }
  // OR, read from a WebSocket conn
  msgType, data, err = conn.ReadMessage()
  if msgType != websocket.BinaryMessage || err != nil {
    return 
  }
  msg, err = ctx.Parse(data)
  if err != nil {
    return
  }
  
  // unmarshal a message
  var payload struct {
    Username string
    Password string
  }
  if err = ctx.Unmarshal(msg, &payload); err != nil {
    return
  }
  print(payload.Username, payload.Password)
}
```

## License

MIT

## TODO

[ ] test
[ ] server
[ ] client
