package runtime

import (
	"context"
	"io"

	"github.com/brimdata/super"
	"github.com/brimdata/super/compiler/parser"
	"github.com/brimdata/super/lakeparse"
	"github.com/brimdata/super/zbuf"
	"github.com/brimdata/super/zio"
	"github.com/segmentio/ksuid"
)

type Compiler interface {
	NewQuery(*Context, *parser.AST, []zio.Reader) (Query, error)
	NewLakeQuery(*Context, *parser.AST, int, *lakeparse.Commitish) (Query, error)
	NewLakeDeleteQuery(*Context, *parser.AST, *lakeparse.Commitish) (DeleteQuery, error)
}

type Query interface {
	zbuf.Puller
	io.Closer
	Progress() zbuf.Progress
	Meter() zbuf.Meter
}

type DeleteQuery interface {
	Query
	DeletionSet() []ksuid.KSUID
}

func AsReader(q Query) zio.Reader {
	return zbuf.PullerReader(q)
}

func CompileQuery(ctx context.Context, zctx *super.Context, c Compiler, ast *parser.AST, readers []zio.Reader) (Query, error) {
	rctx := NewContext(ctx, zctx)
	q, err := c.NewQuery(rctx, ast, readers)
	if err != nil {
		rctx.Cancel()
		return nil, err
	}
	return q, nil
}

func CompileLakeQuery(ctx context.Context, zctx *super.Context, c Compiler, ast *parser.AST, head *lakeparse.Commitish) (Query, error) {
	rctx := NewContext(ctx, zctx)
	q, err := c.NewLakeQuery(rctx, ast, 0, head)
	if err != nil {
		rctx.Cancel()
		return nil, err
	}
	return q, nil
}
