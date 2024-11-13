package parquetio

import (
	"context"
	"errors"
	"io"

	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/brimdata/super"
	"github.com/brimdata/super/pkg/field"
	"github.com/brimdata/super/zio/arrowio"
)

func NewReader(zctx *super.Context, r io.Reader, fields []field.Path) (*arrowio.Reader, error) {
	ras, ok := r.(parquet.ReaderAtSeeker)
	if !ok {
		return nil, errors.New("reader cannot seek")
	}
	pr, err := file.NewParquetReader(ras)
	if err != nil {
		return nil, err
	}
	props := pqarrow.ArrowReadProperties{
		Parallel:  true,
		BatchSize: 256 * 1024,
	}
	fr, err := pqarrow.NewFileReader(pr, props, memory.DefaultAllocator)
	if err != nil {
		pr.Close()
		return nil, err
	}
	cols := columnIndexes(fr.Manifest, fields)
	rr, err := fr.GetRecordReader(context.TODO(), cols, nil)
	if err != nil {
		pr.Close()
		return nil, err
	}
	ar, err := arrowio.NewReaderFromRecordReader(zctx, rr)
	if err != nil {
		pr.Close()
		return nil, err
	}
	return ar, nil
}

func columnIndexes(manifest *pqarrow.SchemaManifest, fields []field.Path) []int {
	var indexes []int
	for _, f := range fields {
		for _, schemaField := range manifest.Fields {
			indexes = appendColumnIndexesForPath(indexes, schemaField, f)
		}
	}
	return indexes
}

func appendColumnIndexesForPath(indexes []int, sf pqarrow.SchemaField, path field.Path) []int {
	if len(path) == 0 || sf.Field.Name != path[0] {
		return indexes
	}
	if len(path) == 1 {
		return appendColumnIndexes(indexes, sf)
	}
	for _, c := range sf.Children {
		indexes = appendColumnIndexesForPath(indexes, c, path[1:])
	}
	return indexes
}

func appendColumnIndexes(indexes []int, sf pqarrow.SchemaField) []int {
	if len(sf.Children) == 0 {
		return append(indexes, sf.ColIndex)
	}
	for _, c := range sf.Children {
		indexes = appendColumnIndexes(indexes, c)
	}
	return indexes
}
