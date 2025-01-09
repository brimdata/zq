package vng

import (
	"io"

	"github.com/brimdata/super/pkg/byteconv"
	"github.com/ronanh/intcomp"
	"golang.org/x/sync/errgroup"
)

type Uint32Encoder struct {
	vals     []uint32
	out      []byte
	bytesLen uint64
}

func (u *Uint32Encoder) Write(v uint32) {
	u.vals = append(u.vals, v)
}

func (u *Uint32Encoder) Encode(group *errgroup.Group) {
	group.Go(func() error {
		u.bytesLen = uint64(len(u.vals) * 4)
		compressed := intcomp.CompressUint32(u.vals, nil)
		u.out = byteconv.ReinterpretSlice[byte](compressed)
		return nil
	})
}

func (u *Uint32Encoder) Emit(w io.Writer) error {
	var err error
	if len(u.out) > 0 {
		_, err = w.Write(u.out)
	}
	return err
}

func (u *Uint32Encoder) Segment(off uint64) (uint64, Segment) {
	len := uint64(len(u.out))
	return off + len, Segment{
		Offset:            off,
		MemLength:         len,
		Length:            u.bytesLen,
		CompressionFormat: CompressionFormatNone,
	}
}

func ReadUint32s(loc Segment, r io.ReaderAt) ([]uint32, error) {
	buf := make([]byte, loc.MemLength)
	if err := loc.Read(r, buf); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	return intcomp.UncompressUint32(byteconv.ReinterpretSlice[uint32](buf), nil), nil
}
