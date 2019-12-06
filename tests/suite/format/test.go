package format

import (
	"github.com/mccanne/zq/pkg/test"
)

var Internal = test.Internal{
	Name:     "format",
	Query:    "*",
	Input:    test.Trim(input),
	Format:   "ndjson",
	Expected: test.Trim(expected),
}

const input = `
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	conn
#fields	foo	bar
#types	string	string
key1 value1	key2 value1
key1 value2	key2 value2`

const expected = `
{"_path":"conn","bar":"key2 value1","foo":"key1 value1"}
{"_path":"conn","bar":"key2 value2","foo":"key1 value2"}
`
