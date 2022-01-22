package zed

import (
	"fmt"

	"github.com/brimdata/zed/zcode"
)

type TypeRecord struct {
	id      int
	Columns []Column
	LUT     map[string]int
}

func NewTypeRecord(id int, columns []Column) *TypeRecord {
	if columns == nil {
		columns = []Column{}
	}
	r := &TypeRecord{
		id:      id,
		Columns: columns,
	}
	r.createLUT()
	return r
}

func (t *TypeRecord) ID() int {
	return t.id
}

//XXX we shouldn't need this... tests are using it
func (t *TypeRecord) Decode(zv zcode.Bytes) ([]Value, error) {
	if zv == nil {
		return nil, nil
	}
	var vals []Value
	for i, it := 0, zv.Iter(); !it.Done(); i++ {
		val := it.Next()
		if i >= len(t.Columns) {
			return nil, fmt.Errorf("too many values for record element %s", val)
		}
		v := Value{t.Columns[i].Type, val}
		vals = append(vals, v)
	}
	return vals, nil
}

func (t *TypeRecord) Marshal(zv zcode.Bytes) interface{} {
	m := make(map[string]*Value)
	it := zv.Iter()
	for _, col := range t.Columns {
		m[col.Name] = &Value{col.Type, it.Next()}
	}
	return m
}

func (t *TypeRecord) ColumnOfField(field string) (int, bool) {
	v, ok := t.LUT[field]
	return v, ok
}

func (t *TypeRecord) TypeOfField(field string) (Type, bool) {
	n, ok := t.LUT[field]
	if !ok {
		return nil, false
	}
	return t.Columns[n].Type, true
}

func (t *TypeRecord) HasField(field string) bool {
	_, ok := t.LUT[field]
	return ok
}

func (t *TypeRecord) createLUT() {
	t.LUT = make(map[string]int)
	for k, col := range t.Columns {
		t.LUT[string(col.Name)] = k
	}
}

func (t *TypeRecord) Kind() string {
	return "record"
}
