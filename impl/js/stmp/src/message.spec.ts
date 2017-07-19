/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-18 19:48:23
 * @version 1.0.0
 * @desc message.spec.ts
 */


import assert = require('assert')
import {BrowserTextureContext, BrowserTextureMessage} from './browser'
import {NodeBinaryContext, NodeBinaryMessage, NodeTextureContext, NodeTextureMessage} from './node'
import {Encoding, Kind, Status} from './message'

const bt = new BrowserTextureContext({})
const nt = new NodeTextureContext({})
const nb = new NodeBinaryContext({})

interface Case {
  btm: BrowserTextureMessage;
  btb: string | [string, Uint8Array];
  ntm: NodeTextureMessage;
  ntb: string | [string, Buffer];
  nbm: NodeBinaryMessage;
  nbb: Buffer;
}

const cases: { [key: string]: Case } = {
  PING: {
    btm: bt.binMessage(),
    btb: '0',
    ntm: nt.binMessage(),
    ntb: '0',
    nbm: nb.binMessage(),
    nbb: Buffer.from([0]),
  },
  REQUEST_NO_PAYLOAD: {
    btm: bt.binMessage(Kind.Request, void 0, void 0, 0x12345678),
    btb: '1|0|0|305419896',
    ntm: nt.binMessage(Kind.Request, void 0, void 0, 0x12345678),
    ntb: '1|0|0|305419896',
    nbm: nb.binMessage(Kind.Request, void 0, void 0, 0x12345678),
    nbb: Buffer.from([Kind.Request, 0x00, 0x00, 0x12, 0x34, 0x56, 0x78]),
  },
  REQUEST_PAYLOAD: {
    btm: bt.binMessage(Kind.Request, Encoding.Json, void 0, 0x12345678, void 0, '["hello","world"]'),
    btb: '1|2|1|305419896|["hello","world"]',
    ntm: nt.binMessage(Kind.Request, Encoding.Json, void 0, 0x12345678, void 0, '["hello","world"]'),
    ntb: '1|2|1|305419896|["hello","world"]',
    nbm: nb.binMessage(Kind.Request, Encoding.Json, void 0, 0x12345678, void 0, Buffer.from(
      [91, 34, 104, 101, 108, 108, 111, 34, 44, 34, 119, 111, 114, 108, 100, 34, 93],
    )),
    nbb: Buffer.from([
      Kind.Request | Encoding.Json, 0x00, 0x01, 0x12, 0x34, 0x56, 0x78, 0x00, 0x00, 0x00, 0x11,
      91, 34, 104, 101, 108, 108, 111, 34, 44, 34, 119, 111, 114, 108, 100, 34, 93, // ["hello","world"]
    ]),
  },
  NOTIFY_NO_PAYLOAD: {
    btm: bt.binMessage(Kind.Notify, void 0, void 0, 0x12345678),
    btb: '2|0|305419896',
    ntm: nt.binMessage(Kind.Notify, void 0, void 0, 0x12345678),
    ntb: '2|0|305419896',
    nbm: nb.binMessage(Kind.Notify, void 0, void 0, 0x12345678),
    nbb: Buffer.from([Kind.Notify, 0x12, 0x34, 0x56, 0x78]),
  },
  NOTIFY_PAYLOAD: {
    btm: bt.binMessage(Kind.Notify, Encoding.Json, void 0, 0x12345678, void 0, '["hello","world"]'),
    btb: '2|2|305419896|["hello","world"]',
    ntm: nt.binMessage(Kind.Notify, Encoding.Json, void 0, 0x12345678, void 0, '["hello","world"]'),
    ntb: '2|2|305419896|["hello","world"]',
    nbm: nb.binMessage(Kind.Notify, Encoding.Json, void 0, 0x12345678, void 0, Buffer.from(
      [91, 34, 104, 101, 108, 108, 111, 34, 44, 34, 119, 111, 114, 108, 100, 34, 93],
    )),
    nbb: Buffer.from([
      Kind.Notify | Encoding.Json, 0x12, 0x34, 0x56, 0x78, 0x00, 0x00, 0x00, 0x11,
      91, 34, 104, 101, 108, 108, 111, 34, 44, 34, 119, 111, 114, 108, 100, 34, 93, // ["hello","world"]
    ]),
  },
  RESPONSE_NO_PAYLOAD: {
    btm: bt.binMessage(Kind.Response, void 0, 0x1234, void 0, Status.BadGateway),
    btb: '3|0|4660|50',
    ntm: nt.binMessage(Kind.Response, void 0, 0x1234, void 0, Status.BadGateway),
    ntb: '3|0|4660|50',
    nbm: nb.binMessage(Kind.Response, void 0, 0x1234, void 0, Status.BadGateway),
    nbb: Buffer.from([Kind.Response, 0x12, 0x34, 0x32]),
  },
  RESPONSE_PAYLOAD: {
    btm: bt.binMessage(Kind.Response, Encoding.Json, 0x1234, void 0, Status.BadGateway, '["hello","world"]'),
    btb: '3|2|4660|50|["hello","world"]',
    ntm: nt.binMessage(Kind.Response, Encoding.Json, 0x1234, void 0, Status.BadGateway, '["hello","world"]'),
    ntb: '3|2|4660|50|["hello","world"]',
    nbm: nb.binMessage(Kind.Response, Encoding.Json, 0x1234, void 0, Status.BadGateway, Buffer.from(
      [91, 34, 104, 101, 108, 108, 111, 34, 44, 34, 119, 111, 114, 108, 100, 34, 93],
    )),
    nbb: Buffer.from([
      Kind.Response | Encoding.Json, 0x12, 0x34, 0x32, 0x00, 0x00, 0x00, 0x11,
      91, 34, 104, 101, 108, 108, 111, 34, 44, 34, 119, 111, 114, 108, 100, 34, 93, // ["hello","world"]
    ]),
  },
}


