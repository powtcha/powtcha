package byteparser

import (
	"errors"
	"powtcha/util"
)

var (
	errBufferEmpty = errors.New("not enough bytes left in buffer")
)

type ParserContinuation func(child *Parser) error

type Parser struct {
	buf []byte
	head uint
}

func NewParser(buf []byte) *Parser {
	return &Parser{
		buf:  buf,
		head: 0,
	}
}

func (p *Parser) Remaining() uint {
	return uint(len(p.buf)) - p.head
}

func (p *Parser) GetUint8() (uint8, error)  {
	if p.Remaining() < 1 {
		return 0, errBufferEmpty
	}
	res := p.buf[p.head]
	p.head++
	return res, nil
}

func (p *Parser) GetUint32() (uint32, error)  {
	if p.Remaining() < 4 {
		return 0, errBufferEmpty
	}
	res := util.BytesToUint32(p.buf[p.head:p.head+4])
	p.head += 4
	return res, nil
}

func (p *Parser) GetBytes(n uint) ([]byte, error)  {
	if p.Remaining() < n {
		return nil, errBufferEmpty
	}
	res := make([]byte, n)
	copy(res, p.buf[p.head:])
	p.head += n
	return res, nil
}

func (p *Parser) GetUint8LengthPrefixed(f ParserContinuation) error {
	length, err := p.GetUint8()
	if err != nil {
		return err
	}
	if p.Remaining() < uint(length) {
		return errBufferEmpty
	}
	buf, err := p.GetBytes(uint(length))
	if err != nil {
		return err
	}
	return f(&Parser{
		buf:  buf,
		head: 0,
	})
}
