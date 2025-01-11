### Operator

&emsp; **over** &mdash; traverse nested values as a lateral query

### Synopsis

```
over <expr> [, <expr>...]
over <expr> [, <expr>...] [with <var>=<expr> [, ... <var>[=<expr>]] => ( <lateral> )
```
The `over` operator traverses complex values to create a new sequence
of derived values (e.g., the elements of an array) and either
(in the first form) sends the new values directly to its output or
(in the second form) sends the values to a scoped computation as indicated
by `<lateral>`, which may represent any SuperPipe [subquery](../lateral-subqueries.md) operating on the
derived sequence of values as [`this`](../pipeline-model.md#the-special-value-this).

Each expression `<expr>` is evaluated in left-to-right order and derived sequences are
generated from each such result depending on its types:
* an array value generates each of its elements,
* a map value generates a sequence of records of the form `{key:<key>,value:<value>}` for each
entry in the map, and
* all other values generate a single value equal to itself.

Records can be converted to maps with the [`flatten` function](../functions/flatten.md)
resulting in a map that can be traversed,
e.g., if `this` is a record, it can be traversed with `over flatten(this)`.

The nested subquery depicted as `<lateral>` is called a [lateral subquery](../lateral-subqueries.md).

### Examples

_Over evaluates each expression and emits it_
```mdtest-command
echo null | super -z -c 'over 1,2,"foo"' -
```

```mdtest-output
1
2
"foo"
```
_The over clause is evaluated once per each input value_
```mdtest-command
echo "null null" | super -z -c 'over 1,2' -
```

```mdtest-output
1
2
1
2
```
_Array elements are enumerated_
```mdtest-command
echo null | super -z -c 'over [1,2],[3,4,5]' -
```

```mdtest-output
1
2
3
4
5
```
_Over traversing an array_
```mdtest-command
echo '{a:[1,2,3]}' | super -z -c 'over a' -
```

```mdtest-output
1
2
3
```
_Filter the traversed values_

```mdtest-command
echo '{a:[6,5,4]} {a:[3,2,1]}' | super -z -c 'over a |> this % 2 == 0' -
```

```mdtest-output
6
4
2
```
_Aggregate the traversed values_

```mdtest-command
echo '{a:[1,2]} {a:[3,4,5]}' | super -z -c 'over a |> sum(this)' -
```

```mdtest-output
15
```
_Aggregate the traversed values in a lateral query_
```mdtest-command
echo '{a:[1,2]} {a:[3,4,5]}' | super -z -c 'over a => ( sum(this) )' -
```

```mdtest-output
3
12
```
_Access the outer values in a lateral query_
```mdtest-command
echo '{a:[1,2],s:"foo"} {a:[3,4,5],s:"bar"}' |
  super -z -c 'over a with s => (sum(this) |> yield {s,sum:this})' -
```

```mdtest-output
{s:"foo",sum:3}
{s:"bar",sum:12}
```
_Traverse a record by flattening it_
```mdtest-command
echo '{s:"foo",r:{a:1,b:2}} {s:"bar",r:{a:3,b:4}} ' |
  super -z -c 'over flatten(r) with s => (yield {s,key:key[0],value})' -
```

```mdtest-output
{s:"foo",key:"a",value:1}
{s:"foo",key:"b",value:2}
{s:"bar",key:"a",value:3}
{s:"bar",key:"b",value:4}
```
