package function

import (
	"math"

	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
)

// https://github.com/brimdata/super/blob/main/docs/language/functions.md#abs.md
type Abs struct {
	zctx *super.Context
}

func (a *Abs) Call(args ...vector.Any) vector.Any {
	vec := vector.Under(args[0])
	switch id := vec.Type().ID(); {
	case super.IsUnsigned(id):
		return vec
	case super.IsSigned(id) || super.IsFloat(id):
		return a.abs(vec)
	}
	return vector.NewWrappedError(a.zctx, "abs: not a number", vec)
}

func (a *Abs) abs(vec vector.Any) vector.Any {
	switch vec := vec.(type) {
	case *vector.Const:
		var val super.Value
		if super.IsFloat(vec.Type().ID()) {
			val = super.NewFloat(vec.Type(), math.Abs(vec.Value().Float()))
		} else {
			v := vec.Value().Int()
			if v < 0 {
				v = -v
			}
			val = super.NewInt(vec.Type(), v)
		}
		return vector.NewConst(val, vec.Len(), vec.Nulls)
	case *vector.View:
		return vector.NewView(a.abs(vec.Any), vec.Index)
	case *vector.Dict:
		return vector.NewDict(a.abs(vec.Any), vec.Index, vec.Counts, vec.Nulls)
	case *vector.Int:
		var ints []int64
		for _, v := range vec.Values {
			if v < 0 {
				v = -v
			}
			ints = append(ints, v)
		}
		return vector.NewInt(vec.Type(), ints, vec.Nulls)
	case *vector.Float:
		var floats []float64
		for _, v := range vec.Values {
			floats = append(floats, math.Abs(v))
		}
		return vector.NewFloat(vec.Type(), floats, vec.Nulls)
	default:
		panic(vec)
	}
}

// https://github.com/brimdata/super/blob/main/docs/language/functions.md#ceil
type Ceil struct {
	zctx *super.Context
}

func (c *Ceil) Call(args ...vector.Any) vector.Any {
	vec := vector.Under(args[0])
	switch id := vec.Type().ID(); {
	case super.IsFloat(id):
		return c.ceil(vec)
	case super.IsNumber(id):
		return vec
	}
	return vector.NewWrappedError(c.zctx, "ceil: not a number", vec)
}

func (c *Ceil) ceil(vec vector.Any) vector.Any {
	switch vec := vec.(type) {
	case *vector.Const:
		val := super.NewFloat(vec.Type(), math.Ceil(vec.Value().Float()))
		return vector.NewConst(val, vec.Len(), vec.Nulls)
	case *vector.View:
		return vector.NewView(c.ceil(vec.Any), vec.Index)
	case *vector.Dict:
		return vector.NewDict(c.ceil(vec.Any), vec.Index, vec.Counts, vec.Nulls)
	case *vector.Float:
		var floats []float64
		for _, v := range vec.Values {
			floats = append(floats, math.Ceil(v))
		}
		return vector.NewFloat(vec.Type(), floats, vec.Nulls)
	default:
		panic(vec)
	}
}
