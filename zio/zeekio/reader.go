package zeekio

import (
	"bytes"
	"fmt"
	"io"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/pkg/skim"
)

const (
	ReadSize    = 64 * 1024
	MaxLineSize = 50 * 1024 * 1024
)

type Reader struct {
	scanner *skim.Scanner
	parser  *Parser
}

func NewReader(reader io.Reader, zctx *zed.Context) *Reader {
	buffer := make([]byte, ReadSize)
	return &Reader{
		scanner: skim.NewScanner(reader, buffer, MaxLineSize),
		parser:  NewParser(zctx),
	}
}

func (r *Reader) Read() (*zed.Value, error) {
	e := func(err error) error {
		if err == nil {
			return err
		}
		return fmt.Errorf("line %d: %w", r.scanner.Stats.Lines, err)
	}

again:
	line, err := r.scanner.ScanLine()
	if line == nil {
		if err != nil {
			return nil, e(err)
		}
		return nil, nil
	}
	// remove newline
	line = bytes.TrimSuffix(line, []byte("\n"))
	if line[0] == '#' {

		if err := r.parser.ParseDirective(line); err != nil {
			return nil, e(err)
		}
		goto again
	}
	rec, err := r.parser.ParseValue(line)
	if err != nil {
		return nil, e(err)
	}
	return rec, nil
}
