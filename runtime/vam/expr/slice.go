package expr

import (
	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
)

type sliceExpr struct {
	zctx                            *super.Context
	containerEval, fromEval, toEval Evaluator
}

func NewSliceExpr(zctx *super.Context, container, from, to Evaluator) Evaluator {
	return &sliceExpr{
		zctx:          zctx,
		containerEval: container,
		fromEval:      from,
		toEval:        to,
	}
}

func (s *sliceExpr) Eval(vec vector.Any) vector.Any {
	vecs := []vector.Any{s.containerEval.Eval(vec)}
	if s.fromEval != nil {
		vecs = append(vecs, s.fromEval.Eval(vec))
	}
	if s.toEval != nil {
		vecs = append(vecs, s.toEval.Eval(vec))
	}
	return vector.Apply(true, s.eval, vecs...)
}

func (s *sliceExpr) eval(vecs ...vector.Any) vector.Any {
	container := vecs[0]
	var from, to vector.Any
	vecs = vecs[1:]
	if s.fromEval != nil {
		from = vecs[0]
		if !super.IsSigned(from.Type().ID()) {
			return vector.NewWrappedError(s.zctx, "slice: from value is not an integer", from)
		}
		vecs = vecs[1:]
	}
	if s.toEval != nil {
		to = vecs[0]
		if !super.IsSigned(to.Type().ID()) {
			return vector.NewWrappedError(s.zctx, "slice: to value is not an integer", from)
		}
	}
	switch vector.KindOf(container) {
	case vector.KindArray, vector.KindSet:
		return s.evalArrayOrSlice(container, from, to)
	case vector.KindBytes, vector.KindString:
		panic("slices on bytes and strings unsupported")
	case vector.KindError:
		return container
	default:
		return vector.NewWrappedError(s.zctx, "sliced value is not array, set, bytes, or string", container)
	}
}

func (s *sliceExpr) evalArrayOrSlice(vec, fromVec, toVec vector.Any) vector.Any {
	from, constFrom := sliceIsConstIndex(fromVec)
	to, constTo := sliceIsConstIndex(toVec)
	slowPath := !constFrom || !constTo
	var index []uint32
	if view, ok := vec.(*vector.View); ok {
		vec, index = view.Any, view.Index
	}
	offsets, inner, nullsIn := arrayOrSetContents(vec)
	newOffsets := []uint32{0}
	var errs []uint32
	var innerIndex []uint32
	var nullsOut *vector.Bool
	for i := range vec.Len() {
		idx := i
		if index != nil {
			idx = index[i]
		}
		if nullsIn.Value(idx) {
			newOffsets = append(newOffsets, newOffsets[len(newOffsets)-1])
			if nullsOut == nil {
				nullsOut = vector.NewBoolEmpty(vec.Len(), nil)
			}
			nullsOut.Set(i)
			continue
		}
		off := offsets[idx]
		size := int64(offsets[idx+1] - off)
		start, end := int64(0), size
		if fromVec != nil {
			if slowPath {
				from, _ = vector.IntValue(fromVec, idx)
			}
			start = sliceIndex(from, size)
		}
		if toVec != nil {
			if slowPath {
				to, _ = vector.IntValue(toVec, idx)
			}
			end = sliceIndex(to, size)
		}
		if start > end || end > size || start < 0 {
			errs = append(errs, i)
			continue
		}
		newOffsets = append(newOffsets, newOffsets[len(newOffsets)-1]+uint32(end-start))
		for k := start; k < end; k++ {
			innerIndex = append(innerIndex, off+uint32(k))
		}

	}
	var out vector.Any
	inner = vector.NewView(inner, innerIndex)
	if vector.KindOf(vec) == vector.KindArray {
		out = vector.NewArray(vec.Type().(*super.TypeArray), newOffsets, inner, nullsOut)
	} else {
		out = vector.NewSet(vec.Type().(*super.TypeSet), newOffsets, inner, nullsOut)
	}
	if nullsOut != nil {
		nullsOut.SetLen(out.Len())
	}
	if len(errs) > 0 {
		errOut := vector.NewStringError(s.zctx, "slice out of bounds", uint32(len(errs)))
		return vector.Combine(out, errs, errOut)
	}
	return out
}

func sliceIsConstIndex(vec vector.Any) (int64, bool) {
	if vec == nil {
		return 0, true
	}
	if c, ok := vec.(*vector.Const); ok && c.Nulls == nil {
		return c.Value().Int(), true
	}
	return 0, false
}

func sliceIndex(idx, size int64) int64 {
	if idx < 0 {
		idx += int64(size)
	}
	return idx
}

func arrayOrSetContents(vec vector.Any) ([]uint32, vector.Any, *vector.Bool) {
	switch vec := vec.(type) {
	case *vector.Array:
		return vec.Offsets, vec.Values, vec.Nulls
	case *vector.Set:
		return vec.Offsets, vec.Values, vec.Nulls
	default:
		panic(vec)
	}
}
