/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-19 14:08:16
 * @version 1.0.0
 * @desc texture.ts
 */


import {
  Encoding,
  ErrorChunkEmpty, ErrorChunkInvalid, ErrorChunkMangled, ErrorIncorrectKind, Kind, Message,
  Offset, PingTexture,
} from './message'
import {Context} from './context'

export type TextureMessage<T> = Message<T | string>

export abstract class TextureContext<T> extends Context<T | string> {
  private parsing: TextureMessage<T> | void
  private header: string | void
  private payload: T | string | void
  private error: Error | void
  private offset: number // the offset of current index
  protected readonly wps = false

  /** reset to initial state */
  reset() {
    this.parsing = void 0
    this.header  = void 0
    this.payload = void 0
    this.error   = void 0
    this.offset  = 0
  }

  isError() {
    return this.error !== void 0
  }

  isParsing() {
    return this.parsing !== void 0 && this.error === void 0
  }

  getError() {
    return this.error
  }

  getChunks(): [string | void, T | string | void] {
    return [this.header, this.payload]
  }

  private abort(reason: string): false {
    this.error = new Error(reason)
    return false
  }

  private field(): string | void {
    if (this.header === void 0) {
      return void 0
    }
    if (this.offset === this.header.length) {
      return void 0
    }
    const offset = this.header.indexOf('|', this.offset)
    if (offset === -1) {
      const ret   = this.header.substr(this.offset)
      this.offset = this.header.length
      return ret
    } else {
      const ret   = this.header.substring(this.offset, offset)
      this.offset = offset + 1
      return ret
    }
  }

  abstract chunkSize(chunk: T): number


  /**
   * @param chunk - the chunk to parse
   * @return  void: parsing, need another chunk
   *          NodeTextureMessage: parse finished
   *          false: error occurred
   */
  parse(chunk: string | T): TextureMessage<T> | void | false {
    if (this.isError()) {
      return false
    }
    if (!chunk || (typeof chunk === 'string' ? chunk.length : this.chunkSize(chunk)) === 0) {
      return this.abort(ErrorChunkEmpty)
    }
    if (!this.parsing) {
      this.header = <string>chunk
      // header must be transported in string
      if (typeof chunk !== 'string') {
        return this.abort(ErrorChunkMangled)
      }
      const parsing = this.parsing = this.binMessage()
      let field = this.field()
      if (field === void 0) {
        return this.abort(ErrorChunkInvalid)
      }
      // KIND
      const kind: Kind = (field.charCodeAt(0) - 0x30) << Offset.Kind
      if (kind !== Kind.Ping && kind !== Kind.Request && kind !== Kind.Notify && kind !== Kind.Response) {
        // 2 bit kind
        return this.abort(ErrorIncorrectKind)
      }
      parsing.kind = kind
      if (this.parsing.kind === Kind.Ping) {
        // PingMessage
        if (this.field() !== void 0) {
          this.abort(ErrorChunkInvalid)
          return void 0
        }
        this.reset()
        return parsing
      }
      // ENCODING
      field = this.field()
      if (field === void 0) {
        return this.abort(ErrorChunkInvalid)
      }
      parsing.encoding = (field.charCodeAt(0) - 0x30) << Offset.Encoding
      // ID
      if (parsing.kind === Kind.Request || parsing.kind === Kind.Response) {
        field = this.field()
        if (field === void 0) {
          return this.abort(ErrorChunkInvalid)
        }
        parsing.id = parseInt(field, 10)
      }
      // ACTION
      if (parsing.kind === Kind.Request || parsing.kind === Kind.Notify) {
        field = this.field()
        if (field === void 0) {
          return this.abort(ErrorChunkInvalid)
        }
        parsing.action = parseInt(field, 10)
      }
      // STATUS
      if (parsing.kind === Kind.Response) {
        field = this.field()
        if (field === void 0) {
          return this.abort(ErrorChunkInvalid)
        }
        parsing.status = parseInt(field, 10)
      }
      // SHOULD NOT contains PS field
      // PAYLOAD
      field = this.field()
      if (parsing.encoding !== Encoding.None) {
        if (field === void 0) {
          return this.abort(ErrorChunkInvalid)
        } else {
          parsing.payload = field
          this.reset()
          return parsing
        }
      } else if (field !== void 0) {
        return this.abort(ErrorChunkInvalid)
      } else {
        parsing.payload = field
        this.reset()
        return parsing
      }
    } else {
      const parsing   = this.parsing
      parsing.payload = chunk
      this.reset()
      return parsing
    }
  }

  stringify(msg: TextureMessage<T>): string | [string, T] {
    if (msg.kind === Kind.Ping) {
      return PingTexture
    }
    const headers: any[] = [msg.kind >> Offset.Kind, msg.encoding >> Offset.Encoding]
    switch (msg.kind) {
      case Kind.Request:
        headers.push(msg.id, msg.action)
        break
      case Kind.Notify:
        headers.push(msg.action)
        break
      case Kind.Response:
        headers.push(msg.id, msg.status)
        break
    }
    if (msg.encoding !== Encoding.None && typeof msg.payload !== 'string' && msg.payload !== void 0) {
      return [headers.join('|'), msg.payload]
    } else if (msg.encoding !== Encoding.None) {
      headers.push(msg.payload)
    }
    return headers.join('|')
  }
}
