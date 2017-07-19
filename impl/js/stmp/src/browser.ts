/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-19 19:32:23
 * @version 1.0.0
 * @desc browser.ts
 */


import {TextureContext} from './texture'
import {Message} from './message'

export const BrowserPingBinary = Uint8Array.of(0)

export type BrowserTextureMessage = Message<Uint8Array | string>

export class BrowserTextureContext extends TextureContext<Uint8Array> {
  payloadSize(payload: Uint8Array | string | void): number {
    return payload === void 0 ? 0 : typeof payload === 'string' ? payload.length : payload.byteLength
  }

  chunkSize(chunk: Uint8Array): number {
    return chunk ? chunk.byteLength : 0
  }
}
