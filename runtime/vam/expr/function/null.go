package function

import (
	"github.com/brimdata/super"
	"github.com/brimdata/super/vector"
)

// https://github.com/brimdata/super/blob/main/docs/language/functions.md#isnull
type IsNull struct{}

func (n *IsNull) Call(args ...vector.Any) vector.Any {
	vec := underAll(args)[0]
	if c, ok := vec.(*vector.Const); ok && c.Value().IsNull() {
		return vector.NewConst(super.True, c.Len(), nil)
	}
	nulls := vector.NullsOf(vec)
	if nulls == nil {
		nulls = vector.NewBoolEmpty(vec.Len(), nil)
	}
	return nulls
}

// https://github.com/brimdata/super/blob/main/docs/language/functions.md#isnotnull
type IsNotNull struct{}

func (n *IsNotNull) Call(args ...vector.Any) vector.Any {
	vec := underAll(args)[0]
	if c, ok := vec.(*vector.Const); ok && c.Value().IsNull() {
		return vector.NewConst(super.False, c.Len(), nil)
	}
	nulls := vector.NullsOf(vec)
	if nulls == nil {
		nulls = vector.NewBoolEmpty(vec.Len(), nil)
	}
	return vector.Not(nulls)
}
