package semantic

import (
	"errors"
	"fmt"
	"strings"

	"github.com/brimdata/super"
	"github.com/brimdata/super/compiler/ast"
	"github.com/brimdata/super/compiler/dag"
	"github.com/brimdata/super/compiler/kernel"
	"github.com/brimdata/super/order"
	"github.com/brimdata/super/pkg/field"
	"github.com/brimdata/super/runtime/sam/expr/agg"
	"github.com/brimdata/super/zson"
)

// Analyze a SQL select expression which may have arbitrary nested subqueries
// and may or may not have its sources embedded.
func (a *analyzer) semSelect(sel *ast.Select, seq dag.Seq) dag.Seq {
	from := sel.From
	if len(seq) > 0 {
		if from != nil {
			a.error(sel, errors.New("SELECT cannot have both an embedded FROM claue and input from parents"))
			return append(seq, badOp())
		}
	} else if from == nil {
		// XXX need to insert null source here, e.g., so "select 1" query works
		a.error(sel, errors.New("SELECT without a FROM claue not yet supported"))
		return append(seq, badOp())
	}
	if from != nil {
		seq = a.semFrom(sel.From, seq)
	}
	if sel.Value {
		return a.semSelectValue(sel, seq)
	}
	selection, err := a.newSQLSelection(sel.Args)
	if err != nil {
		a.error(sel, err)
		return dag.Seq{badOp()}
	}
	if sel.Where != nil {
		seq = append(seq, dag.NewFilter(a.semExpr(sel.Where)))
	}
	if sel.GroupBy != nil {
		groupby, err := a.convertSQLGroupBy(sel.GroupBy, selection)
		if err != nil {
			a.error(sel, err)
			seq = append(seq, badOp())
		} else {
			seq = append(seq, groupby)
			if sel.Having != nil {
				seq = append(seq, dag.NewFilter(a.semExpr(sel.Having)))
			}
		}
	} else if sel.Args != nil {
		if sel.Having != nil {
			a.error(sel, errors.New("HAVING clause used without GROUP BY"))
			seq = append(seq, badOp())
		}
		// GroupBy will do the cutting but if there's no GroupBy,
		// then we need a cut for the select expressions.
		// For SELECT *, cutter is nil.
		selector, err := convertSQLSelect(selection)
		if err != nil {
			a.error(sel, err)
			seq = append(seq, badOp())
		} else {
			seq = append(seq, selector)
		}
	}
	if len(seq) == 0 {
		seq = dag.Seq{dag.PassOp}
	}
	if sel.Distinct {
		seq = a.semDistinct(seq)
	}
	return seq
}

func (a *analyzer) semSelectValue(sel *ast.Select, seq dag.Seq) dag.Seq {
	if sel.GroupBy != nil {
		a.error(sel, errors.New("SELECT VALUE cannot be used with GROUP BY"))
		seq = append(seq, badOp())
	}
	if sel.Having != nil {
		a.error(sel, errors.New("SELECT VALUE cannot be used with HAVING"))
		seq = append(seq, badOp())
	}
	exprs := make([]dag.Expr, 0, len(sel.Args))
	for _, assignment := range sel.Args {
		if assignment.LHS != nil {
			a.error(sel, errors.New("SELECT VALUE cannot AS clause in selection"))
		}
		exprs = append(exprs, a.semExpr(assignment.RHS))
	}
	seq = append(seq, &dag.Yield{
		Kind:  "Yield",
		Exprs: exprs,
	})
	if sel.Where != nil {
		seq = append(seq, dag.NewFilter(a.semExpr(sel.Where)))
	}
	if sel.Distinct {
		seq = a.semDistinct(seq)
	}
	return seq
}

func (a *analyzer) semDistinct(seq dag.Seq) dag.Seq {
	seq = append(seq, &dag.Sort{
		Kind: "Sort",
		Args: []dag.SortExpr{
			{
				Key:   &dag.This{Kind: "This"},
				Order: order.Asc,
			},
		},
	})
	return append(seq, &dag.Uniq{
		Kind: "Uniq",
	})
}

func (a *analyzer) semSQLOp(op ast.Op, seq dag.Seq) dag.Seq {
	switch op := op.(type) {
	case *ast.SQLPipe:
		if len(seq) > 0 {
			panic("semSQLOp: SQL pipes can't have parents")
		}
		return a.semSeq(op.Ops)
	case *ast.Select:
		return a.semSelect(op, seq)
	case *ast.SQLJoin:
		return a.semSQLJoin(op, seq)
	case *ast.OrderBy:
		nullsFirst, ok := nullsFirst(op.Exprs)
		if !ok {
			a.error(op, errors.New("differring nulls first/last clauses not yet supported"))
			return append(seq, badOp())
		}
		var exprs []dag.SortExpr
		for _, e := range op.Exprs {
			exprs = append(exprs, a.semSortExpr(e))
		}
		return append(a.semSQLOp(op.Op, seq), &dag.Sort{
			Kind:       "Sort",
			Args:       exprs,
			NullsFirst: nullsFirst,
			Reverse:    false, //XXX this should go away
		})
	case *ast.Limit:
		e := a.semExpr(op.Count)
		var err error
		val, err := kernel.EvalAtCompileTime(a.zctx, e)
		if err != nil {
			a.error(op.Count, err)
			return append(seq, badOp())
		}
		if !super.IsInteger(val.Type().ID()) {
			a.error(op.Count, fmt.Errorf("expression value must be an integer value: %s", zson.FormatValue(val)))
			return append(seq, badOp())
		}
		limit := val.AsInt()
		if limit < 1 {
			a.error(op.Count, errors.New("expression value must be a positive integer"))
		}
		head := &dag.Head{
			Kind:  "Head",
			Count: int(limit),
		}
		return append(a.semSQLOp(op.Op, seq), head)
	default:
		panic(fmt.Sprintf("semSQLOp: unknown op: %#v", op))
	}
}

