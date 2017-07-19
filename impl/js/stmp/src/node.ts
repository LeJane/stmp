/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-19 14:08:16
 * @version 1.0.0
 * @desc node.ts
 */

// Uint8Array => Buffer

import {Context} from './context'
import {Encoding, ErrorChunkMangled, ErrorUnexpectedStage, Flag, Kind, Message} from './message'
import {TextureContext} from './texture'

export type NodeBinaryMessage = Message<Buffer>
export type NodeTextureMessage = Message<Buffer | string>

export const NodePingBinary = Buffer.from([0])

enum BinaryStage {
  None = 0,
  Header,
  Id,
  Action,
  Status,
  Ps,
  Payload,
  Error,
}

export class NodeBinaryContext extends Context<Buffer> {
  private parsing: NodeBinaryMessage
  private stage: BinaryStage
  private chunks: Buffer[]
  private error?: Error
  // current using chunk index in chunks to get bytes
  private activeChunk: number
  // current parsing message consumed byteLength
  // start from active chunk
  private activeOffset: number
  // total byteLength of all the chunks
  private freeByteLength: number
  // the uint8array constructor
  protected readonly wps = true

  reset() {
    this.stage          = BinaryStage.None
    this.error          = void 0
    this.chunks         = []
    this.freeByteLength = 0
    this.activeChunk    = 0
    this.activeOffset   = 0
  }

  isError() {
    return this.stage === BinaryStage.Error
  }

  getError(): Error {
    return <Error>this.error
  }

  getChunks() {
    return this.chunks.slice()
  }

  payloadSize(data: Buffer | void): number {
    return data === void 0 ? 0 : data.byteLength
  }

  private next(size: number): Buffer | void {
    if (this.freeByteLength < size) {
      return
    }
    this.freeByteLength -= size
    if (this.chunks[this.activeChunk].byteLength > this.activeOffset + size) {
      // do not need to update the active chunk
      this.activeOffset += size
      return this.chunks[this.activeChunk].slice(this.activeOffset - size, this.activeOffset)
    }
    if (this.chunks[this.activeChunk].byteLength === this.activeOffset + size) {
      this.activeOffset = 0
      this.activeChunk++
      return this.chunks[this.activeChunk - 1].slice(-size)
    }
    // hack for node construct a Buffer
    const temp = new Buffer(size)
    temp.set(this.chunks[this.activeChunk].slice(this.activeOffset))
    size = this.chunks[this.activeChunk].byteLength - this.activeOffset
    this.activeChunk++
    for (; this.activeChunk < this.chunks.length; this.activeChunk++) {
      if (this.chunks[this.activeChunk].byteLength < size) {
        // is not enough, use next
        temp.set(this.chunks[this.activeChunk], temp.byteLength - size - this.chunks[this.activeChunk].byteLength)
        size -= this.chunks[this.activeChunk].byteLength // remain size
      } else {
        // if enough, finish
        break
      }
    }
    if (size === this.chunks[this.activeChunk].byteLength) {
      // full used
      temp.set(this.chunks[this.activeChunk], -size)
      this.activeChunk++
      this.activeOffset = 0
    } else {
      temp.set(this.chunks[this.activeChunk], temp.byteLength - size)
      this.activeOffset = size
    }
    return temp
  }

  /** clean used chunks */
  private resume() {
    this.stage = BinaryStage.None
    this.chunks.splice(0, this.activeChunk)
    this.activeChunk = 0
    if (this.activeOffset !== 0) {
      // must exist at least one chunk
      this.chunks[0]    = this.chunks[0].slice(this.activeOffset)
      this.activeOffset = 0
    }
  }

  private abort(reason: string) {
    this.error = new Error(reason)
    this.stage = BinaryStage.Error
  }