describe('message marshal/unmarshal', () => {
  for (let name in cases) {
    const _case = cases[name]
    it(name + ': Browser.Texture.Stringify', () => {
      assert.deepEqual(bt.stringify(_case.btm), _case.btb)
    })
    const btb = _case.btb // TS cannot parse object property access context
    if (Array.isArray(btb)) {
      it(name + ': Browser.Texture.Parse.Split', () => {
        assert.equal(bt.parse(btb[0]), void 0)
        assert.deepEqual(bt.parse(btb[1]), _case.btm)
      })
    } else {
      it(name + ': Browser.Texture.Parse.Entire', () => {
        assert.deepEqual(bt.parse(btb), _case.btm)
      })
      it(name + ': Browser.Texture.Parse.Error', () => {
        assert.ifError(bt.getError())
      })
    }
    it(name + ': Node.Texture.Stringify', () => {
      assert.deepEqual(bt.stringify(_case.ntm), _case.ntb)
    })
    const ntb = _case.ntb // TS cannot parse object property access context
    if (Array.isArray(ntb)) {
      it(name + ': Node.Texture.Parse.Split', () => {
        assert.equal(nt.parse(ntb[0]), void 0)
        assert.deepEqual(nt.parse(ntb[1]), _case.ntm)
      })
    } else {
      it(name + ': Node.Texture.Parse.Entire', () => {
        assert.deepEqual(nt.parse(ntb), _case.ntm)
      })
    }
    it(name + ': Node.Binary.Stringify', () => {
      assert.deepEqual(nb.stringify(_case.nbm), _case.nbb)
    })
    it(name + ': Node.Binary.Parse.Entire', () => {
      assert.deepEqual(nb.parse(_case.nbb), [_case.nbm])
    })
    it(name + ': Node.Binary.Parse.Error', () => {
      assert.ifError(nb.getError())
    })
  }
})
