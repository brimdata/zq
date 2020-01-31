package proc_test

import (
	"testing"

	"github.com/mccanne/zq/proc"
)

func TestTop(t *testing.T) {
	const in = `
#0:record[foo:count]
0:[-;]
0:[1;]
0:[2;]
0:[3;]
0:[4;]
0:[5;]
`
	const out = `
#0:record[foo:count]
0:[5;]
0:[4;]
0:[3;]
`
	proc.TestOneProc(t, in, out, "top 3 foo")
}
