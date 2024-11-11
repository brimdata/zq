package expr

import (
	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
)

type VectorElem struct {
	Value  Evaluator
	Spread Evaluator
}

type ArrayExpr struct {
	elems []VectorElem
	zctx  *super.Context
}

func NewArrayExpr(zctx *super.Context, elems []VectorElem) *ArrayExpr {
	return &ArrayExpr{
		elems: elems,
		zctx:  zctx,
	}
}

func (a *ArrayExpr) Eval(this vector.Any) vector.Any {
	var vecs []vector.Any
	for _, e := range a.elems {
		if e.Spread != nil {
			vecs = append(vecs, e.Spread.Eval(this))
		} else {
			vecs = append(vecs, e.Value.Eval(this))
		}
	}
	return vector.Apply(false, a.eval, vecs...)
}

func (a *ArrayExpr) eval(vecs ...vector.Any) vector.Any {
	n := vecs[0].Len()
	if n == 0 {
		return vector.NewConst(super.Null, 0, nil)
	}
	spreadOffs := make([][]uint32, len(a.elems))
	viewIndexes := make([][]uint32, len(a.elems))
	for i, elem := range a.elems {
		if elem.Spread != nil {
			vecs[i], spreadOffs[i], viewIndexes[i] = a.unwrapSpread(vecs[i])
		}
	}
	offsets := []uint32{0}
	var tags []uint32
	for i := range n {
		var size uint32
		for tag, spreadOff := range spreadOffs {
			if len(spreadOff) == 0 {
				tags = append(tags, uint32(tag))
				size++
				continue
			} else {
				if index := viewIndexes[tag]; index != nil {
					i = index[i]
				}
				off := spreadOff[i]
				for end := spreadOff[i+1]; off < end; off++ {
					tags = append(tags, uint32(tag))
					size++
				}
			}
		}
		offsets = append(offsets, offsets[i]+size)
	}
	var typ super.Type
	var innerVec vector.Any
	if len(vecs) == 1 {
		typ = vecs[0].Type()
		innerVec = vecs[0]
	} else {
		var all []super.Type
		for _, vec := range vecs {
			all = append(all, vec.Type())
		}
		types := super.UniqueTypes(all)
		if len(types) == 1 {
			typ = types[0]
			innerVec = vector.NewDynamic(tags, vecs)
		} else {
			typ = a.zctx.LookupTypeUnion(types)
			innerVec = vector.NewUnion(typ.(*super.TypeUnion), tags, vecs, nil)
		}
	}
	return vector.NewArray(a.zctx.LookupTypeArray(typ), offsets, innerVec, nil)
}

func (a *ArrayExpr) unwrapSpread(vec vector.Any) (vector.Any, []uint32, []uint32) {
	switch vec := vec.(type) {
	case *vector.Array:
		return vec.Values, vec.Offsets, nil
	case *vector.Set:
		return vec.Values, vec.Offsets, nil
	case *vector.View:
		vals, offsets, _ := a.unwrapSpread(vec.Any)
		return vals, offsets, vec.Index
	}
	return vec, nil, nil
}