  parse(chunk: Buffer): NodeBinaryMessage[] | false {
    if (this.stage === BinaryStage.Error) {
      // If error already, will not consume the data
      return false
    }
    if (chunk && chunk.byteLength > 0) {
      this.chunks.push(chunk)
      this.freeByteLength += chunk.byteLength
    } else {
      return []
    }
    const output: NodeBinaryMessage[] = []
    if (this.freeByteLength === 0) {
      return output
    }
    while (true) {
      if (this.stage === BinaryStage.None || this.stage === BinaryStage.Header) {
        this.stage = BinaryStage.Header
        const buf  = this.next(1)
        if (buf === void 0) {
          return output
        }
        const flags = buf[0]
        if (flags !== 0 && flags < 0x40) {
          // the message is texture
          this.abort(ErrorChunkMangled)
          return false
        }
        this.parsing          = this.binMessage()
        this.parsing.kind     = flags & Flag.Kind
        this.parsing.encoding = flags & Flag.Encoding
        if (this.parsing.kind === Kind.Ping) {
          output.push(this.parsing)
          this.resume() // PingMessage END
          continue
        }
        if (this.parsing.kind === Kind.Request || this.parsing.kind === Kind.Response) {
          this.stage = BinaryStage.Id
        } else if (this.parsing.kind === Kind.Notify) {
          this.stage = BinaryStage.Action
        } // should no other cases occur
      }
      if (this.stage === BinaryStage.Id) {
        const buf = this.next(2)
        if (buf === void 0) {
          return output
        }
        this.parsing.id = (buf[0] << 8) | buf[1]
        if (this.parsing.kind === Kind.Request) {
          this.stage = BinaryStage.Action
        } else {
          this.stage = BinaryStage.Status
        }
      }
      if (this.stage === BinaryStage.Action) {
        const buf = this.next(4)
        if (buf === void 0) {
          return output
        }
        this.parsing.action = (buf[0] << 24)
                              | (buf[1] << 16)
                              | (buf[2] << 8)
                              | buf[3]
        if (this.parsing.encoding !== Encoding.None) {
          // MUST with PS field
          this.stage = BinaryStage.Ps
        } else {
          output.push(this.parsing)
          this.resume() // RequestMessage/NotifyMessage !WP END
          continue
        }
      }
      if (this.stage === BinaryStage.Status) {
        const buf = this.next(1)
        if (buf === void 0) {
          return output
        }
        this.parsing.status = buf[0]
        if (this.parsing.encoding !== Encoding.None) {
          this.stage = BinaryStage.Ps
        } else {
          output.push(this.parsing)
          this.resume() // ResponseMessage !WP END
          continue
        }
      }
      if (this.stage === BinaryStage.Ps) {
        const buf = this.next(4)
        if (buf === void 0) {
          return output
        }
        this.parsing.payloadSize = (buf[0] << 24)
                                   | (buf[1] << 16)
                                   | (buf[2] << 8)
                                   | buf[3]

        this.stage = BinaryStage.Payload
      }
      if (this.stage === BinaryStage.Payload) {
        if (this.parsing.payloadSize > 0) {
          const buf = this.next(this.parsing.payloadSize)
          if (buf === void 0) {
            return output
          }
          this.parsing.payload = buf
        }
        output.push(this.parsing)
        this.resume()
        continue
      }
      this.abort(ErrorUnexpectedStage)
      return false
    }
  }

  stringify(msg: NodeBinaryMessage): Buffer
  stringify(msg: NodeBinaryMessage, split: true): Buffer[]
  stringify(msg: NodeBinaryMessage, split = false): Buffer | Buffer[] {
    if (msg.kind === Kind.Ping) {
      return split === true ? [NodePingBinary] : NodePingBinary
    }
    let headerSize = 1
    if (msg.encoding !== Encoding.None) {
      headerSize += 4 // PS
    }
    switch (msg.kind) {
      case Kind.Request:
        headerSize += 6
        break
      case Kind.Notify:
        headerSize += 4
        break
      case Kind.Response:
        headerSize += 3
        break
    }
    const header = Buffer.allocUnsafe(
      msg.encoding !== Encoding.None && split !== true
        ? headerSize + msg.payloadSize
        : headerSize,
    )
    header[0]    = msg.kind | msg.encoding
    switch (msg.kind) {
      case Kind.Request:
        header[1] = msg.id >> 8
        header[2] = msg.id
        header[3] = msg.action >> 24
        header[4] = msg.action >> 16 // will be auto chunk
        header[5] = msg.action >> 8
        header[6] = msg.action
        break
      case Kind.Notify:
        header[1] = msg.action >> 24
        header[2] = msg.action >> 16
        header[3] = msg.action >> 8
        header[4] = msg.action
        break
      case Kind.Response:
        header[1] = msg.id >> 8
        header[2] = msg.id
        header[3] = msg.status
        break
    }
    if (msg.encoding !== Encoding.None) {
      header[headerSize - 4] = msg.payloadSize >> 24
      header[headerSize - 3] = msg.payloadSize >> 16
      header[headerSize - 2] = msg.payloadSize >> 8
      header[headerSize - 1] = msg.payloadSize
    }
    if (split !== true) {
      if (msg.encoding !== Encoding.None && msg.payload) {
        msg.payload.copy(header, headerSize)
      }
      return header
    } else if (msg.encoding !== Encoding.None && msg.payload) {
      return [header, msg.payload]
    } else {
      return [header]
    }
  }
}

export class NodeTextureContext extends TextureContext<Buffer> {
  payloadSize(payload: Buffer | string | void): number {
    return payload === void 0 ? 0 : typeof payload === 'string' ? payload.length : payload.byteLength
  }

  chunkSize(chunk: Buffer): number {
    return chunk ? chunk.byteLength : 0
  }
}
