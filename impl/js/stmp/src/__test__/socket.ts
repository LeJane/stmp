/*!
 *
 * Copyright 2017 - acrazing
 *
 * @author acrazing joking.young@gmail.com
 * @since 2017-07-16 22:14:55
 * @version 1.0.0
 * @desc socket.ts
 */


import {Server} from 'ws'

export function wsServer() {
  const server = new Server({port: 9999})
  server.addListener('connection', (client) => {
    client.on('message', (data) => {
      console.log(data instanceof Buffer ? data.toString() : data)
    })
  })
}