// For now, each joining table is on the right...
// We don't have logic to not care about the side of the JOIN ON keys...
func (a *analyzer) semSQLJoin(join *ast.SQLJoin, seq dag.Seq) dag.Seq {
	// XXX  For now we require an alias on the
	// right side and combine the entire right side value into the row
	// using the existing join semantics of assignment where the lval
	// lives in the left record and the rval comes from the right.
	if join.Right.Alias == nil {
		a.error(join.Right, errors.New("SQL joins currently require a table alias on the right lef of the join"))
		seq = append(seq, badOp())
	}
	leftKey, rightKey, err := a.semJoinCond(join.Cond)
	if err != nil {
		a.error(join.Cond, errors.New("SQL joins currently limited to equijoin on fields"))
		return append(seq, badOp())
	}
	//XXX need to pass down parent
	leftPath := a.semFromElem(join.Left)
	rightPath := a.semFromElem(join.Right)

	alias := join.Right.Alias.Text
	assignment := dag.Assignment{
		Kind: "Assignment",
		LHS:  pathOf(alias),
		RHS:  &dag.This{Kind: "This", Path: field.Path{alias}},
	}
	par := &dag.Fork{
		Kind:  "Fork",
		Paths: []dag.Seq{{dag.PassOp}, rightPath},
	}
	dagJoin := &dag.Join{
		Kind:     "Join",
		Style:    join.Style,
		LeftDir:  order.Unknown,
		LeftKey:  leftKey,
		RightDir: order.Unknown,
		RightKey: rightKey,
		Args:     []dag.Assignment{assignment},
	}
	seq = leftPath
	seq = append(seq, par)
	return append(seq, dagJoin)
}

func (a *analyzer) semJoinCond(cond ast.JoinExpr) (*dag.This, *dag.This, error) {
	switch cond := cond.(type) {
	case *ast.JoinOn:
		binary, ok := cond.Expr.(*ast.BinaryExpr)
		if !ok || binary.Op != "==" {
			return nil, nil, errors.New("only equijoins currently supported")
		}
		//XXX we currently require field expressions
		// need to generalize this but that will require work on the
		// runtime join implementation.
		leftKey, ok := a.semField(binary.LHS).(*dag.This)
		if !ok {
			return nil, nil, errors.New("join keys must be field references")
		}
		rightKey, ok := a.semField(binary.RHS).(*dag.This)
		if !ok {
			return nil, nil, errors.New("join keys must be field references")
		}
		return leftKey, rightKey, nil
	case *ast.JoinUsing:
		panic("XXX TBD - JoinUsing")
	default:
		panic(fmt.Sprintf("semJoinCond: unknown type: %T", cond))
	}
}

func nullsFirst(exprs []ast.SortExpr) (bool, bool) {
	if len(exprs) == 0 {
		panic("nullsFirst()")
	}
	if !hasNullsFirst(exprs) {
		return false, true
	}
	// If the nulls firsts are all the same, then we can use
	// nullsfirst; otherwise, if they differ, the runtime currently
	// can't support it.
	for _, e := range exprs {
		if e.Nulls == nil || e.Nulls.Name != "first" {
			return false, false
		}
	}
	return true, true
}

func hasNullsFirst(exprs []ast.SortExpr) bool {
	for _, e := range exprs {
		if e.Nulls != nil && e.Nulls.Name == "first" {
			return true
		}
	}
	return false
}

func convertSQLSelect(selection sqlSelection) (dag.Op, error) {
	// This is a straight select without a group-by.
	// If all the expressions are aggregators, then we build a group-by.
	// If it's mixed, we return an error.  Otherwise, we do a simple cut.
	var nagg int
	for _, p := range selection {
		if p.agg != nil {
			nagg++
		}
	}
	if nagg == 0 {
		return selection.cut(), nil
	}
	if nagg != len(selection) {
		return nil, errors.New("cannot mix aggregations and non-aggregations without a GROUP BY")
	}
	// Note here that we reconstruct the group-by aggregators instead of
	// using the assignments in ast.SqlExpression.Select since the SQL peg
	// parser does not know whether they are aggregators or function calls,
	// but the sqlPick elements have this determined.  So we take the LHS
	// from the original expression and mix it with the agg that was put
	// in sqlPick.
	var assignments []dag.Assignment
	for _, p := range selection {
		a := dag.Assignment{
			Kind: "Assignment",
			LHS:  p.assignment.LHS,
			RHS:  p.agg,
		}
		assignments = append(assignments, a)
	}
	return &dag.Summarize{
		Kind: "Summarize",
		Aggs: assignments,
	}, nil
}

