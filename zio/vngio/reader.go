package vngio

import (
	"errors"
	"io"

	"github.com/brimdata/super"
	"github.com/brimdata/super/pkg/field"
	"github.com/brimdata/super/vng"
	"github.com/brimdata/super/zio"
)

func NewReader(zctx *super.Context, r io.Reader, fields []field.Path) (zio.Reader, error) {
	ra, ok := r.(io.ReaderAt)
	if !ok {
		return nil, errors.New("Super Columnar requires a seekable input")
	}
	o, err := vng.NewObject(ra)
	if err != nil {
		return nil, err
	}
	return o.NewReader(zctx)
}
