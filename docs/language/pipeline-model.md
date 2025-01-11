---
weight: 2
title: Pipeline Model
---

In SuperPipe, each operator takes its input from the output of its upstream operator beginning
either with a data source or with an implied source.

All available operators are listed on the [reference page](operators/_index.md).

## Pipeline Sources

In addition to the data sources specified as files on the `zq` command line,
a source may also be specified with the [`from` operator](operators/from.md).

When running on the command-line, `from` may refer to a file, an HTTP
endpoint, or an [S3](../integrations/amazon-s3.md) URI.  When running in a [SuperDB data lake](../commands/super-db.md), `from` typically
refers to a collection of data called a "data pool" and is referenced using
the pool's name much as SQL references database tables by their name.

For more detail, see the reference page of the [`from` operator](operators/from.md),
but as an example, you might use the `get` form of `from` to fetch data from an
HTTP endpoint and process it with `super`, in this case, to extract the description
and license of a GitHub repository:
```
super -f text -c 'get https://api.github.com/repos/brimdata/super
            |> yield description,license.name'
```
When a query is run on the command-line with `super`, the `from` source is
typically omitted and implied instead by the command-line file arguments.
The input may be stdin via `-` as in
```
echo '"hello, world"' | super  -
```
The examples throughout the language documentation use this "echo pattern"
to standard input of `super -` to illustrate language semantics.
Note that in these examples, the input values are expressed as a sequence of values serialized
in the [Super JSON format](../formats/jsup.md)
and the `super` query text expressed as the `-c` argument of the `super` command
is expressed in the syntax of the SuperPipe language described here.

## Pipeline Operators

Each operator is identified by name and performs a specific operation
on a stream of records.

Some operators, like
[`summarize`](operators/summarize.md) or [`sort`](operators/sort.md),
read all of their input before producing output, though
`summarize` can produce incremental results when the group-by key is
aligned with the order of the input.

For large queries that process all of their input, time may pass before
seeing any output.

On the other hand, most operators produce incremental output by operating
on values as they are produced.  For example, a long running query that
produces incremental output will stream results as they are produced, i.e.,
running `zq` to standard output will display results incrementally.

The [`search`](operators/search.md) and [`where`](operators/where.md)
operators "find" values in their input and drop
the ones that do not match what is being looked for.

The [`yield` operator](operators/yield.md) emits one or more output values
for each input value based on arbitrary [expressions](expressions.md),
providing a convenient means to derive arbitrary output values as a function
of each input value, much like the map concept in the MapReduce framework.

The [`fork` operator](operators/fork.md) copies its input to parallel
branches of a pipeline.  The output of these parallel branches can be combined
in a number of ways:
* merged in sorted order using the [`merge` operator](operators/merge.md),
* joined using the [`join` operator](operators/join.md), or
* combined in an undefined order using the implied [`combine` operator](operators/combine.md).

A pipeline can also be split to multiple branches using the
[`switch` operator](operators/switch.md), in which data is routed to only one
corresponding branch (or dropped) based on the switch clauses.

Switch operators typically
involve multiline SuperPipe programs, which are easiest to edit in a file.  For example,
suppose this text is in a file called `switch.spq`:
```mdtest-input switch.spq
switch this (
  case 1 => yield {val:this,message:"one"}
  case 2 => yield {val:this,message:"two"}
  default => yield {val:this,message:"many"}
) |> merge val
```
Then, running `super` with `-I switch.spq` like so:
```mdtest-command
echo '1 2 3 4' | super -z -I switch.spq -
```
produces
```mdtest-output
{val:1,message:"one"}
{val:2,message:"two"}
{val:3,message:"many"}
{val:4,message:"many"}
```
Note that the output order of the switch branches is undefined (indeed they run
in parallel on multiple threads).  To establish a consistent sequence order,
a [`merge` operator](operators/merge.md)
may be applied at the output of the switch specifying a sort key upon which
to order the upstream data.  Often such order does not matter (e.g., when the output
of the switch hits an [aggregator](aggregates/_index.md)), in which case it is typically more performant
to omit the merge (though the SuperDB runtime will often delete such unnecessary
operations automatically as part optimizing queries when they are compiled).

If no `merge` or `join` is indicated downstream of a `fork` or `switch`,
then the implied `combine` operator is presumed.  In this case, values are
forwarded from the switch to the downstream operator in an undefined order.

## The Special Value `this`

