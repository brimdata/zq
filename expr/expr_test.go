package expr_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/brimsec/zq/ast"
	"github.com/brimsec/zq/expr"
	"github.com/brimsec/zq/zio/detector"
	"github.com/brimsec/zq/zng"
	"github.com/brimsec/zq/zng/resolver"
	"github.com/brimsec/zq/zql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// XXX copied from filter_test.go where could we put a single copy of this?
func parseOneRecord(zngsrc string) (*zng.Record, error) {
	ior := strings.NewReader(zngsrc)
	reader, err := detector.LookupReader("zng", ior, resolver.NewContext())
	if err != nil {
		return nil, err
	}

	rec, err := reader.Read()
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.New("expected to read one record")
	}
	rec.Keep()

	rec2, err := reader.Read()
	if err != nil {
		return nil, err
	}
	if rec2 != nil {
		return nil, errors.New("got more than one record")
	}
	return rec, nil
}

func compileExpr(s string) (expr.ExpressionEvaluator, error) {
	parsed, err := zql.Parse("", []byte(s), zql.Entrypoint("Expression"))
	if err != nil {
		return nil, err
	}

	node, ok := parsed.(ast.Expression)
	if !ok {
		return nil, errors.New("expected Expression")
	}

	return expr.CompileExpr(node)
}

// Compile and evaluate a zql expression against a provided Record.
// Returns the resulting Value if successful or an error otherwise
// (which could be failure to compile the expression or failure while
// evaluating the expression).
func evaluate(e string, record *zng.Record) (zng.Value, error) {
	eval, err := compileExpr(e)
	if err != nil {
		return zng.Value{}, err
	}

	// And execute it.
	return eval(record)
}

func testSuccessful(t *testing.T, e string, record *zng.Record, expect zng.Value) {
	t.Run(e, func(t *testing.T) {
		result, err := evaluate(e, record)
		require.NoError(t, err)

		assert.Equal(t, expect.Type, result.Type, "result type is correct")
		assert.Equal(t, expect.Bytes, result.Bytes, "result value is correct")
	})
}

func zbool(b bool) zng.Value {
	return zng.Value{zng.TypeBool, zng.EncodeBool(b)}
}

func zint32(v int32) zng.Value {
	return zng.Value{zng.TypeInt32, zng.EncodeInt(int64(v))}
}

func zint64(v int64) zng.Value {
	return zng.Value{zng.TypeInt64, zng.EncodeInt(v)}
}

func zfloat64(f float64) zng.Value {
	return zng.Value{zng.TypeFloat64, zng.EncodeFloat64(f)}
}

func zstring(s string) zng.Value {
	return zng.Value{zng.TypeString, zng.EncodeString(s)}
}

func TestExpressions(t *testing.T) {
	TestPrimitives(t)
	TestLogical(t)
	TestCompareEquality(t)
	TestCompareRelative(t)
	TestArithmetic(t)
}

func TestPrimitives(t *testing.T) {
	record, err := parseOneRecord(`
#0:record[x:int32,f:float64,s:string]
0:[10;2.5;hello;]`)
	require.NoError(t, err)

	// Test simple literals
	testSuccessful(t, "50", record, zint64(50))
	testSuccessful(t, "3.14", record, zfloat64(3.14))

	// Test good field references
	testSuccessful(t, "x", record, zint32(10))
	testSuccessful(t, "f", record, zfloat64(2.5))
	testSuccessful(t, "s", record, zstring("hello"))

	// Test bad field references
	_, err = evaluate("doesnexist", record)
	assert.Error(t, err, expr.ErrNoSuchField, "referencing non-existent field gave the correct error")
}

func TestLogical(t *testing.T) {
	record, err := parseOneRecord(`
#0:record[t:bool,f:bool]
0:[T;F;]`)
	require.NoError(t, err)

	testSuccessful(t, "t AND t", record, zbool(true))
	testSuccessful(t, "t AND f", record, zbool(false))
	testSuccessful(t, "f AND t", record, zbool(false))
	testSuccessful(t, "f AND f", record, zbool(false))

	testSuccessful(t, "t OR t", record, zbool(true))
	testSuccessful(t, "t OR f", record, zbool(true))
	testSuccessful(t, "f OR t", record, zbool(true))
	testSuccessful(t, "f OR f", record, zbool(false))
}

