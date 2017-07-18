# STMP

The simplest message protocol.

*This is a protocol to organize your network communication, rather than serialize/unserialize message payload.*

Currently, the most popular message protocol in embed devices is [MQTT](http://mqtt.org/), but it is so complex
to handle `QoS`. This protocol removed this feature, just use for send data, the `QoS` should be managed by the
upper application.

## Version

The version contains two fields: `MAJOR` and `MINOR`. Each one field ranges from `0` to `15`, `0` to `0xF` in hex.
And could be print as string in format `MAJOR.MINOR`.

The version is `0.1` currently, in drafting.

## Message Fields

A message includes any fields of the follow, and some fields is required, others is optional. If any field of
the following headers exists, **MUST** be arranged in the following order.

### KIND

**Required**, the message kind, the values means as follow:

- `0`: `Ping Message`
- `1`: `Request Message`
- `2`: `Notify Message`
- `3`: `Response Message`

### ENCODING

**Required**, this means the payload encoding type, just like `Content-Type` in HTTP protocol, but this is a flag to
represent it. This field means maybe different in different sense, according to the two peer how to comprehend it.
But there is some reserved values as follow:

- `0`: Reserved, means **without payload**
- `1`: Protocol Buffers, see [Protocol Buffers](https://developers.google.com/protocol-buffers/)
- `2`: JSON, see [JSON](http://www.json.org)
- `3`: MessagePack, see [MessagePack](http://msgpack.org/index.html)
- `4`: BSON, see [BSON](http://bsonspec.org/)
- `5`: Raw, means the payload is a raw binary bytes or string

### ID

**Optional**, the message id, from `0x0000` to `0xFFFF`, this is determined by the `KIND` field.

### ACTION

**Optional**, the request action id, from `0x00000000` to `0xFFFFFFFF`, this is use for application to distinguish the
request resource.

Actions from `0x00` to `0xFF` is reserved, some of those is using now, and list as follow:

- `0x00`: CheckVersion, this is use for check the protocol version.

#### STATUS

**Optional**, the response status code, from `0x00` to `0xFF`, this is use for response message, the codes from
`0x00` to `0x7F` is reserved for internal usages. And the codes from `0x80` to `0xFF` is user defined. The reserved
code list as follow: (just change the code value from http)

- `0x00`: Ok, 200
- `0x10`: MovedPermanently, 301
- `0x11`: Found, 302
- `0x12`: NotModified, 304
- `0x20`: BadRequest, 400
- `0x21`: Unauthorized, 401
- `0x22`: PaymentRequired, 402
- `0x23`: Forbidden, 403
- `0x24`: NotFound, 404
- `0x25`: RequestTimeout, 408
- `0x26`: RequestEntityTooLarge, 413
- `0x27`: TooManyRequests, 429
- `0x30`: InternalServerError, 500
- `0x31`: NotImplemented, 501
- `0x32`: BadGateway, 502
- `0x33`: ServiceUnavailable, 503
- `0x34`: GatewayTimeout, 504
- `0x35`: VersionNotSupported, 505

### PS

**Optional**, the payload size, from `0x00000000` to `0xFFFFFFFF`, this is determined by the network protocol
environment, so this is a configure option when parse/serialize a message. If a message without payload (`ENCODING`
is `0`), this field **SHOULD NOT** exists in any protocol.

### PAYLOAD

**Optional**, the size is determined by `ENCODING` field, if it is `0`, this field should not exists, else should
be marshaled by the codec bind to the `ENCODING` type.

## Messages

### Ping Message

This is a heartbeat packet. This message should not be replied. Each peer should keep a timer to send this packet,
if a peer does not receive this message in time, **MUST** close the connection immediately.

This message **MUST NOT** contains any optional fields, that means all the fixed fields value is `0`.

### Request Message

This means a request from a peer, the other peer should send response to the peer in time, if the response is timeout,
the peer should emit a timeout error to application. A peer received this message must send a `Response Message` to
the other peer, and the `ID` is same to the message.

This message **MAYBE** contains `PAYLOAD` and `PS`, **MUST NOT** contains `STATUS`, **MUST** contains `ID` and `ACTION`.

### Notify Message

This means a notify message from a peer, the other peer should not response to the peer.

This message **MAYBE** contains `PAYLOAD` and `PS`, **MUST NOT** contains `ID` and `STATUS`, **MUST** contains `ACTION`.

### Response Message

This means a response message for a `Request Message`, the `ID` must same to the request message.

This message **MAYBE** contains `PAYLOAD` and `PS`, **MUST NOT** contains `ACTION`, **MUST** contains `ID` and `STATUS`.

## Binary Message Protocol

In most cases, the message protocol is use for embed device, the environment could handle bytes easily, and
serialize payload with binary protocol just like Protocol Buffers, MessagePack is fast. The protocol use binary
struct directly.

NOTE:

- all multi-bytes flags/value use `BE` format.

The entire message structure as follow:

    |   0 ... 7   |  8 ... 15  |  16 ... 23  |  24 ... 31  |
    | FixedHeader |           ID             |    ACTION   |
    |               ACTION                   |    STATUS   |
    |                         PS                           |
    |                 PAYLOAD    ...                       |

The first 1 byte is for fixed header, it exists in any kind of message, and other field is determined by the fixed
header. If the field should not exists, it will not exist, and the follow-up fields will move forward.

The fixed header structure as follow:

    |   0   |   1   |   2   |   3   |   4   |   5   |   6   |   7   |
    |     KIND      |       ENCODING        |   0   |   0   |   0   |
    
The last one bit is reserved because it is useless currently.

## Texture Message Protocol

Sometimes, specially, in browser, the environment does not support manipulate bytes directly, or the performance is
poor. So use string is better rather than binary(in this case is Uint8Array). So, this is a special case to handle it.

The case includes the following features could use this:

1. The network protocol could distinguish binary/string message directly.
2. The environment could handle UTF-8 encoded string.

If the message is a binary bytes, is same to upon, else the message should be:

All fields and message types is same to upon, and we just need to change the serialize result.

All fields is joined by string `|`, that means a full message format is follow:

```text
KIND(1)|ENCODING(1)|ID?(1-5)|ACTION?(1-10)|STATUS?(1-3)|PS?(1-10)|PAYLOAD?(...)
```

For a specified kind of message, some fields maybe not exists, and that field will not exist in the text. For example,
for a `Request Message`, without `PAYPLOAD` and `PS`, the message should be as follow:

```text
1|0|ID|ACTION
```

Specially, for a `Ping Message` all field is `0`, So, a `Ping Message` should be serialized as a 1 byte string `'0'`.

```text
0
```

## Distinguish Texture and Binary Message

As the protocol definitions. If a message is texture, the first byte must be one of the chars
`'0'`, `'1'`, `'2'`, `'3'`, which means `0x30`, `0x31`, `0x32`, `0x33` in hex. And if a message is binary, the
first byte must be one of the following case:

- be `0x00`, this is a ping message
- greater than `0b01000000`, the first 2 bits is the flag of `KIND`, if is `Ping Message`, the entire byte **MUST**
be `0x00`, else the first 2 bits **MUST NOT** be `0x00`, so the value must greater than `0b01000000`, `0x40` in hex.

So, the first byte of one message is enough to distinguish texture and binary message.

## Actions

Some actions is reserved for protocol internal usage.

### Check Version 0x00

When the connection established, the client must send a `Request Message` with `Check Version` action to check version,
the payload of the request is the version list that the client could handle. The server must response a `Ok`
message to the client if it could handle one of the versions sent by client when the server receive the request. If not,
the server must response a `VersionNotSupported` message to the client and then close the connection.

The `Check Version Request Message` and the response's `ENCODING` **MUST** be `Raw`. The payload is concat the version
list directly.

### Serialize Protocol Version

If the message is binary, a version should cost 1 byte, the first 4 bits is the `MAJOR` field and the last 4 bits is
the `MINOR` version, both range from `0x0` to `0xF`.

If the message is texture, a version should cost 2 bytes, the first byte is the `MAJOR` field and the last byte is the
`MINOR` version, both range from `'0'` to `'F'`, case insensitive.

## License

MIT