In SuperPipe, there are no looping constructs and variables are limited to binding
values between [lateral scopes](lateral-subqueries.md#lateral-scope).
Instead, the input sequence
to an operator is produced continuously and any output values are derived
from input values.

In contrast to SQL, where a query may refer to input tables by name,
there are no explicit tables and an operator instead refers
to its input values using the special identifier `this`.

For example, sorting the following input
```mdtest-command
echo '"foo" "bar" "BAZ"' | super -z -c sort -
```
produces this case-sensitive output:
```mdtest-output
"BAZ"
"bar"
"foo"
```
But we can make the sort case-insensitive by applying a [function](functions/_index.md) to the
input values with the expression `lower(this)`, which converts
each value to lower-case for use in in the sort without actually modifying
the input value, e.g.,
```
echo '"foo" "bar" "BAZ"' | super -z -c 'sort lower(this)' -
```
produces
```
"bar"
"BAZ"
"foo"
```

## Implied Field References

A common SuperPipe use case is to process sequences of record-oriented data
(e.g., arising from formats like JSON or Avro) in the form of events
or structured logs.  In this case, the input values to the operators
are [records](../formats/zed.md#21-record) and the fields of a record are referenced with the dot operator.

For example, if the input above were a sequence of records instead of strings
and perhaps contained a second field, e.g.,
```
{s:"foo",x:1}
{s:"bar",x:2}
{s:"BAZ",x:3}
```
Then we could refer to the field `s` using `this.s` and sort the records
as above with `sort this.s`, which would give
```
{s:"BAZ",x:3}
{s:"bar",x:2}
{s:"foo",x:1}
```
This pattern is so common that field references to `this` may be shortened
by simply referring to the field by name wherever an expression is expected,
e.g.,
```
sort s
```
is shorthand for `sort this.s`

## Field Assignments

A typical operation in records involves
adding or changing the fields of a record using the [`put` operator](operators/put.md)
or extracting a subset of fields using the [`cut` operator](operators/cut.md).
Also, when aggregating data using group-by keys, the group-by assignments
create new named record fields.

In all of these cases, the SuperPipe language uses the token `:=` to denote
field assignment.  For example,
```
put x:=y+1
```
or
```
summarize salary:=sum(income) by address:=lower(address)
```
This style of "assignment" to a record value is distinguished from the `=`
token which binds a locally scoped name to a value that can be referenced
in later expressions.

## Implied Operators

When SuperPipe is utilized in an application like [SuperDB Desktop](https://zui.brimdata.io),
queries are often composed interactively in a "search bar" experience.
The language design here attempts to support both this "lean forward" pattern of usage
along with a "coding style" of query writing where the queries might be large
and complex, e.g., to perform transformations in a data pipeline, where
the SuperPipe queries are stored under source-code control perhaps in GitHub or
in Zui's query library.

To facilitate both a programming-like model as well as an ad hoc search
experience, SuperPipe has a canonical, long form that can be abbreviated
using syntax that supports an agile, interactive query workflow.
To this end, SuperPipe allows certain operator names to be optionally omitted when
they can be inferred from context.  For example, the expression following
the [`summarize` operator](operators/summarize.md)
```
summarize count() by id
```
is unambiguously an aggregation and can be shortened to
```
count() by id
```
Likewise, a very common lean-forward use pattern is "searching" so by default,
expressions are interpreted as keyword searches, e.g.,
```
search foo bar or x > 100
```
is abbreviated
```
foo bar or x > 100
```
Furthermore, if an operator-free expression is not valid syntax for
a search expression but is a valid [expression](expressions.md),
then the abbreviation is treated as having an implied `yield` operator, e.g.,
```
{s:lower(s)}
```
is shorthand for
```
yield {s:lower(s)}
```
When operator names are omitted, `search` has precedence over `yield`, so
```
foo
```
is interpreted as a search for the string "foo" rather than a yield of
the implied record field named `foo`.

Another common query pattern involves adding or mutating fields of records
where the input is presumed to be a sequence of records.
The [`put` operator](operators/put.md) provides this mechanism and the `put`
keyword is implied by the [field assignment](#field-assignments) syntax `:=`.

For example, the operation
```
put y:=2*x+1
```
can be expressed simply as
```
y:=2*x+1
```
When composing long-form queries that are shared via SuperDB Desktop or managed in GitHub,
it is best practice to include all operator names in the source text.

In summary, if no operator name is given, the implied operator is determined
from the operator-less source text, in the order given, as follows:
* If the text can be interpreted as a search expression, then the operator is `search`.
* If the text can be interpreted as a boolean expression, then the operator is `where`.
* If the text can be interpreted as one or more field assignments, then the operator is `put`.
* If the text can be interpreted as an aggregation, then the operator is `summarize`.
* If the text can be interpreted as an expression, then the operator is `yield`.
* Otherwise, the text causes a compile-time error.

When in doubt, you can always check what the compiler is doing under the hood
by running `zq` with the `-C` flag to print the parsed query in "canonical form", e.g.,
```mdtest-command
super -C -c '? foo'
super -C -c 'is(<foo>)'
super -C -c 'count()'
super -C -c '{a:x+1,b:y-1}'
super -C -c 'a:=x+1,b:=y-1'
```
produces
```mdtest-output
search foo
where is(<foo>)
summarize
    count()
yield {a:x+1,b:y-1}
put a:=x+1,b:=y-1
```
