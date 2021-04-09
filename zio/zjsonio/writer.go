package zjsonio

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/brimdata/zed/compiler/ast"
	"github.com/brimdata/zed/zng"
	"github.com/brimdata/zed/zson"
)

type Object struct {
	Schema string        `json:"schema"`
	Types  []ast.Type    `json:"types,omitempty"`
	Values []interface{} `json:"values"`
}

func unmarshal(b []byte) (*Object, error) {
	var template struct {
		Schema string        `json:"schema"`
		Types  []interface{} `json:"types,omitempty"`
		Values []interface{} `json:"values"`
	}
	if err := json.Unmarshal(b, &template); err != nil {
		return nil, err
	}
	var types []ast.Type
	for _, t := range template.Types {
		object, err := unpacker.UnpackMap(t)
		if object == nil || err != nil {
			return nil, err
		}
		typ, ok := object.(ast.Type)
		if !ok {
			return nil, fmt.Errorf("zjson: types object is not a type: %s", string(b))
		}
		types = append(types, typ)
	}
	return &Object{
		Schema: template.Schema,
		Types:  types,
		Values: template.Values,
	}, nil
}

type Writer struct {
	writer io.WriteCloser
	stream *Stream
}

func NewWriter(w io.WriteCloser, zctx *zson.Context) *Writer {
	return &Writer{
		writer: w,
		stream: NewStream(zctx),
	}
}

func (w *Writer) Close() error {
	return w.writer.Close()
}

func (w *Writer) Write(r *zng.Record) error {
	rec, err := w.stream.Transform(r)
	if err != nil {
		return err
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(b)
	if err != nil {
		return err
	}
	return w.write("\n")
}

func (w *Writer) write(s string) error {
	_, err := w.writer.Write([]byte(s))
	return err
}
