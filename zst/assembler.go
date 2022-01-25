package zst

import (
	"errors"
	"io"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/pkg/storage"
	"github.com/brimdata/zed/zcode"
	"github.com/brimdata/zed/zio"
	"github.com/brimdata/zed/zst/column"
)

var ErrBadTypeNumber = errors.New("ZST: bad type number in root reassembly map")

// Assembler reads a columnar ZST object to generate a stream of zed.Values.
// It also has methods to read metadata for test and debugging.
type Assembler struct {
	root    *column.IntReader
	readers []column.Reader
	types   []zed.Type
	builder zcode.Builder
	err     error
}

var _ zio.Reader = (*Assembler)(nil)

type Assembly struct {
	root  zed.Value
	types []zed.Type
	maps  []*zed.Value
}

func NewAssembler(a *Assembly, seeker *storage.Seeker) (*Assembler, error) {
	root, err := column.NewIntReader(a.root, seeker)
	if err != nil {
		return nil, err
	}
	var readers []column.Reader
	for k := range a.types {
		val := a.maps[k]
		reader, err := column.NewReader(a.types[k], *val, seeker)
		if err != nil {
			return nil, err
		}
		readers = append(readers, reader)
	}
	return &Assembler{
		root:    root,
		types:   a.types,
		readers: readers,
	}, nil
}

func (a *Assembler) Read() (*zed.Value, error) {
	a.builder.Reset()
	typeNo, err := a.root.Read()
	if err == io.EOF {
		return nil, nil
	}
	if typeNo < 0 || int(typeNo) >= len(a.readers) {
		return nil, ErrBadTypeNumber
	}
	reader := a.readers[typeNo]
	if reader == nil {
		return nil, ErrBadTypeNumber
	}
	err = reader.Read(&a.builder)
	if err != nil {
		return nil, err
	}
	body, err := a.builder.Bytes().Body()
	if err != nil {
		return nil, err
	}
	rec := zed.NewValue(a.types[typeNo], body)
	//XXX if we had a buffer pool where records could be built back to
	// back in batches, then we could get rid of this extra allocation
	// and copy on every record
	return rec.Copy(), nil
}
