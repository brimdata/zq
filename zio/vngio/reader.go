package vngio

import (
	"errors"
	"io"
	"math"

	"github.com/brimdata/super"
	"github.com/brimdata/super/pkg/field"
	"github.com/brimdata/super/vng"
	"github.com/brimdata/super/zio"
)

type reader struct {
	zctx     *super.Context
	objects  []*vng.Object
	n        int
	readerAt io.ReaderAt
	reader   zio.Reader
}

func NewReader(zctx *super.Context, r io.Reader, fields []field.Path) (zio.Reader, error) {
	ra, ok := r.(io.ReaderAt)
	if !ok {
		return nil, errors.New("Super Columnar requires a seekable input")
	}
	objects, err := readObjects(ra)
	if err != nil {
		return nil, err
	}
	return &reader{
		zctx:     zctx,
		objects:  objects,
		readerAt: ra,
	}, nil
}

func (r *reader) Read() (*super.Value, error) {
again:
	if r.reader == nil {
		if r.n >= len(r.objects) {
			return nil, nil
		}
		o := r.objects[r.n]
		r.n++
		var err error
		if r.reader, err = o.NewReader(r.zctx); err != nil {
			return nil, err
		}
	}
	v, err := r.reader.Read()
	if v == nil && err == nil {
		r.reader = nil
		goto again
	}
	return v, err
}

func (r *reader) Close() error {
	if closer, ok := r.readerAt.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func readObjects(r io.ReaderAt) ([]*vng.Object, error) {
	var objects []*vng.Object
	var start int64
	for {
		// NewObject puts the right end to the passed in SectionReader and since
		// the end is unkown just pass MaxInt64.
		o, err := vng.NewObject(io.NewSectionReader(r, start, math.MaxInt64))
		if err != nil {
			if err == io.EOF && len(objects) > 0 {
				return objects, nil
			}
			return nil, err
		}
		objects = append(objects, o)
		start += int64(o.Size())
	}
}
