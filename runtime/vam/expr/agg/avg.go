package agg

import (
	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
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

func (a *avg) Result() super.Value {
	if a.count > 0 {
		return super.NewFloat64(a.sum / float64(a.count))
	}
	return super.NullFloat64
}
