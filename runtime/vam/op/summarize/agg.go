package summarize

import (
	"fmt"

	"github.com/brimdata/super"
	"github.com/brimdata/super/runtime/vam/expr"
	"github.com/brimdata/super/runtime/vam/expr/agg"
	"github.com/brimdata/super/vector"
	"github.com/brimdata/super/zcode"
)

// XXX use super.Value for slow path stuff, e.g., when the group-by key is
// a complex type.  when we improve the super.Value impl this will get better.

// one aggTable per fixed set of types of aggs and keys.
type aggTable interface {
	update([]vector.Any, []vector.Any)
	materialize() vector.Any
}

type superTable struct {
	table   map[string]aggRow
	aggs    []*expr.Aggregator
	builder *vector.RecordBuilder
}

var _ aggTable = (*superTable)(nil)

type aggRow struct {
	keys  []super.Value
	funcs []agg.Func
}

func (s *superTable) update(keys []vector.Any, args []vector.Any) {
	m := make(map[string][]uint32)
	var n uint32
	if len(keys) > 0 {
		n = keys[0].Len()
	} else {
		n = args[0].Len()
	}
	var keyBytes []byte
	for slot := uint32(0); slot < n; slot++ {
		keyBytes = keyBytes[:0]
		for _, key := range keys {
			keyBytes = key.AppendKey(keyBytes, slot)
		}
		m[string(keyBytes)] = append(m[string(keyBytes)], slot)
	}
	for rowKey, index := range m {
		row, ok := s.table[rowKey]
		if !ok {
			row = s.newRow(keys, index)
			s.table[rowKey] = row
		}
		for i, arg := range args {
			if len(m) > 1 {
				arg = vector.NewView(arg, index)
			}
			row.funcs[i].Consume(arg)
		}
	}
}

func (s *superTable) newRow(keys []vector.Any, index []uint32) aggRow {
	var row aggRow
	for _, agg := range s.aggs {
		row.funcs = append(row.funcs, agg.Pattern())
	}
	var b zcode.Builder
	for _, key := range keys {
		b.Reset()
		key.Serialize(&b, index[0])
		row.keys = append(row.keys, super.NewValue(key.Type(), b.Bytes().Body()))
	}
	return row
}

func (s *superTable) materialize() vector.Any {
	var vecs []vector.Any
	var tags []uint32
	// XXX This should reasonably concat all materialize rows together instead
	// of this crazy Dynamic hack.
	for _, row := range s.table {
		tags = append(tags, uint32(len(tags)))
		vecs = append(vecs, s.materializeRow(row))
	}
	return vector.NewDynamic(tags, vecs)
}

func (s *superTable) materializeRow(row aggRow) vector.Any {
	var vecs []vector.Any
	for _, key := range row.keys {
		vecs = append(vecs, vector.NewConst(key, 1, nil))
	}
	for _, fn := range row.funcs {
		val := fn.Result()
		vecs = append(vecs, vector.NewConst(val, 1, nil))
	}
	return s.builder.New(vecs)
}

type countByString struct {
	nulls   uint64
	table   map[string]uint64
	builder *vector.RecordBuilder
}

func newCountByString(b *vector.RecordBuilder) aggTable {
	return &countByString{builder: b, table: make(map[string]uint64)}
}

func (c *countByString) update(keys []vector.Any, _ []vector.Any) {
	switch val := keys[0].(type) {
	case *vector.String:
		c.count(val)
	case *vector.Dict:
		c.countDict(val.Any.(*vector.String), val.Counts, val.Nulls)
	case *vector.Const:
		c.countFixed(val)
	default:
		panic(fmt.Sprintf("UNKNOWN %T", val))
	}
}

func (c *countByString) count(vec *vector.String) {
	offs := vec.Offsets
	bytes := vec.Bytes
	if vec.Nulls == nil {
		for k := range vec.Len() {
			c.table[string(bytes[offs[k]:offs[k+1]])]++
		}
	} else {
		for k := range vec.Len() {
			if vec.Nulls.Value(k) {
				c.nulls++
			} else {
				c.table[string(bytes[offs[k]:offs[k+1]])]++
			}
		}
	}
}

