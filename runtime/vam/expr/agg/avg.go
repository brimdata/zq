package agg

import (
	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
	"github.com/brimdata/super/zcode"
)

type avg struct {
	sum   float64
	count uint64
}

var _ Func = (*avg)(nil)

func (a *avg) Consume(vec vector.Any) {
	if isNull(vec) {
		return
	}
	a.count += uint64(vec.Len())
	a.sum = sum(a.sum, vec)
}

func (a *avg) Result(*super.Context) super.Value {
	if a.count > 0 {
		return super.NewFloat64(a.sum / float64(a.count))
	}
	return super.NullFloat64
}

const (
	sumName   = "sum"
	countName = "count"
)

func (a *avg) ConsumeAsPartial(partial vector.Any) {
	rec, ok := partial.(*vector.Record)
	if !ok || rec.Len() != 1 {
		panic("avg: invalid partial")
	}
	si, ok1 := rec.Typ.IndexOfField(sumName)
	ci, ok2 := rec.Typ.IndexOfField(countName)
	if !ok1 || !ok2 {
		panic("avg: invalid partial")
	}
	sumVal, ok1 := rec.Fields[si].(*vector.Const)
	countVal, ok2 := rec.Fields[ci].(*vector.Const)
	if !ok1 || !ok2 || sumVal.Type() != super.TypeFloat64 || countVal.Type() != super.TypeUint64 {
		panic("avg: invalid partial")
	}
	a.sum += sumVal.Value().Float()
	a.count += countVal.Value().Uint()
}

func (a *avg) ResultAsPartial(zctx *super.Context) super.Value {
	var zv zcode.Bytes
	zv = super.NewFloat64(a.sum).Encode(zv)
	zv = super.NewUint64(a.count).Encode(zv)
	typ := zctx.MustLookupTypeRecord([]super.Field{
		super.NewField(sumName, super.TypeFloat64),
		super.NewField(countName, super.TypeUint64),
	})
	return super.NewValue(typ, zv)
}
