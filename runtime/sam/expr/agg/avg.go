package agg

import (
	"errors"
	"fmt"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/runtime/sam/expr/coerce"
	"github.com/brimdata/zed/zcode"
	"github.com/brimdata/zed/zson"
)

type Avg struct {
	sum   float64
	count uint64
}

var _ Function = (*Avg)(nil)

func (a *Avg) Consume(val zed.Value) {
	if val.IsNull() {
		return
	}
	if d, ok := coerce.ToFloat(val); ok {
		a.sum += float64(d)
		a.count++
	}
}

func (a *Avg) Result(*zed.Context, *zed.Arena) zed.Value {
	if a.count > 0 {
		return zed.NewFloat64(a.sum / float64(a.count))
	}
	return zed.NullFloat64
}

const (
	sumName   = "sum"
	countName = "count"
)

func (a *Avg) ConsumeAsPartial(arena *zed.Arena, partial zed.Value) {
	sumVal := partial.Deref(arena, sumName)
	if sumVal == nil {
		panic(errors.New("avg: partial sum is missing"))
	}
	if sumVal.Type() != zed.TypeFloat64 {
		panic(fmt.Errorf("avg: partial sum has bad type: %s", zson.FormatValue(*sumVal)))
	}
	countVal := partial.Deref(arena, countName)
	if countVal == nil {
		panic("avg: partial count is missing")
	}
	if countVal.Type() != zed.TypeUint64 {
		panic(fmt.Errorf("avg: partial count has bad type: %s", zson.FormatValue(*countVal)))
	}
	a.sum += sumVal.Float()
	a.count += countVal.Uint()
}

func (a *Avg) ResultAsPartial(zctx *zed.Context, arena *zed.Arena) zed.Value {
	var zv zcode.Bytes
	zv = zed.NewFloat64(a.sum).Encode(zv)
	zv = zed.NewUint64(a.count).Encode(zv)
	typ := zctx.MustLookupTypeRecord([]zed.Field{
		zed.NewField(sumName, zed.TypeFloat64),
		zed.NewField(countName, zed.TypeUint64),
	})
	return arena.New(typ, zv)
}
