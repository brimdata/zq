package agg

import (
	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
)

type count struct {
	count uint64
}

func (a *count) Consume(vec vector.Any) {
	if c, ok := vec.(*vector.Const); ok {
		val := c.Value()
		if !val.IsNull() && !val.IsError() {
			a.count += uint64(vec.Len())
		}
		return
	}
	if _, ok := vector.Under(vec).Type().(*super.TypeError); ok {
		return
	}
	nulls := vector.NullsOf(vec)
	if nulls == nil {
		a.count += uint64(vec.Len())
		return
	}
	for i := range vec.Len() {
		if !nulls.Value(i) {
			a.count++
		}
	}
}

func (a *count) Result(*super.Context) super.Value {
	return super.NewUint64(a.count)
}

func (a *count) ConsumeAsPartial(partial vector.Any) {
	c, ok := partial.(*vector.Const)
	if !ok || c.Len() != 1 || partial.Type() != super.TypeUint64 {
		panic("count: bad partial")
	}
	a.count += c.Value().Uint()
}

func (a *count) ResultAsPartial(*super.Context) super.Value {
	return a.Result(nil)
}
