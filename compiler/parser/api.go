package parser

import (
	"errors"

	"github.com/brimdata/super/compiler/ast"
	"github.com/brimdata/super/compiler/srcfiles"
)

type AST struct {
	seq   ast.Seq
	files *srcfiles.List
}

func (a *AST) Parsed() ast.Seq {
	return a.seq
}

func (a *AST) Copy() ast.Seq {
	return ast.CopySeq(a.seq)
}

func (a *AST) Files() *srcfiles.List {
	return a.files
}

func (a *AST) ConvertToDeleteWhere() error {
	if len(a.seq) == 0 {
		return errors.New("internal error: AST seq cannot be empty")
	}
	a.seq.Prepend(&ast.Delete{Kind: "Delete"})
	return nil
}

// ParseQuery parses a query text and an optional set of include files and
// tracks include file names and line numbers for error reporting.
func ParseQuery(query string, filenames ...string) (*AST, error) {
	files, err := srcfiles.Concat(filenames, query)
	if err != nil {
		return nil, err
	}
	p, err := Parse("", []byte(files.Text), Recover(false))
	if err != nil {
		if err := convertParseErrs(err, files); err != nil {
			return nil, err
		}
		return nil, files.Error()
	}
	return &AST{sliceOf[ast.Op](p), files}, nil
}

func convertParseErrs(err error, files *srcfiles.List) error {
	errs, ok := err.(errList)
	if !ok {
		return err
	}
	for _, e := range errs {
		pe, ok := e.(*parserError)
		if !ok {
			return err
		}
		files.AddError("parse error", pe.pos.offset, -1)
	}
	return nil
}
