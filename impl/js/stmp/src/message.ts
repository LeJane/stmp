/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-16 21:31:43
 * @version 1.0.0
 * @desc message.ts
 */

export enum Offset {
  Kind     = 6,
  Encoding = 3,
}

// use literals for avoid error TS2535
export enum Kind {
  Ping     = 0b00000000,
  Request  = 0b01000000,
  Notify   = 0b10000000,
  Response = 0b11000000,
}

export enum Encoding {
  None        = 0b000000,
  Protocobuf  = 0b001000,
  Json        = 0b010000,
  MessagePack = 0b011000,
  Bson        = 0b100000,
  Raw         = 0b101000,
}

export enum Flag {
  Kind     = 0b11000000,
  Encoding = 0b00111000,
}

export enum Status {
  Ok                    = 0x00,

  MovedPermanently      = 0x10,
  Found                 = 0x11,
  NotModified           = 0x12,

  BadRequest            = 0x20,
  Unauthorized          = 0x21,
  PaymentRequired       = 0x22,
  Forbidden             = 0x23,
  NotFound              = 0x24,
  RequestTimeout        = 0x25,
  RequestEntityTooLarge = 0x26,
  TooManyRequests       = 0x27,

  InternalServerError   = 0x30,
  NotImplemented        = 0x31,
  BadGateway            = 0x32,
  ServiceUnavailable    = 0x33,
  GatewayTimeout        = 0x34,
  VersionNotSupported   = 0x35,
}

export enum Protocol {
  None,
  Binary,
  Texture,
}

export const ErrorNotImplemented  = 'Not Implemented'
export const ErrorChunkMangled    = 'Chunk Mangled'
export const ErrorChunkEmpty      = 'Chunk Empty'
export const ErrorChunkInvalid    = 'Chunk Invalid'
export const ErrorIncorrectKind   = 'Incorrect Kind'
export const ErrorUnexpectedStage = 'Unexpected Stage'

export interface Message<T> {
  kind: Kind;
  encoding: Encoding | number;
  id: number;
  action: number;
  status: Status | number;
  // If encoding === Encoding.None or the network environment
  // does not need the PS field, this field should be 0.
  payloadSize: number;
  payload: T | void;
  data: any;
}

export const PingMessage: Message<any> = {
  kind: Kind.Ping,
  encoding: Encoding.None,
  id: 0,
  action: 0,
  status: Status.Ok,
  payloadSize: 0,
  payload: void 0,
  data: void 0,
}

export const PingTexture = '0'