func TestCompareEquality(t *testing.T) {
	record, err := parseOneRecord(`
#0:record[b:bool,u8:byte,i16:int16,u16:uint16,i32:int32,u32:uint32,i64:int64,u64:uint64,f:float64,s:string,bs:bstring,i:ip,p:port,net:net,t:time,d:duration]
0:[t;0;1;2;3;4;5;6;7.5;hello;world;10.1.1.1;443;10.1.0.0/16;1583794452;1000;]`)
	require.NoError(t, err)

	testSuccessful(t, "b = true", record, zbool(true))
	testSuccessful(t, "b = false", record, zbool(false))
	testSuccessful(t, "b != true", record, zbool(false))
	testSuccessful(t, "b != false", record, zbool(true))

	testSuccessful(t, "u8 = 0", record, zbool(true))
	testSuccessful(t, "u8 = 1", record, zbool(false))
	testSuccessful(t, "u8 != 0", record, zbool(false))
	testSuccessful(t, "u8 != 1", record, zbool(true))
	testSuccessful(t, "u8 = i16", record, zbool(false))
	testSuccessful(t, "u8 != i16", record, zbool(true))

	testSuccessful(t, "i16 = 1", record, zbool(true))
	testSuccessful(t, "i16 = 2", record, zbool(false))
	testSuccessful(t, "i16 != 1", record, zbool(false))
	testSuccessful(t, "i16 != 2", record, zbool(true))
	testSuccessful(t, "i16 = i32", record, zbool(false))
	testSuccessful(t, "i16 != i32", record, zbool(true))

	testSuccessful(t, "u16 = 2", record, zbool(true))
	testSuccessful(t, "u16 = 3", record, zbool(false))
	testSuccessful(t, "u16 != 2", record, zbool(false))
	testSuccessful(t, "u16 != 3", record, zbool(true))
	testSuccessful(t, "u16 = u32", record, zbool(false))
	testSuccessful(t, "u16 != u32", record, zbool(true))

	testSuccessful(t, "i32 = 3", record, zbool(true))
	testSuccessful(t, "i32 = 4", record, zbool(false))
	testSuccessful(t, "i32 != 3", record, zbool(false))
	testSuccessful(t, "i32 != 4", record, zbool(true))
	testSuccessful(t, "i32 = i64", record, zbool(false))
	testSuccessful(t, "i32 != i64", record, zbool(true))

	testSuccessful(t, "u32 = 4", record, zbool(true))
	testSuccessful(t, "u32 = 5", record, zbool(false))
	testSuccessful(t, "u32 != 4", record, zbool(false))
	testSuccessful(t, "u32 != 5", record, zbool(true))
	testSuccessful(t, "u32 = u64", record, zbool(false))
	testSuccessful(t, "u32 != u64", record, zbool(true))

	testSuccessful(t, "i64 = 5", record, zbool(true))
	testSuccessful(t, "i64 = 6", record, zbool(false))
	testSuccessful(t, "i64 != 5", record, zbool(false))
	testSuccessful(t, "i64 != 6", record, zbool(true))
	testSuccessful(t, "i64 = i32", record, zbool(false))
	testSuccessful(t, "i64 != i32", record, zbool(true))

	testSuccessful(t, "u64 = 6", record, zbool(true))
	testSuccessful(t, "u64 = 7", record, zbool(false))
	testSuccessful(t, "u64 != 6", record, zbool(false))
	testSuccessful(t, "u64 != 7", record, zbool(true))
	testSuccessful(t, "u64 = u32", record, zbool(false))
	testSuccessful(t, "u64 != u32", record, zbool(true))

	testSuccessful(t, "f = 7.5", record, zbool(true))
	testSuccessful(t, "f = 6.5", record, zbool(false))
	testSuccessful(t, "f != 7.5", record, zbool(false))
	testSuccessful(t, "f != 6.5", record, zbool(true))

	// XXX compare float/int
	// testSuccessful(t, "f = u32", record, zbool(false))
	// testSuccessful(t, "f != u32", record, zbool(true))

	testSuccessful(t, `s = "hello"`, record, zbool(true))
	testSuccessful(t, `s != "hello"`, record, zbool(false))
	testSuccessful(t, `s = "world"`, record, zbool(false))
	testSuccessful(t, `s != "world"`, record, zbool(true))
	testSuccessful(t, `bs = "world"`, record, zbool(true))
	testSuccessful(t, `bs != "world"`, record, zbool(false))
	testSuccessful(t, `bs = "hello"`, record, zbool(false))
	testSuccessful(t, `bs != "hello"`, record, zbool(true))
	testSuccessful(t, "s = bs", record, zbool(false))
	testSuccessful(t, "s != bs", record, zbool(true))

	// XXX add ip, port, net, time, duration
}