func (c *countByString) countDict(vec *vector.String, counts []uint32, nulls *vector.Bool) {
	offs := vec.Offsets
	bytes := vec.Bytes
	for k := range vec.Len() {
		c.table[string(bytes[offs[k]:offs[k+1]])] = uint64(counts[k])
	}
	if nulls != nil {
		for k := range nulls.Len() {
			if nulls.Value(k) {
				c.nulls++
			}
		}
	}
}

func (c *countByString) countFixed(vec *vector.Const) {
	val := vec.Value()
	switch val.Type().ID() {
	case super.IDString:
		var nullCnt uint64
		if vec.Nulls != nil {
			for k := range vec.Len() {
				if vec.Nulls.Value(k) {
					nullCnt++
				}
			}
			c.nulls += nullCnt
		}
		c.table[super.DecodeString(val.Bytes())] += uint64(vec.Len()) - nullCnt
	case super.IDNull:
		c.nulls += uint64(vec.Len())
	}
}

func (c *countByString) materialize() vector.Any {
	length := len(c.table)
	counts := make([]uint64, length)
	var bytes []byte
	offs := make([]uint32, length+1)
	var k int
	for key, count := range c.table {
		offs[k] = uint32(len(bytes))
		bytes = append(bytes, key...)
		counts[k] = count
		k++
	}
	offs[k] = uint32(len(bytes))
	var nulls *vector.Bool
	if c.nulls > 0 {
		length++
		counts = append(counts, c.nulls)
		offs = append(offs, uint32(len(bytes)))
		nulls = vector.NewBoolEmpty(uint32(length), nil)
		nulls.Set(uint32(length - 1))
	}
	keyVec := vector.NewString(offs, bytes, nulls)
	countVec := vector.NewUint(super.TypeUint64, counts, nil)
	return c.builder.New([]vector.Any{keyVec, countVec})
}

type Sum struct {
	parent vector.Puller
	zctx   *super.Context
	field  expr.Evaluator
	name   string
	sum    int64
	done   bool
}

func NewSum(zctx *super.Context, parent vector.Puller, name string) *Sum {
	return &Sum{
		parent: parent,
		zctx:   zctx,
		field:  expr.NewDotExpr(zctx, &expr.This{}, name),
		name:   name,
	}
}

func (c *Sum) Pull(done bool) (vector.Any, error) {
	if done {
		_, err := c.parent.Pull(done)
		return nil, err
	}
	if c.done {
		return nil, nil
	}
	for {
		//XXX check context Done
		// XXX PullVec returns a single vector and enumerates through the
		// different underlying types that match a particular projection
		vec, err := c.parent.Pull(false)
		if err != nil {
			return nil, err
		}
		if vec == nil {
			c.done = true
			return c.materialize(c.zctx, c.name), nil
		}
		c.update(vec)
	}
}

func (c *Sum) update(vec vector.Any) {
	if vec, ok := vec.(*vector.Dynamic); ok {
		for _, vec := range vec.Values {
			c.update(vec)
		}
		return
	}
	switch vec := c.field.Eval(vec).(type) {
	case *vector.Int:
		for _, x := range vec.Values {
			c.sum += x
		}
	case *vector.Uint:
		for _, x := range vec.Values {
			c.sum += int64(x)
		}
	case *vector.Dict:
		switch number := vec.Any.(type) {
		case *vector.Int:
			for k, val := range number.Values {
				c.sum += val * int64(vec.Counts[k])
			}
		case *vector.Uint:
			for k, val := range number.Values {
				c.sum += int64(val) * int64(vec.Counts[k])
			}
		}
	}
}

func (c *Sum) materialize(zctx *super.Context, name string) *vector.Record {
	typ := zctx.MustLookupTypeRecord([]super.Field{
		{Type: super.TypeInt64, Name: "sum"},
	})
	return vector.NewRecord(typ, []vector.Any{vector.NewInt(super.TypeInt64, []int64{c.sum}, nil)}, 1, nil)
}
