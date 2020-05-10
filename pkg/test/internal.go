package test

import (
	"context"
	"fmt"
	"strings"

	"github.com/brimsec/zq/driver"
	"github.com/brimsec/zq/emitter"
	"github.com/brimsec/zq/pkg/nano"
	"github.com/brimsec/zq/zbuf"
	"github.com/brimsec/zq/zio"
	"github.com/brimsec/zq/zio/detector"
	"github.com/brimsec/zq/zng/resolver"
	"github.com/brimsec/zq/zql"
	"go.uber.org/zap"
)

type Internal struct {
	Name         string
	Query        string
	Input        string
	InputFormat  string // defaults to "auto", like zq
	OutputFormat string // defaults to "tzng", like zq
	Expected     string
	ExpectedErr  error
}

func Trim(s string) string {
	return strings.TrimSpace(s) + "\n"
}

func stringReader(input string, ifmt string, zctx *resolver.Context) (zbuf.Reader, error) {
	if ifmt == "" {
		return detector.NewReader(strings.NewReader(input), zctx, "", detector.OpenConfig{})
	}
	zr, err := detector.LookupReader(strings.NewReader(input), zctx, ifmt)
	if err != nil {
		return nil, err
	}
	if zr == nil {
		return nil, fmt.Errorf("unknown input format %s", ifmt)
	}
	return zr, nil
}

func newEmitter(ofmt string) (*emitter.Bytes, error) {
	if ofmt == "" {
		ofmt = "tzng"
	}
	// XXX text format options not supported
	return emitter.NewBytes(&zio.WriterFlags{Format: ofmt})
}

func (i *Internal) Run() (string, error) {
	program, err := zql.ParseProc(i.Query)
	if err != nil {
		return "", fmt.Errorf("parse error: %s (%s)", err, i.Query)
	}
	reader, err := stringReader(i.Input, i.InputFormat, resolver.NewContext())
	if err != nil {
		return "", err
	}
	mux, err := driver.Compile(context.Background(), program, reader, false, nano.MaxSpan, zap.NewNop())
	if err != nil {
		return "", err
	}
	output, err := newEmitter(i.OutputFormat)
	if err != nil {
		return "", err
	}
	d := driver.NewCLI(output)
	if err := driver.Run(mux, d, nil); err != nil {
		return "", err
	}
	return string(output.Bytes()), nil
}