func (a *analyzer) convertSQLGroupBy(groupByKeys []ast.Expr, selection sqlSelection) (dag.Op, error) {
	var keys field.List
	for _, key := range groupByKeys {
		name := a.sqlField(key)
		//XXX is this the best way to handle nil
		if name != nil {
			keys = append(keys, name)
		}
	}
	// Make sure all group-by keys are in the selection.
	all := selection.fields()
	for _, key := range keys {
		//XXX fix this for select *?
		if !key.In(all) {
			if key.HasPrefixIn(all) {
				return nil, fmt.Errorf("'%s': GROUP BY key cannot be a sub-field of the selected value", key)
			}
			return nil, fmt.Errorf("'%s': GROUP BY key not in selection", key)
		}
	}
	// Make sure all scalars are in the group-by keys.
	scalars := selection.scalars()
	for _, f := range scalars.fields() {
		if !f.In(keys) {
			return nil, fmt.Errorf("'%s': selected expression is missing from GROUP BY clause (and is not an aggregation)", f)
		}
	}
	// Now that the selection and keys have been checked, build the
	// key expressions from the scalars of the select and build the
	// aggregators (aka reducers) from the aggregation functions present
	// in the select clause.
	var keyExprs []dag.Assignment
	for _, p := range scalars {
		keyExprs = append(keyExprs, p.assignment)
	}
	var aggExprs []dag.Assignment
	for _, p := range selection.aggs() {
		aggExprs = append(aggExprs, dag.Assignment{
			Kind: "Assignment",
			LHS:  p.assignment.LHS,
			RHS:  p.agg,
		})
	}
	// XXX how to override limit for spills?
	return &dag.Summarize{
		Kind: "Summarize",
		Keys: keyExprs,
		Aggs: aggExprs,
	}, nil
}

// A sqlPick is one column of a select statement.  We bookkeep here whether
// a column is a scalar expression or an aggregation by looking up the function
// name and seeing if it's an aggregator or not.  We also infer the column
// names so we can do SQL error checking relating the selections to the group-by keys.
type sqlPick struct {
	name       field.Path
	agg        *dag.Agg
	assignment dag.Assignment
}

type sqlSelection []sqlPick

func (a *analyzer) newSQLSelection(assignments []ast.Assignment) (sqlSelection, error) {
	var s sqlSelection
	for _, as := range assignments {
		name, err := a.deriveAs(as)
		if err != nil {
			return nil, err
		}
		agg, err := a.isAgg(as.RHS)
		if err != nil {
			return nil, err
		}
		assignment := a.semAssignment(as)
		s = append(s, sqlPick{name, agg, assignment})
	}
	return s, nil
}

func (s sqlSelection) fields() field.List {
	var fields field.List
	for _, p := range s {
		fields = append(fields, p.name)
	}
	return fields
}

func (s sqlSelection) aggs() sqlSelection {
	var aggs sqlSelection
	for _, p := range s {
		if p.agg != nil {
			aggs = append(aggs, p)
		}
	}
	return aggs
}

func (s sqlSelection) scalars() sqlSelection {
	var scalars sqlSelection
	for _, p := range s {
		if p.agg == nil {
			scalars = append(scalars, p)
		}
	}
	return scalars
}

func (s sqlSelection) cut() *dag.Cut {
	if len(s) == 0 {
		return nil
	}
	var a []dag.Assignment
	for _, p := range s {
		a = append(a, p.assignment)
	}
	return &dag.Cut{
		Kind: "Cut",
		Args: a,
	}
}

func (a *analyzer) isAgg(e ast.Expr) (*dag.Agg, error) {
	call, ok := e.(*ast.Call)
	if !ok {
		//XXX this doesn't work for aggs inside of expressions, sum(x)+sum(y)
		return nil, nil
	}
	nameLower := strings.ToLower(call.Name.Name)
	if _, err := agg.NewPattern(nameLower, true); err != nil {
		return nil, nil
	}
	var arg ast.Expr
	if len(call.Args) > 1 {
		return nil, fmt.Errorf("%s: wrong number of arguments", call.Name.Name)
	}
	if len(call.Args) == 1 {
		arg = call.Args[0]
	}
	var dagArg dag.Expr
	if arg != nil {
		dagArg = a.semExpr(arg)
	}
	return &dag.Agg{
		Kind: "Agg",
		Name: nameLower,
		Expr: dagArg,
	}, nil
}

func (a *analyzer) deriveAs(as ast.Assignment) (field.Path, error) {
	sa := a.semAssignment(as)
	if this, ok := sa.LHS.(*dag.This); ok {
		return this.Path, nil
	}
	return nil, fmt.Errorf("AS clause not a field")
}

func (a *analyzer) sqlField(e ast.Expr) field.Path {
	if f, ok := a.semField(e).(*dag.This); ok {
		return f.Path
	}
	return nil
}