func TestCompareRelative(t *testing.T) {
	record, err := parseOneRecord(`
#0:record[u8a:byte,u8b:byte,i16a:int16,i16b:int16,u16a:uint16,u16b:uint16,i32a:int32,i32b:int32,u32a:uint32,u32b:uint32,i64a:int64,i64b:int64,u64a:uint64,u64b:uint64]
0:[1;2;1;2;1;2;1;2;1;2;1;2;1;2;]`)
	require.NoError(t, err)

	types := []string{"u8", "i16", "u16", "i32", "u32", "i64", "u64"}
	for _, t1 := range types {
		for _, t2 := range types {
			// For every pair of types, test:

			// a < comparison that is true
			exp := fmt.Sprintf("%sa < %sb", t1, t2)
			testSuccessful(t, exp, record, zbool(true))

			// a < comparison that is false
			exp = fmt.Sprintf("%sb < %sa", t1, t2)
			testSuccessful(t, exp, record, zbool(false))

			// a <= comparison that is true (becuase <)
			exp = fmt.Sprintf("%sa <= %sb", t1, t2)
			testSuccessful(t, exp, record, zbool(true))

			// a <= comparison that is false (because >)
			exp = fmt.Sprintf("%sb <= %sa", t1, t2)
			testSuccessful(t, exp, record, zbool(false))

			// a <= comparison that is true (because ==)
			exp = fmt.Sprintf("%sa <= %sa", t1, t2)
			testSuccessful(t, exp, record, zbool(true))

			// a > comparison that is true
			exp = fmt.Sprintf("%sb > %sa", t1, t2)
			testSuccessful(t, exp, record, zbool(true))

			// a > comparison that is false
			exp = fmt.Sprintf("%sa > %sb", t1, t2)
			testSuccessful(t, exp, record, zbool(false))

			// a >= comparison that is true (because >)
			exp = fmt.Sprintf("%sb >= %sa", t1, t2)
			testSuccessful(t, exp, record, zbool(true))

			// a >= comparison that is false (because <)
			exp = fmt.Sprintf("%sa >= %sb", t1, t2)
			testSuccessful(t, exp, record, zbool(false))

			// a >= comparison that is true (because ==)
			exp = fmt.Sprintf("%sa >= %sa", t1, t2)
			testSuccessful(t, exp, record, zbool(true))
		}
	}

	// XXX strings

	// XXX port, ip, net, time, duration
}

func TestArithmetic(t *testing.T) {
	// XXX should test combinations between all primitive int types
	record, err := parseOneRecord(`
#0:record[x:int32,f:float64,i:ip]
0:[10;2.5;10.1.1.1;]`)
	require.NoError(t, err)

	// Test integer arithmetic
	testSuccessful(t, "100 + 23", record, zint64(123))
	testSuccessful(t, "x + 5", record, zint64(15))
	testSuccessful(t, "5 + x", record, zint64(15))
	testSuccessful(t, "x - 5", record, zint64(5))
	testSuccessful(t, "0 - x", record, zint64(-10))
	testSuccessful(t, "x + 5 - 3", record, zint64(12))
	testSuccessful(t, "x*2", record, zint64(20))
	testSuccessful(t, "5*x*2", record, zint64(100))
	testSuccessful(t, "x/3", record, zint64(3))
	testSuccessful(t, "20/x", record, zint64(2))

	// Test precedence of arithmetic operations
	testSuccessful(t, "x + 1 * 10", record, zint64(20))
	testSuccessful(t, "(x + 1) * 10", record, zint64(110))

	// Test arithmetic with floats
	testSuccessful(t, "f + 1.0", record, zfloat64(3.5))
	testSuccessful(t, "1.0 + f", record, zfloat64(3.5))
	testSuccessful(t, "f - 1.0", record, zfloat64(1.5))
	testSuccessful(t, "0.0 - f", record, zfloat64(-2.5))
	testSuccessful(t, "f * 1.5", record, zfloat64(3.75))
	testSuccessful(t, "1.5 * f", record, zfloat64(3.75))
	testSuccessful(t, "f / 1.25", record, zfloat64(2.0))
	testSuccessful(t, "5.0 / f", record, zfloat64(2.0))

	// Test arithmetic mixing float + int
	testSuccessful(t, "f + 5", record, zfloat64(7.5))
	testSuccessful(t, "5 + f", record, zfloat64(7.5))
	testSuccessful(t, "f + x", record, zfloat64(12.5))
	testSuccessful(t, "x + f", record, zfloat64(12.5))
	testSuccessful(t, "x - f", record, zfloat64(7.5))
	testSuccessful(t, "f - x", record, zfloat64(-7.5))
	testSuccessful(t, "x*f", record, zfloat64(25.0))
	testSuccessful(t, "f*x", record, zfloat64(25.0))
	testSuccessful(t, "x/f", record, zfloat64(4.0))
	testSuccessful(t, "f/x", record, zfloat64(0.25))

	// Test that addition fails on an unsupported type
	_, err = evaluate("i + 1", record)
	assert.Error(t, err, expr.ErrIncompatibleTypes, "adding incompatible types gave the correct error")

	_, err = evaluate("x + i", record)
	assert.Error(t, err, expr.ErrIncompatibleTypes, "adding incompatible types gave the correct error")

	_, err = evaluate("f + i", record)
	assert.Error(t, err, expr.ErrIncompatibleTypes, "adding incompatible types gave the correct error")

	record, err = parseOneRecord(`
#0:record[s:string]
0:[hello;]`)
	require.NoError(t, err)

	// Test string concatenation
	testSuccessful(t, `s + " world"`, record, zstring("hello world"))

	// Test string arithmetic other than + fails
	_, err = evaluate(`s - " world"`, record)
	assert.Error(t, err, expr.ErrIncompatibleTypes, "subtracting strings gave the correct error")

	_, err = evaluate(`s * " world"`, record)
	assert.Error(t, err, expr.ErrIncompatibleTypes, "multiplying strings gave the correct error")

	_, err = evaluate(`s / " world"`, record)
	assert.Error(t, err, expr.ErrIncompatibleTypes, "dividing strings gave the correct error")
}
