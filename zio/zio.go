package zio

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/brimdata/zed/zio/textio"
	"github.com/brimdata/zed/zio/zngio"
	"github.com/brimdata/zed/zio/zsonio"
	"github.com/brimdata/zed/zio/zstio"
)

type ReaderOpts struct {
	Format string
	Zng    zngio.ReaderOpts
	AwsCfg *aws.Config
}

type WriterOpts struct {
	Format string
	UTF8   bool
	Text   textio.WriterOpts
	Zng    zngio.WriterOpts
	ZSON   zsonio.WriterOpts
	Zst    zstio.WriterOpts
}

func Extension(format string) string {
	switch format {
	case "tzng":
		return ".tzng"
	case "zeek":
		return ".log"
	case "ndjson":
		return ".ndjson"
	case "zjson":
		return ".ndjson"
	case "text":
		return ".txt"
	case "table":
		return ".tbl"
	case "zng":
		return ".zng"
	case "zson":
		return ".zson"
	case "csv":
		return ".csv"
	case "zst":
		return ".zst"
	default:
		return ""
	}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}
