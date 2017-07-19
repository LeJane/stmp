/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-17 00:46:07
 * @version 1.0.0
 * @desc text.spec.ts
 */
import assert = require('assert')
import {encode, getUtf8ByteLength, read} from './text'

const cases = [
  'a',
  '\u00FF',
  '中',
  '😂',
]

cases.reduce((str, char) => {
  str += char
  assert.equal(getUtf8ByteLength(str), Buffer.from(str).byteLength, 'length: ' + str)
  assert.deepEqual(encode(str), Buffer.from(str), 'encoding: ' + str)
  assert.equal(str, read(Buffer.from(str)), 'reading: ' + str)
  return str
}, '')
