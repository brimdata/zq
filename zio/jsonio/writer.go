package jsonio

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/brimdata/super"
	"github.com/brimdata/super/pkg/nano"
	"github.com/brimdata/super/pkg/terminal/color"
	"github.com/brimdata/super/zcode"
	"github.com/brimdata/super/zson"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"time"
)

var (
	boolColor   = []byte("\x1b[1m")
	fieldColor  = []byte("\x1b[34;1m")
	nullColor   = []byte("\x1b[2m")
	numberColor = []byte("\x1b[36m")
	puncColor   = []byte{} // no color
	stringColor = []byte("\x1b[32m")
)

type Writer struct {
	io.Closer
	writer     *bufio.Writer
	writer0    io.WriteCloser
	byteWriter *ByteArrayWriter
	tab        int
	isYaml     bool

	// Use json.Encoder for primitive Values. Have to use
	// json.Encoder instead of json.Marshal because it's
	// the only way to turn off HTML escaping.
	primEnc *json.Encoder
	primBuf bytes.Buffer
}
type ByteArrayWriter struct {
	data []byte // 用于保存写入的字节数据
}

func NewByteArrayWriter() *ByteArrayWriter {
	return &ByteArrayWriter{}
}

// Write 实现 io.Writer 接口，将数据写入 data
func (w *ByteArrayWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

// Close 实现 io.Closer 接口，关闭数据流
func (w *ByteArrayWriter) Close() error {
	// 对于简单的内存写入，这里不需要做复杂的清理操作
	// 但如果是文件流等，通常会做清理工作
	// 这里我们可以清空数据或者做其他的关闭操作
	return nil
}

// GetData 获取写入的数据
func (w *ByteArrayWriter) GetData() []byte {
	return w.data
}

type WriterOpts struct {
	Pretty int
	IsYaml bool
}

func NewWriter(writer io.WriteCloser, opts WriterOpts) *Writer {
	b := NewByteArrayWriter()
	w := &Writer{
		Closer:     writer,
		writer:     bufio.NewWriter(b),
		byteWriter: b,
		writer0:    writer,
		tab:        opts.Pretty,
		isYaml:     opts.IsYaml,
	}
	w.primEnc = json.NewEncoder(&w.primBuf)
	w.primEnc.SetEscapeHTML(false)
	return w
}
func JsonToYaml(jsonData []byte) []byte {
	// 1. 解析 JSON 数据到 Go 数据结构
	var jsonObj interface{}
	err := json.Unmarshal(jsonData, &jsonObj)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	// 2. 将 Go 数据结构转换为 YAML 格式
	yamlData, err := yaml.Marshal(jsonObj)
	if err != nil {
		log.Fatal("Error marshalling to YAML:", err)
	}
	return yamlData
}

func (w *Writer) Write(val super.Value) error {
	// writeAny doesn't return an error because any error that occurs will be
	// surfaced with w.writer.Flush is called.
	w.writeAny(0, val)
	w.writer.WriteByte('\n')
	w.writer.Flush()
	data := w.byteWriter.GetData()
	var err error
	if w.isYaml {
		_, err = w.writer0.Write(JsonToYaml(data))
	} else {
		_, err = w.writer0.Write(data)
	}
	return err
}

func (w *Writer) writeAny(tab int, val super.Value) {
	val = val.Under()
	if val.IsNull() {
		w.writeColor([]byte("null"), nullColor)
		return
	}
	if val.Type().ID() < super.IDTypeComplex {
		w.writePrimitive(val)
		return
	}
	switch typ := val.Type().(type) {
	case *super.TypeRecord:
		w.writeRecord(tab, typ, val.Bytes())
	case *super.TypeArray:
		w.writeArray(tab, typ.Type, val.Bytes())
	case *super.TypeSet:
		w.writeArray(tab, typ.Type, val.Bytes())
	case *super.TypeMap:
		w.writeMap(tab, typ, val.Bytes())
	case *super.TypeEnum:
		w.writeEnum(typ, val.Bytes())
	case *super.TypeError:
		w.writeError(tab, typ, val.Bytes())
	default:
		panic(fmt.Sprintf("unsupported type: %s", zson.FormatType(typ)))
	}
}

func (w *Writer) writeRecord(tab int, typ *super.TypeRecord, bytes zcode.Bytes) {
	tab += w.tab
	w.punc('{')
	if len(bytes) == 0 {
		w.punc('}')
		return
	}
	it := bytes.Iter()
	for i, f := range typ.Fields {
		if i != 0 {
			w.punc(',')
		}
		w.writeEntry(tab, f.Name, super.NewValue(f.Type, it.Next()))
	}
	w.newline()
	w.indent(tab - w.tab)
	w.punc('}')
}

func (w *Writer) writeArray(tab int, typ super.Type, bytes zcode.Bytes) {
	tab += w.tab
	w.punc('[')
	if len(bytes) == 0 {
		w.punc(']')
		return
	}
	it := bytes.Iter()
	for i := 0; !it.Done(); i++ {
		if i != 0 {
			w.punc(',')
		}
		w.newline()
		w.indent(tab)
		w.writeAny(tab, super.NewValue(typ, it.Next()))
	}
	w.newline()
	w.indent(tab - w.tab)
	w.punc(']')
}

func (w *Writer) writeMap(tab int, typ *super.TypeMap, bytes zcode.Bytes) {
	tab += w.tab
	w.punc('{')
	if len(bytes) == 0 {
		w.punc('}')
		return
	}
	it := bytes.Iter()
	for i := 0; !it.Done(); i++ {
		if i != 0 {
			w.punc(',')
		}
		key := mapKey(typ.KeyType, it.Next())
		w.writeEntry(tab, key, super.NewValue(typ.ValType, it.Next()))
	}
	w.newline()
	w.indent(tab - w.tab)
	w.punc('}')
}

func mapKey(typ super.Type, b zcode.Bytes) string {
	val := super.NewValue(typ, b)
	switch val.Type().Kind() {
	case super.PrimitiveKind:
		if val.Type().ID() == super.IDString {
			// Don't quote strings.
			return val.AsString()
		}
		return zson.FormatPrimitive(val.Type(), val.Bytes())
	case super.UnionKind:
		// Untagged, decorated ZSON so
		// |{0:1,0(uint64):2,0(=t):3,"0":4}| gets unique keys.
		typ, bytes := typ.(*super.TypeUnion).Untag(b)
		return zson.FormatValue(super.NewValue(typ, bytes))
	case super.EnumKind:
		return convertEnum(typ.(*super.TypeEnum), b)
	default:
		return zson.FormatValue(val)
	}
}

func (w *Writer) writeEnum(typ *super.TypeEnum, bytes zcode.Bytes) {
	w.writeColor(w.marshalJSON(convertEnum(typ, bytes)), stringColor)
}

func convertEnum(typ *super.TypeEnum, bytes zcode.Bytes) string {
	if k := int(super.DecodeUint(bytes)); k < len(typ.Symbols) {
		return typ.Symbols[k]
	}
	return "<bad enum>"
}

func (w *Writer) writeError(tab int, typ *super.TypeError, bytes zcode.Bytes) {
	tab += w.tab
	w.punc('{')
	w.writeEntry(tab, "error", super.NewValue(typ.Type, bytes))
	w.newline()
	w.indent(tab - w.tab)
	w.punc('}')
}

func (w *Writer) writeEntry(tab int, name string, val super.Value) {
	w.newline()
	w.indent(tab)
	w.writeColor(w.marshalJSON(name), fieldColor)
	w.punc(':')
	if w.tab != 0 {
		w.writer.WriteByte(' ')
	}
	w.writeAny(tab, val)
}

func (w *Writer) writePrimitive(val super.Value) {
	var v any
	c := stringColor
	switch id := val.Type().ID(); {
	case id == super.IDDuration:
		v = nano.Duration(val.Int()).String()
	case id == super.IDTime:
		v = nano.Ts(val.Int()).Time().Format(time.RFC3339Nano)
	case super.IsSigned(id):
		v, c = val.Int(), numberColor
	case super.IsUnsigned(id):
		v, c = val.Uint(), numberColor
	case super.IsFloat(id):
		v, c = val.Float(), numberColor
	case id == super.IDBool:
		v, c = val.AsBool(), boolColor
	case id == super.IDBytes:
		v = "0x" + hex.EncodeToString(val.Bytes())
	case id == super.IDString:
		v = val.AsString()
	case id == super.IDIP:
		v = super.DecodeIP(val.Bytes()).String()
	case id == super.IDNet:
		v = super.DecodeNet(val.Bytes()).String()
	case id == super.IDType:
		v = zson.FormatValue(val)
	default:
		panic(fmt.Sprintf("unsupported id=%d", id))
	}
	w.writeColor(w.marshalJSON(v), c)
}

func (w *Writer) marshalJSON(v any) []byte {
	w.primBuf.Reset()
	if err := w.primEnc.Encode(v); err != nil {
		panic(err)
	}
	return bytes.TrimSpace(w.primBuf.Bytes())
}

func (w *Writer) punc(b byte) {
	w.writeColor([]byte{b}, puncColor)
}

func (w *Writer) writeColor(b []byte, code []byte) {
	if color.Enabled {
		w.writer.Write(code)
		defer w.writer.WriteString(color.Reset.String())
	}
	w.writer.Write(b)
}

func (w *Writer) newline() {
	if w.tab > 0 {
		w.writer.WriteByte('\n')
	}
}

func (w *Writer) indent(tab int) {
	w.writer.Write(bytes.Repeat([]byte(" "), tab))
}
