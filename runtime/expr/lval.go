package expr

import (
	"errors"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/pkg/field"
	"github.com/brimdata/zed/runtime/expr/dynfield"
	"github.com/brimdata/zed/zson"
)

type Lval struct {
	Elems      []LvalElem
	cache      []zed.Value
	fieldCache field.Path
}

func NewLval(evals []LvalElem) *Lval {
	return &Lval{Elems: evals}
}

// Eval returns the path of the lval. If there's an error the returned *zed.Value
// will not be nill.
func (l *Lval) Eval(ectx Context, this *zed.Value) (dynfield.Path, error) {
	l.cache = l.cache[:0]
	for _, e := range l.Elems {
		val, err := e.Eval(ectx, this)
		if err != nil {
			return nil, err
		}
		l.cache = append(l.cache, *val)
	}
	return l.cache, nil
}

func (l *Lval) EvalAsRecordPath(ectx Context, this *zed.Value) (field.Path, error) {
	l.fieldCache = l.fieldCache[:0]
	for _, e := range l.Elems {
		val, err := e.Eval(ectx, this)
		if err != nil {
			return nil, err
		}
		if !val.IsString() {
			// XXX Add context to error so we know what element is failing but
			// let's wait until we can test this so we have a feel for what we
			// want to see.
			return nil, errors.New("field reference is not a string")
		}
		l.fieldCache = append(l.fieldCache, val.AsString())
	}
	return l.fieldCache, nil
}

// Path returns the receiver's path.  Path returns false when the receiver
// contains a dynamic element.
func (l *Lval) Path() (field.Path, bool) {
	var path field.Path
	for _, e := range l.Elems {
		s, ok := e.(*StaticLvalElem)
		if !ok {
			return nil, false
		}
		path = append(path, s.Name)
	}
	return path, true
}

type LvalElem interface {
	Eval(ectx Context, this *zed.Value) (*zed.Value, error)
}

type StaticLvalElem struct {
	Name string
}

func (l *StaticLvalElem) Eval(_ Context, _ *zed.Value) (*zed.Value, error) {
	return zed.NewString(l.Name), nil
}

type ExprLvalElem struct {
	caster Evaluator
	eval   Evaluator
}

func NewExprLvalElem(zctx *zed.Context, e Evaluator) *ExprLvalElem {
	return &ExprLvalElem{
		eval:   e,
		caster: LookupPrimitiveCaster(zctx, zed.TypeString),
	}
}

func (l *ExprLvalElem) Eval(ectx Context, this *zed.Value) (*zed.Value, error) {
	val := l.eval.Eval(ectx, this)
	if val.IsError() {
		return nil, lvalErr(ectx, val)
	}
	return val, nil
}

func lvalErr(ectx Context, errVal *zed.Value) error {
	val := ectx.NewValue(errVal.Type.(*zed.TypeError).Type, errVal.Bytes())
	if val.IsString() {
		return errors.New(val.AsString())
	}
	return errors.New(zson.FormatValue(val))
}
