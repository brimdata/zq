package vng

import (
	"io"

	"github.com/brimdata/super"
	"github.com/brimdata/super/zcode"
)

type Int64Encoder struct {
	PrimitiveEncoder
}

func NewInt64Encoder() *Int64Encoder {
	return &Int64Encoder{*NewPrimitiveEncoder(super.TypeInt64, false)}
}

func (p *Int64Encoder) Write(v int64) {
	p.PrimitiveEncoder.Write(super.EncodeInt(v))
}

func ReadUint32s(loc Segment, r io.ReaderAt) ([]uint32, error) {
	buf := make([]byte, loc.MemLength)
	if err := loc.Read(r, buf); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	var vals []uint32
	for it := zcode.Iter(buf); !it.Done(); {
		vals = append(vals, uint32(super.DecodeInt(it.Next())))
	}
	return vals, nil
}
