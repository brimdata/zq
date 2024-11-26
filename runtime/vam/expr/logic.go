package expr

import (
	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
)

type Not struct {
	zctx *super.Context
	expr Evaluator
}

var _ Evaluator = (*Not)(nil)

func NewLogicalNot(zctx *super.Context, e Evaluator) *Not {
	return &Not{zctx, e}
}

func (n *Not) Eval(val vector.Any) vector.Any {
	return evalBool(n.zctx, n.eval, n.expr.Eval(val))
}

func (n *Not) eval(vecs ...vector.Any) vector.Any {
	switch vec := vecs[0].(type) {
	case *vector.Bool:
		bits := make([]uint64, len(vec.Bits))
		for k := range bits {
			bits[k] = ^vec.Bits[k]
		}
		return vec.CopyWithBits(bits)
	case *vector.Const:
		return vector.NewConst(super.NewBool(!vec.Value().Bool()), vec.Len(), vec.Nulls)
	case *vector.Error:
		return vec
	default:
		panic(vec)
	}
}

type And struct {
	zctx *super.Context
	lhs  Evaluator
	rhs  Evaluator
}

func NewLogicalAnd(zctx *super.Context, lhs, rhs Evaluator) *And {
	return &And{zctx, lhs, rhs}
}

type Or struct {
	zctx *super.Context
	lhs  Evaluator
	rhs  Evaluator
}

func NewLogicalOr(zctx *super.Context, lhs, rhs Evaluator) *Or {
	return &Or{zctx, lhs, rhs}
}

func (a *And) Eval(val vector.Any) vector.Any {
	return evalBool(a.zctx, a.eval, a.lhs.Eval(val), a.rhs.Eval(val))
}

func (a *And) eval(vecs ...vector.Any) vector.Any {
	if vecs[0].Len() == 0 {
		return vecs[0]
	}
	lhs, rhs := vector.Under(vecs[0]), vector.Under(vecs[1])
	if _, ok := lhs.(*vector.Error); ok {
		return a.andError(lhs, rhs)
	}
	if _, ok := rhs.(*vector.Error); ok {
		return a.andError(rhs, lhs)
	}
	blhs, brhs := toBool(lhs), toBool(rhs)
	out := vector.And(blhs, brhs)
	if blhs.Nulls == nil && brhs.Nulls == nil {
		return out
	}
	// any and false = false
	// null and true = null
	notfalse := vector.And(vector.Or(blhs, blhs.Nulls), vector.Or(brhs, brhs.Nulls))
	out.Nulls = vector.And(notfalse, vector.Or(blhs.Nulls, brhs.Nulls))
	return out
}

func (a *And) andError(err vector.Any, vec vector.Any) vector.Any {
	if _, ok := vec.(*vector.Error); ok {
		return err
	}
	b := toBool(vec)
	// anything and FALSE = FALSE
	isError := vector.Or(b, b.Nulls)
	var index []uint32
	for i := range err.Len() {
		if isError.Value(i) {
			index = append(index, i)
		}
	}
	if len(index) > 0 {
		base := vector.NewInverseView(vec, index)
		return vector.Combine(base, index, vector.NewView(err, index))
	}
	return vec
}

func (o *Or) Eval(val vector.Any) vector.Any {
	return evalBool(o.zctx, o.eval, o.lhs.Eval(val), o.rhs.Eval(val))
}

func (o *Or) eval(vecs ...vector.Any) vector.Any {
	if vecs[0].Len() == 0 {
		return vecs[0]
	}
	lhs, rhs := vector.Under(vecs[0]), vector.Under(vecs[1])
	if _, ok := lhs.(*vector.Error); ok {
		return o.orError(lhs, rhs)
	}
	if _, ok := rhs.(*vector.Error); ok {
		return o.orError(rhs, lhs)
	}
	blhs, brhs := toBool(lhs), toBool(rhs)
	out := vector.Or(blhs, brhs)
	if blhs.Nulls == nil && brhs.Nulls == nil {
		return out
	}
	nulls := vector.Or(blhs.Nulls, brhs.Nulls)
	out.Nulls = vector.And(vector.Not(out), nulls)
	return out
}

func (o *Or) orError(err, vec vector.Any) vector.Any {
	if _, ok := vec.(*vector.Error); ok {
		return err
	}
	b := toBool(vec)
	// not error if true or null
	notError := vector.Or(b, b.Nulls)
	var index []uint32
	for i := range b.Len() {
		if !notError.Value(i) {
			index = append(index, i)
		}
	}
	if len(index) > 0 {
		base := vector.NewInverseView(vec, index)
		return vector.Combine(base, index, vector.NewView(err, index))
	}
	return vec
}

// evalBool evaluates e using val to computs a boolean result.  For elements
// of the result that are not boolean, an error is calculated for each non-bool
// slot and they are returned as an error.  If all of the value slots are errors,
// then the return value is nil.
func evalBool(zctx *super.Context, fn func(...vector.Any) vector.Any, vecs ...vector.Any) vector.Any {
	return vector.Apply(false, func(vecs ...vector.Any) vector.Any {
		for i, vec := range vecs {
			if vec := vector.Under(vec); vec.Type() == super.TypeBool || vector.KindOf(vec) == vector.KindError {
				vecs[i] = vec
			} else {
				vecs[i] = vector.NewWrappedError(zctx, "not type bool", vec)
			}
		}
		return fn(vecs...)
	}, vecs...)
}

func toBool(vec vector.Any) *vector.Bool {
	switch vec := vec.(type) {
	case *vector.Const:
		val := vec.Value()
		if val.Bool() {
			out := trueBool(vec.Len())
			out.Nulls = vec.Nulls
			return out
		} else {
			return vector.NewBoolEmpty(vec.Len(), vec.Nulls)
		}
	case *vector.Bool:
		return vec
	default:
		panic(vec)
	}
}

func trueBool(n uint32) *vector.Bool {
	vec := vector.NewBoolEmpty(n, nil)
	for i := range vec.Bits {
		vec.Bits[i] = ^uint64(0)
	}
	return vec
}
