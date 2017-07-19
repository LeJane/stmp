/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-19 18:30:37
 * @version 1.0.0
 * @desc context.ts
 */

import {Encoding, Kind, Message, Status} from './message'

export interface ContextOptions {
}

export abstract class Context<T> {
  protected options: ContextOptions
  protected readonly wps: boolean
  private id: number

  constructor(options: ContextOptions) {
    this.options = options
    this.id      = 0
    this.reset()
  }

  abstract reset(): void

  nextId() {
    return (this.id++) & 0xFFFF
  }

  abstract payloadSize(payload: T | void): number

  binMessage(
    kind              = Kind.Ping,
    encoding          = Encoding.None,
    id                = kind === Kind.Request ? this.nextId() : 0,
    action            = 0,
    status            = Status.Ok,
    payload: T | void = void 0,
    payloadSize       = this.payloadSize(payload),
  ): Message<T> {
    return {
      kind,
      encoding: payloadSize === 0 ? Encoding.None : encoding,
      id,
      action,
      status,
      payloadSize: this.wps ? payloadSize : 0,
      // TODO distinguish WP, WPS, PAYLOAD field
      payload: payloadSize === 0 ? void 0 : payload,
      data: void 0,
    }
  }
}
