# STMP

The simplest message protocol.

*This is a protocol to organize your network communication, rather than serialize/unserialize message payload.*

Currently, the most popular message protocol in embed devices is MQTT, but it is so complex to handle `QoS`. This
protocol removed this feature, just use for send data, the `QoS` should be managed by the top application.

## Version

Current is `0.1`, drafting.

## Protocol

### Fixed Header

The fixed header include `1` byte, the definition and size as follow:

    |   0   |   1   |   2   |   3   |   4   |   5   |   6   |   7   |
    |     KIND      |   WP  |  WPS  |       Encoding        |   0   |

The means of each field as follow:

#### KIND

2bit, the message kind, the values means as follow:

- `0b00`: Ping Message
- `0b01`: Request Message
- `0b10`: Notify Message
- `0b11`: Response Message

#### WP

1bit, with payload or not, the values means as follow:

- `0b0`: without payload
- `0b1`: with payload

If this field is `0`, The `WPS` field **MUST** be `0`, and the `Encoding` field must be `0000`

#### WPS

1bit, with payload size or not, this is use for distinguish TCP/UDP and WebSockets, etc.
some protocol contains message size already.

- `0b0`: without payload size
- `0b1`: with payload size

If this field is `1`, The `WP` field **MUST** be `1`.

#### Encoding

3bit, this means the payload encoding type, just like `Content-Type` in HTTP protocol, but this is a flag to
represent it. This field means maybe different in different sense, according to the two peer how to comprehend it.
But there is some reserved values as follow:

- `0`: Reserved, means the payload is a raw binary bytes
- `1`: Protocol Buffers, see [Protocol Buffers](https://developers.google.com/protocol-buffers/)
- `2`: JSON, see [JSON](http://www.json.org)
- `3`: MessagePack, see [MessagePack](http://msgpack.org/index.html)
- `4`: BSON, see [BSON](http://bsonspec.org/)

### Optional headers

If any field of the follow headers exists, **SHOULD** arrange in the follow order.

#### ID

2bytes, the message id, from `0x0000` to `0xFFFF`, this is determined by the `KIND` field

#### ACTION

4bytes, the asking action id, from `0x00000000` to `0xFFFFFFFF`, this is use for application to distinguish the
request resource.

#### STATUS

1byte, the response status code, from `0x00` to `0xFF`, this is use for response message, the codes from `0x00` to `0x7F`
is reserved for internal usages. And the codes from `0x80` to `0xFF` is user defined. The reserved code list as follow:
(just change the code value from http)

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

#### PS

4bytes, the payload size, from `0x00000000` to `0xFFFFFFFF`, this is determined by the `WP` and `WPS` field

#### PAYLOAD

The size is determined by `PS` field, the format is determined by `Encoding` field.

### Messages

#### Ping Message

This is a heartbeat packet, **SHOULD NOT** with payload, that means the fixed header must be `0b00000000`.
This message should not be replied. Each peer should keep a timer to send this packet, if a peer does not receive
this peer in time, **MUST** close the connection immediately.

#### Request Message

This means a request from a peer, the other peer should send response to the peer in time, if the response is timeout,
the peer should emit a timeout error to application.

This message **MAYBE** contains `PAYLOAD`, and **MUST** contains `ID` and `ACTION` field.

A peer received this message must send a `Response Message` to the other peer, and the `ID` is same to the message.

#### Notify Message

This means a notify message from a peer, the other peer should not response to the peer.

This message **MAYBE** contains `PAYLOAD` field, **MUST NOT** contains `ID` field, and **MUST** contains `ACTION` field.

#### Response Message

This means a response message for a `Request Message`, the `ID` must same to the request.

This message **MAYBE** contains `PAYLOAD`, **MUST NOT** contains `ACTION` field, and **MUST** contains `ID` field.
