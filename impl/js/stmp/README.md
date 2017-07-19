# STMP

The simplest message protocol for network communication.

## Install

```bash
# Via npm
npm install -S stmp

# Via yarn
yarn add stmp
```

## Usage

### Browser Context

In browser: just support texture serialize message

```typescript
import {BrowserTextureContext} from './lib/browser'
import {Kind, Encoding, Status} from './lib/message'

const ctx = new BrowserTextureContext({})

// init a message
const msg = ctx.binMessage(Kind.Request, Encoding.Json, void 0, 0x12345678, void 0, JSON.stringify({"hello": "world"}))

// serialize it
const packet = ctx.stringify(msg)

// parse a message from string
ctx.parse(packet)

// Sometimes, we need to send binary data with WebSockets, we allow you to send header
// and payload in two packet. For example:
const payload = Uint8Array.of(0x01, 0x02, 0x03, 0x04)

const msg2 = ctx.binMessage(Kind.Response, Encoding.Raw, 0x1234, void 0, Status.Unauthorized, payload)

// serialize will generate two chunk, first one is string header, second one is original Uint8Array payload
const packets = ctx.stringify(msg)

// And then you can send it to WebSockets in two frame.
const ws = new WebSocket('')
ws.send(packets[0])
ws.send(packets[1])

// And you can parse two packets from WebSockets as follow:
ws.addEventListener('message', ev => {
  let msg = ctx.parse(ev.data)
  if (msg === void 0) {
    // continue
  } else if(msg === false) {
    // error occurred
    console.error(ctx.getError())
    ws.close()
  } else {
    // do any staff as a message
    console.log(msg)
  }
})
```

### Node Context

In node: we allow you to handle texture message and binary message.

The texture context is same to browser, just change the import path:

```typescript
import {NodeTextureContext} from './lib/node'

const ctx = new NodeTextureContext({})
// do any staff with the ctx
```

The binary context is different when parse data, because this is used for TCP/UDP etc. The message is continuous.
So it just work like a stream pipe line. the different as follow:

```typescript
import {NodeBinaryContext} from './lib/node'

const ctx = new NodeBinaryContext({})

const msgs = ctx.parse(Buffer.from([0x01]))
// the parse result is a message array rather than a message instance
// anything else is absolutely same to NodeTextureContext.
```

## License

MIT

## TODO

- [ ] Marshal/Unmarshal payload
- [ ] Server/Client/Router
