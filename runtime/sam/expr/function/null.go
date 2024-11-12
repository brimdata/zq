package function

import "github.com/brimdata/super"

// https://github.com/brimdata/super/blob/main/docs/language/functions.md#isnull
type IsNull struct{}

func (n *IsNull) Call(_ super.Allocator, args []super.Value) super.Value {
	if args[0].IsNull() {
		return super.True
	}
	return super.False
}

// https://github.com/brimdata/super/blob/main/docs/language/functions.md#isnotnull
type IsNotNull struct{}

func (n *IsNotNull) Call(_ super.Allocator, args []super.Value) super.Value {
	if !args[0].IsNull() {
		return super.True
	}
	return super.False
}
