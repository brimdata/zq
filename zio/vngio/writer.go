package vngio

import (
	"io"

	"github.com/brimdata/zed/vng"
)

// NewWriter returns a writer to w with reasonable default options.
func NewWriter(w io.WriteCloser) (*vng.Writer, error) {
	return vng.NewWriter(w)
}
