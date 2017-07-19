/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-17 00:20:42
 * @version 1.0.0
 * @desc text.ts
 */


export function getUtf8ByteLength(input: string): number {
  let size  = 0
  const max = input.length
  let i     = 0
  let point: number
  // 😂
  for (; i < max;) {
    point = input.charCodeAt(i)
    i++
    if (point > 0xD7FF && point < 0xDC00) {
      point = ((point - 0xD800) << 10) + (input.charCodeAt(i) - 0xDC00) + 0x10000
      i++
    }
    if (point < 0x0080) {
      size += 1
    } else if (point < 0x0800) {
      size += 2
    } else if (point < 0x10000) {
      size += 3
    } else {
      size += 4
    }
  }
  return size
}

export function write(buffer: Uint8Array, input: string, offset = 0) {
  let point: number
  let i     = 0
  let char  = offset
  const max = input.length
  for (; i < max;) {
    point = input.charCodeAt(i)
    i++
    if (point > 0xD7FF && point < 0xDC00) {
      point = ((point - 0xD800) << 10) + (input.charCodeAt(i) - 0xDC00) + 0x10000
      i++
    }
    if (point < 0x0080) {
      buffer[char++] = point
    } else if (point < 0x0800) {
      buffer[char++] = 0b11000000 | (point >> 6)
      buffer[char++] = 0b10000000 | (point & 0b10111111)
    } else if (point < 0x10000) {
      buffer[char++] = 0b11100000 | (point >> 12)
      buffer[char++] = 0b10000000 | ((point >> 6) & 0b10111111)
      buffer[char++] = 0b10000000 | (point & 0b10111111)
    } else {
      buffer[char++] = 0b11110000 | (point >> 18)
      buffer[char++] = 0b10000000 | ((point >> 12) & 0b10111111)
      buffer[char++] = 0b10000000 | ((point >> 6) & 0b10111111)
      buffer[char++] = 0b10000000 | (point & 0b10111111)
    }
  }
}

export function read(buffer: Uint8Array, offset = 0): string {
  let output = ''
  let i      = offset
  const max  = buffer.byteLength
  let point
  for (; i < max;) {
    if (buffer[i] >> 7 == 0) {
      point = buffer[i]
      i += 1
    } else if (buffer[i] >> 5 == 0b110) {
      point = ((buffer[i] & 0b11111) << 6)
              | (buffer[i + 1] & 0b111111)
      i += 2
    } else if (buffer[i] >> 4 == 0b1110) {
      point = ((buffer[i] & 0b1111) << 12)
              | ((buffer[i + 1] & 0b111111) << 6)
              | (buffer[i + 2] & 0b111111)
      i += 3
    } else {
      point = ((buffer[i] & 0b111) << 18)
              | ((buffer[i + 1] & 0b111111) << 12)
              | ((buffer[i + 2] & 0b111111) << 6)
              | (buffer[i + 3] & 0b111111)
      i += 4
    }
    if (point > 0xFFFF) {
      point -= 0x10000
      output += String.fromCharCode((point >> 10) + 0xD800, (point & 0b1111111111) + 0xDC00)
    } else {
      output += String.fromCharCode(point)
    }
  }
  return output
}

export function encode(input: string, prefix = 0): Uint8Array {
  const size   = prefix + getUtf8ByteLength(input)
  const buffer = new Uint8Array(size)
  write(buffer, input, prefix)
  return buffer
}
