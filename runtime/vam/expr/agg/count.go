package agg

import (
	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
)

type count struct {
	count uint64
}

func (a *count) Consume(vec vector.Any) {
	a.count += uint64(vec.Len())
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
