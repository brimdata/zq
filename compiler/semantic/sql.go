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
	proj, ok := a.newProjection(sel.Args)
	if !ok {
		return dag.Seq{badOp()}
	}
	if sel.Where != nil {
		seq = append(seq, dag.NewFilter(a.semExpr(sel.Where)))
	}
	if sel.GroupBy != nil {
		if proj.hasStar() {
			a.error(sel, errors.New("aggregate mixed with *-selector not yet supported"))
			return append(seq, badOp())
		}
		groupby, ok := a.semGroupBy(sel.GroupBy, proj)
		if !ok {
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
		selector, err := projectSelect(proj)
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

func projectSelect(selection projection) (dag.Op, error) {
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
		return selection.buildOp(), nil
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
			LHS:  p.item.LHS,
			RHS:  p.agg,
		}
		assignments = append(assignments, a)
	}
	return &dag.Summarize{
		Kind: "Summarize",
		Aggs: assignments,
	}, nil
}

func (a *analyzer) semGroupBy(exprs []ast.Expr, proj projection) (dag.Op, bool) {
	// Unlike the original zed runtime, SQL group-by elements do not have explicit
	// keys and may just be a single identifier or an expression.  We don't quite
	// capture the correct scoping here but this is a start before we implement
	// more sophisticated scoping and identifier bindings.  For our binding-in-the-data
	// approach, we can create temp fields for unnamed group-by expressions and
	// drop them on exit from the scope.  For now, we allow only path expressions
	// and match them with equivalent path expressions in the selection.
	var paths field.List
	for _, e := range exprs {
		this, ok := a.semGroupByKey(e)
		if !ok {
			return nil, false
		}
		paths = append(paths, this.Path)
	}
	// Make sure all group-by keys are in the selection.
	all := proj.paths()
	for k, path := range paths {
		if !path.In(all) {
			if path.HasPrefixIn(all) {
				a.error(exprs[k], fmt.Errorf("'%s': GROUP BY key cannot be a sub-field of the selected value", path))
			}
			a.error(exprs[k], fmt.Errorf("'%s': GROUP BY key not in selection", path))
			return nil, false
		}
	}
	// Make sure all scalars are in the group-by keys.
	scalars := proj.scalars()
	for k, path := range scalars.paths() {
		if !path.In(paths) {
			a.error(exprs[k], fmt.Errorf("'%s': selected expression is missing from GROUP BY clause (and is not an aggregation)", path))
			return nil, false
		}
	}
	// Now that the selection and keys have been checked, build the
	// key expressions from the scalars of the select and build the
	// aggregators from the aggregation functions present in the select clause.
	var keyExprs []dag.Assignment
	for _, p := range scalars {
		keyExprs = append(keyExprs, p.item)
	}
	var aggExprs []dag.Assignment
	for _, p := range proj.aggs() {
		aggExprs = append(aggExprs, dag.Assignment{
			Kind: "Assignment",
			LHS:  p.item.LHS, //XXX is this right?
			RHS:  p.agg,
		})
	}
	return &dag.Summarize{
		Kind: "Summarize",
		Keys: keyExprs,
		Aggs: aggExprs,
	}, true
}

// Column of a select statement.  We bookkeep here whether
// a column is a scalar expression or an aggregation by looking up the function
// name and seeing if it's an aggregator or not.  We also infer the column
// names so we can do SQL error checking relating the selections to the group-by keys.
type column struct {
	path field.Path
	agg  *dag.Agg
	item dag.Assignment
}

func (c column) isStar() bool {
	return c.item.LHS == nil && c.item.RHS == nil
}

func isStar(a ast.Assignment) bool {
	return a.LHS == nil && a.RHS == nil
}

type projection []column

func (a *analyzer) newProjection(assignments []ast.Assignment) (projection, bool) {
	conflict := make(map[string]struct{})
	var proj projection
	for _, as := range assignments {
		if isStar(as) {
			proj = append(proj, column{})
			continue
		}
		// We currently support only path expressions as group-by keys and we need to
		// get the name from the selection in case there is an as clause.
		// must be selected by
		path, err := a.deriveAs(as)
		if err != nil {
			a.error(lhsNode(as), err)
			return nil, false
		}
		leaf := path.Leaf()
		if _, ok := conflict[leaf]; ok {
			a.error(lhsNode(as), fmt.Errorf("%q: conflicting name in projection; try an AS clause", leaf))
			return nil, false
		}
		agg, err := a.isAgg(as.RHS)
		if err != nil {
			a.error(as.RHS, err)
			return nil, false
		}
		assignment := a.semAssignment(as)
		proj = append(proj, column{path, agg, assignment})
	}
	return proj, true
}

func lhsNode(as ast.Assignment) ast.Node {
	n := as.LHS
	if n == nil {
		n = as.RHS
	}
	return n
}

func (p projection) hasStar() bool {
	for _, col := range p {
		if col.isStar() {
			return true
		}
	}
	return false
}

func (p projection) paths() field.List {
	var fields field.List
	for _, col := range p {
		fields = append(fields, col.path)
	}
	return fields
}

func (p projection) aggs() projection {
	var aggs projection
	for _, col := range p {
		if col.agg != nil {
			aggs = append(aggs, col)
		}
	}
	return aggs
}

func (p projection) scalars() projection {
	var scalars projection
	for _, col := range p {
		if col.agg == nil {
			scalars = append(scalars, col)
		}
	}
	return scalars
}

func (p projection) buildOp() *dag.Yield {
	if len(p) == 0 {
		return nil
	}
	var elems []dag.RecordElem
	for _, col := range p {
		var elem dag.RecordElem
		if col.isStar() {
			elem = &dag.Spread{
				Kind: "Spread",
				Expr: &dag.This{Kind: "This"},
			}
		} else {
			elem = &dag.Field{
				Kind:  "Field",
				Name:  col.path.Leaf(),
				Value: col.item.RHS,
			}
		}
		elems = append(elems, elem)
	}
	return &dag.Yield{
		Kind: "Yield",
		Exprs: []dag.Expr{
			&dag.RecordExpr{
				Kind:  "RecordExpr",
				Elems: elems,
			},
		},
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

func (a *analyzer) semGroupByKey(in ast.Expr) (*dag.This, bool) {
	e := a.semExpr(in)
	this, ok := e.(*dag.This)
	if !ok {
		a.error(in, errors.New("GROUP BY expressions are not yet supported; try expression in the selection with an AS"))
		return nil, false
	}
	if len(this.Path) == 0 {
		a.error(in, errors.New("cannot use 'this' as GROUP BY expression"))
		return nil, false
	}
	return this, true
}
