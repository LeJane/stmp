package stmp

import "io"

type Context struct {
	codecs   map[byte]Codec
	nextId   uint16
	wps      byte
	encoding byte
	texture  bool
	rd       io.Reader
}

func NewContext(wps byte, encoding byte, rd io.Reader, texture bool) *Context {
	return &Context{
		codecs: map[byte]Codec{},
		nextId: 0,
		wps: wps,
		encoding: encoding,
		texture: texture,
		rd: rd,
	}
}

func (c *Context) Clone(rd io.Reader) *Context {
	return &Context{
		codecs: c.codecs,
		nextId: 0,
		wps: c.wps,
		encoding: c.encoding,
		texture: c.texture,
		rd: rd,
	}
}

func (c *Context) Read() (*Message, error) {
	return Read(c.rd)
}

func (c *Context) Parse(data []byte) (*Message, error) {
	return Parse(data)
}
