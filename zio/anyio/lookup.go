package anyio

import (
	"bytes"
	"fmt"
	"io"

	"encoding/json"
	"github.com/brimdata/super"
	"github.com/brimdata/super/zio"
	"github.com/brimdata/super/zio/arrowio"
	"github.com/brimdata/super/zio/csvio"
	"github.com/brimdata/super/zio/jsonio"
	"github.com/brimdata/super/zio/lineio"
	"github.com/brimdata/super/zio/parquetio"
	"github.com/brimdata/super/zio/vngio"
	"github.com/brimdata/super/zio/zeekio"
	"github.com/brimdata/super/zio/zjsonio"
	"github.com/brimdata/super/zio/zngio"
	"github.com/brimdata/super/zio/zsonio"
	"gopkg.in/yaml.v3"
)

func lookupReader(zctx *super.Context, r io.Reader, opts ReaderOpts) (zio.ReadCloser, error) {
	switch opts.Format {
	case "arrows":
		return arrowio.NewReader(zctx, r)
	case "bsup":
		return zngio.NewReaderWithOpts(zctx, r, opts.ZNG), nil
	case "csup":
		zr, err := vngio.NewReader(zctx, r, opts.Fields)
		if err != nil {
			return nil, err
		}
		return zio.NopReadCloser(zr), nil
	case "csv":
		return zio.NopReadCloser(csvio.NewReader(zctx, r, opts.CSV)), nil
	case "jsup":
		return zio.NopReadCloser(zsonio.NewReader(zctx, r)), nil
	case "line":
		return zio.NopReadCloser(lineio.NewReader(r)), nil
	case "json":
		return zio.NopReadCloser(jsonio.NewReader(zctx, r)), nil
	case "yaml":
		r1, err := yamlToJSONReader(r)
		if err != nil {
			return nil, err
		}
		return zio.NopReadCloser(jsonio.NewReader(zctx, r1)), nil
	case "parquet":
		zr, err := parquetio.NewReader(zctx, r, opts.Fields)
		if err != nil {
			return nil, err
		}
		return zio.NopReadCloser(zr), nil
	case "tsv":
		opts.CSV.Delim = '\t'
		return zio.NopReadCloser(csvio.NewReader(zctx, r, opts.CSV)), nil
	case "zeek":
		return zio.NopReadCloser(zeekio.NewReader(zctx, r)), nil
	case "zjson":
		return zio.NopReadCloser(zjsonio.NewReader(zctx, r)), nil
	}
	return nil, fmt.Errorf("no such format: \"%s\"", opts.Format)
}

// 将 io.Reader 中的 YAML 数据转换为 JSON，并返回一个新的 io.Reader
func yamlToJSONReader(r io.Reader) (io.Reader, error) {
	// 读取 YAML 数据
	yamlData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read yaml data: %v", err)
	}

	// 用于保存解析后的数据
	var data interface{}

	// 解析 YAML 数据
	err = yaml.Unmarshal(yamlData, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}

	// 将解析后的数据转为 JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %v", err)
	}

	// 返回一个新的 io.Reader（bytes.Buffer）
	return bytes.NewReader(jsonData), nil
}
