---
sidebar_position: 1
sidebar_label: super
---

# `super`

> **TL;DR** `super` is a command-line tool that uses [SuperSQL](../language/README.md)
> to query a variety of data formats in files, over HTTP, or in [S3](../integrations/amazon-s3.md)
> storage. Best performance is achieved when operating on data in binary formats such as
> [Super Binary](../formats/bsup.md), [Super Columnar](../formats/csup.md),
> [Parquet](https://github.com/apache/parquet-format), or
> [Arrow](https://arrow.apache.org/docs/format/Columnar.html#ipc-streaming-format).

## Usage

```
super [ options ] [ -c query ] input [ input ... ]
```

`super` is a command-line tool for processing data in diverse input
formats, providing data wrangling, search, analytics, and extensive transformations
using the [SuperSQL](../language/README.md) dialect of SQL. Any SQL query expression
may be extended with [pipe syntax](https://research.google/pubs/sql-has-problems-we-can-fix-them-pipe-syntax-in-sql/)
to filter, transform, and/or analyze input data.
Super's SQL pipes dialect is extensive, so much so that it can resemble
a log-search experience despite its SQL foundation.

The `super` command works with data from ephemeral sources like files and URLs.
If you want to persist your data into a data lake for persistent storage,
check out the [`super db`](super-db.md) set of commands.

By invoking the `-c` option, a query expressed in the [SuperSQL language](../language/README.md)
may be specified and applied to the input stream.

Super's data model is based on super-structured data, meaning that all data
is both strongly _and_ dynamically typed and need not conform to a homogeneous
schema.  The type structure is self-describing so it's easy to daisy-chain
queries and inspect data at any point in a complex query or data pipeline.
For example, there's no need for a set of Parquet input files to all be
schema-compatible and it's easy to mix and match Parquet with JSON across
queries.

When processing JSON data, all values are converted to strongly typed values
that fit naturally alongside relational data so there is no need for a separate
"JSON type".  Unlike SQL systems that integrate JSON data,
there isn't a JSON way to do things and a separate relational way
to do things.

Because there are no schemas, there is no schema inference, so inferred schemas
do not haphazardly change when input data changes in subtle ways.

Each `input` argument to `super` must be a file path, an HTTP or HTTPS URL,
an S3 URL, or standard input specified with `-`.
These input arguments are treated as if a SQL "from" operator precedes
the provided query, e.g.,
```
super -c "from example.json | select typeof(this)"
```
is equivalent to
```
super -c "select typeof(this)" example.json
```
Output is written to one or more files or to standard output in the format specified.

When multiple input files are specified, they are processed in the order given as
if the data were provided by a single, concatenated "from" clause.

If no query is specified with `-c`, the inputs are scanned without modification
and output in the desired format as [described below](#input-formats),
providing a convenient means to convert files from one format to another, e.g.,
```
super -f arrows file1.json file2.parquet file3.csv > file-combined.arrows
```
When `super` is run with a query that has no "from" operator and no input arguments,
the SuperSQL query is fed a single `null` value analogous to SQL's default
input of a single empty row of an unnamed table.
This provides a convenient means to explore examples or run in a
"calculator mode", e.g.,
```mdtest-command
super -z -c '1+1'
```
emits
```mdtest-output
2
```
Note that SuperSQL's has syntactic shortcuts for interactive data exploration and
an expression that stands alone is a shortcut for `select value`, e.g., the query text
```
1+1
```
is equivalent to
```
select value 1+1
```
To learn more about shortcuts, refer to the SuperSQL
[documentation on shortcuts](../language/pipeline-model.md#implied-operators).

For built-in command help and a listing of all available options,
simply run `super` with no arguments.

## Data Formats

`super` supports a number of [input](#input-formats) and [output](#output-formats) formats, but the super formats
([Super Binary](../formats/bsup.md),
[Super Columnar](../formats/csup.md),
and [Super JSON](../formats/jsup.md)) tend to be the most versatile and
easy to work with.

`super` typically operates on binary-encoded data and when you want to inspect
human-readable bits of output, you merely format it as Super JSON, which is the
default format when output is directed to the terminal.  Super Binary is the default
when redirecting to a non-terminal output like a file or pipe.

Unless the `-i` option specifies a specific input format,
each input's format is [automatically inferred](#auto-detection)
and each input is scanned
in the order appearing on the command line forming the input stream.

### Input Formats

`super` currently supports the following input formats:

|  Option   | Auto | Specification                            |
|-----------|------|------------------------------------------|
| `arrows`  |  yes | [Arrow IPC Stream Format](https://arrow.apache.org/docs/format/Columnar.html#ipc-streaming-format) |
| `bsup`    |  yes | [Super Binary](../formats/bsup.md) |
| `csup`    |  yes | [Super Columnar](../formats/csup.md) |
| `csv`     |  yes | [Comma-Separated Values (RFC 4180)](https://www.rfc-editor.org/rfc/rfc4180.html) |
| `json`    |  yes | [JSON (RFC 8259)](https://www.rfc-editor.org/rfc/rfc8259.html) |
| `jsup`    |  yes | [Super JSON](../formats/jsup.md) |
| `zjson`   |  yes | [Super JSON over JSON](../formats/zjson.md) |
| `line`    |  no  | One string value per input line |
| `parquet` |  yes | [Apache Parquet](https://github.com/apache/parquet-format) |
| `tsv`     |  yes | [Tab-Separated Values](https://en.wikipedia.org/wiki/Tab-separated_values) |
| `zeek`    |  yes | [Zeek Logs](https://docs.zeek.org/en/master/logs/index.html) |

The input format is typically [detected automatically](#auto-detection) and the formats for which
"Auto" is "yes" in the table above support _auto-detection_.
Formats without auto-detection require the `-i` option.

### Hard-wired Input Format

The input format is specified with the `-i` flag.

When `-i` is specified, all of the inputs on the command-line must be
in the indicated format.

### Auto-detection

When using _auto-detection_, each input's format is independently determined
so it is possible to easily blend different input formats into a unified
output format.

For example, suppose this content is in a file `sample.csv`:
```mdtest-input sample.csv
a,b
1,foo
2,bar
```
and this content is in `sample.json`
```mdtest-input sample.json
{"a":3,"b":"baz"}
```
then the command
```mdtest-command
super -z sample.csv sample.json
```
would produce this output in the default Super JSON format
```mdtest-output
{a:1.,b:"foo"}
{a:2.,b:"bar"}
{a:3,b:"baz"}
```

### JSON Auto-detection: Super vs. Plain

Since [Super JSON](../formats/jsup.md) is a superset of plain JSON, `super` must be careful how it distinguishes the two cases when performing auto-inference.
While you can always clarify your intent
with the `-i jsup` or `-i json`, `super` attempts to "just do the right thing"
when you run it with Super JSON vs. plain JSON.

While `super` can parse any JSON using its built-in Super JSON parser this is typically
not desirable because (1) the Super JSON parser is not particularly performant and
(2) all JSON numbers are floating point but the Super JSON parser will parse as
JSON any number that appears without a decimal point as an integer type.

:::tip note
The reason `super` is not particularly performant for Super JSON is that the [Super Binary](../formats/bsup.md) or
[Super Columnar](../formats/csup.md) formats are semantically equivalent to Super JSON but much more efficient and
the design intent is that these efficient binary formats should be used in
use cases where performance matters.  Super JSON is typically used only when
data needs to be human-readable in interactive settings or in automated tests.
:::

To this end, `super` uses a heuristic to select between Super JSON and plain JSON when the
`-i` option is not specified. Specifically, plain JSON is selected when the first values
of the input are parsable as valid JSON and includes a JSON object either
as an outer object or as a value nested somewhere within a JSON array.

This heuristic almost always works in practice because Super JSON records
typically omit quotes around field names.

### Output Formats

`super` currently supports the following output formats:

|  Option   | Specification                            |
|-----------|------------------------------------------|
| `arrows`  | [Arrow IPC Stream Format](https://arrow.apache.org/docs/format/Columnar.html#ipc-streaming-format) |
| `bsup`    | [Super Binary](../formats/bsup.md) |
| `csup`    | [Super Columnar](../formats/csup.md) |
| `csv`     | [Comma-Separated Values (RFC 4180)](https://www.rfc-editor.org/rfc/rfc4180.html) |
| `json`    | [JSON (RFC 8259)](https://www.rfc-editor.org/rfc/rfc8259.html) |
| `jsup`    | [Super JSON](../formats/jsup.md) |
| `zjson`   | [Super JSON over JSON](../formats/zjson.md) |
| `lake`    | [SuperDB Data Lake Metadata Output](#superdb-data-lake-metadata-output) |
| `parquet` | [Apache Parquet](https://github.com/apache/parquet-format) |
| `table`   | (described [below](#simplified-text-outputs)) |
| `text`    | (described [below](#simplified-text-outputs)) |
| `tsv`     | [Tab-Separated Values](https://en.wikipedia.org/wiki/Tab-separated_values) |
| `zeek`    | [Zeek Logs](https://docs.zeek.org/en/master/logs/index.html) |

The output format defaults to either Super JSON or Super Binary and may be specified
with the `-f` option.

Since Super JSON is a common format choice, the `-z` flag is a shortcut for
`-f jsup`.  Also, `-Z` is a shortcut for `-f jsup` with `-pretty 4` as
[described below](#pretty-printing).

And since plain JSON is another common format choice, the `-j` flag is a shortcut for
`-f json` and `-J` is a shortcut for pretty printing JSON.

### Output Format Selection

When the format is not specified with `-f`, it defaults to Super JSON if the output
is a terminal and to Super Binary otherwise.

While this can cause an occasional surprise (e.g., forgetting `-f` or `-z`
in a scripted test that works fine on the command line but fails in CI),
we felt that the design of having a uniform default had worse consequences:
* If the default format were Super JSON, it would be very easy to create pipelines
and deploy to production systems that were accidentally using Super JSON instead of
the much more efficient Super Binary format because the `-f bsup` had been mistakenly
omitted from some command.  The beauty of SuperDB is that all of this "just works"
but it would otherwise perform poorly.
* If the default format were Super Binary, then users would be endlessly annoyed by
binary output to their terminal when forgetting to type `-f jsup`.

In practice, we have found that the output defaults
"just do the right thing" almost all of the time.

### Pretty Printing

Super JSON and plain JSON text may be "pretty printed" with the `-pretty` option, which takes
the number of spaces to use for indentation.  As this is a common option,
the `-Z` option is a shortcut for `-f jsup -pretty 4` and `-J` is a shortcut
for `-f json -pretty 4`.

For example,
```mdtest-command
echo '{a:{b:1,c:[1,2]},d:"foo"}' | super -Z -
```
produces
```mdtest-output
{
    a: {
        b: 1,
        c: [
            1,
            2
        ]
    },
    d: "foo"
}
```
and
```mdtest-command
echo '{a:{b:1,c:[1,2]},d:"foo"}' | super -f jsup -pretty 2 -
```
produces
```mdtest-output
{
  a: {
    b: 1,
    c: [
      1,
      2
    ]
  },
  d: "foo"
}
```

When pretty printing, colorization is enabled by default when writing to a terminal,
and can be disabled with `-color false`.

### Pipeline-friendly Super Binary

Though it's a compressed format, Super Binary data is self-describing and stream-oriented
and thus is pipeline friendly.

Since data is self-describing you can simply take Super Binary output
of one command and pipe it to the input of another.  It doesn't matter if the value
sequence is scalars, complex types, or records.  There is no need to declare
or register schemas or "protos" with the downstream entities.

In particular, Super Binary data can simply be concatenated together, e.g.,
```mdtest-command
super -f bsup -c 'select value 1, [1,2,3]' > a.bsup
super -f bsup -c 'select value {s:"hello"}, {s:"world"}' > b.bsup
cat a.bsup b.bsup | super -z -
```
produces
```mdtest-output
1
[1,2,3]
{s:"hello"}
{s:"world"}
```
And while this Super JSON output is human readable, the Super Binary files are binary, e.g.,
```mdtest-command
super -f bsup -c 'select value 1,[ 1,2,3]' > a.bsup
hexdump -C a.bsup
```
produces
```mdtest-output
00000000  02 00 01 09 1b 00 09 02  02 1e 07 02 02 02 04 02  |................|
00000010  06 ff                                             |..|
00000012
```

### Schema-rigid Outputs

Certain data formats like [Arrow](https://arrow.apache.org/docs/format/Columnar.html#ipc-streaming-format)
and [Parquet](https://github.com/apache/parquet-format) are "schema rigid" in the sense that
they require a schema to be defined before values can be written into the file
and all the values in the file must conform to this schema.

SuperDB, however, has a fine-grained type system instead of schemas such that a sequence
of data values is completely self-describing and may be heterogeneous in nature.
This creates a challenge converting the type-flexible super-structured data formats to a schema-rigid
format like Arrow and Parquet.

For example, this seemingly simple conversion:
```mdtest-command fails
echo '{x:1}{s:"hello"}' | super -o out.parquet -f parquet -
```
causes this error
```mdtest-output
parquetio: encountered multiple types (consider 'fuse'): {x:int64} and {s:string}
```

#### Fusing Schemas

As suggested by the error above, the [`fuse` operator](../language/operators/fuse.md) can merge different record
types into a blended type, e.g., here we create the file and read it back:
```mdtest-command
echo '{x:1}{s:"hello"}' | super -o out.parquet -f parquet -c fuse -
super -z out.parquet
```
but the data was necessarily changed (by inserting nulls):
```mdtest-output
{x:1,s:null(string)}
{x:null(int64),s:"hello"}
```

#### Splitting Schemas

Another common approach to dealing with the schema-rigid limitation of Arrow and
Parquet is to create a separate file for each schema.

`super` can do this too with the `-split` option, which specifies a path
to a directory for the output files.  If the path is `.`, then files
are written to the current directory.

The files are named using the `-o` option as a prefix and the suffix is
`-<n>.<ext>` where the `<ext>` is determined from the output format and
where `<n>` is a unique integer for each distinct output file.

For example, the example above would produce two output files,
which can then be read separately to reproduce the original data, e.g.,
```mdtest-command
echo '{x:1}{s:"hello"}' | super -o out -split . -f parquet -
super -z out-*.parquet
```
produces the original data
```mdtest-output
{x:1}
{s:"hello"}
```

While the `-split` option is most useful for schema-rigid formats, it can
be used with any output format.

### Simplified Text Outputs

The `text` and `table` formats simplify data to fit within the
limitations of text-based output. Because they do not capture all the
information required to reconstruct the original data, they are not supported
input formats. They may be a good fit for use with other text-based shell
tools, but due to their limitations should be used with care.

In `text` output, minimal formatting is applied, e.g., strings are shown
without quotes and brackets are dropped from [arrays](../formats/zed.md#22-array)
and [sets](../formats/zed.md#23-set). [Records](../formats/zed.md#21-record)
are printed as tab-separated field values without their corresponding field
names. For example:

```mdtest-command
echo '"hi" {hello:"world",good:"bye"} [1,2,3]' | super -f text -
```
produces
```mdtest-output
hi
world	bye
1,2,3
```

The `table` format includes header lines showing the field names in records.
For example:

```mdtest-command
echo '{word:"one",digit:1} {word:"two",digit:2}' | super -f table -
```
produces
```mdtest-output
word digit
one  1
two  2
```

If a new record type is encountered in the input stream that does not match
the previously-printed header line, a new header line will be output.
For example:

```mdtest-command
echo '{word:"one",digit: 1} {word:"hello",style:"greeting"}' |
  super -f table -
```
produces
```mdtest-output
word digit
one  1
word  style
hello greeting
```

If this is undesirable, the [`fuse` operator](../language/operators/fuse.md)
may prove useful to unify the input stream under a single record type that can
be described with a single header line. Doing this to our last example, we find

```mdtest-command
echo '{word:"one",digit:1} {word:"hello",style:"greeting"}' |
  super -f table -c 'fuse' -
```
now produces
```mdtest-output
word  digit style
one   1     -
hello -     greeting
```

### SuperDB Data Lake Metadata Output

The `lake` format is used to pretty-print lake metadata, such as in
[`super db` sub-command](super-db.md) outputs.  Because it's `super db`'s default output format,
it's rare to request it explicitly via `-f`.  However, since it's possible for
`super db` to [generate output in any supported format](super-db.md#super-db-commands),
the `lake` format is useful to reverse this.

For example, imagine you'd executed a [meta-query](super-db.md#meta-queries) via
`super db query -Z "from :pools"` and saved the output in this file `pools.jsup`.

```mdtest-input pools.jsup
{
    ts: 2024-07-19T19:28:22.893089Z,
    name: "MyPool",
    id: 0x132870564f00de22d252b3438c656691c87842c2 (=ksuid.KSUID),
    layout: {
        order: "desc" (=order.Which),
        keys: [
            [
                "ts"
            ] (=field.Path)
        ] (=field.List)
    } (=order.SortKey),
    seek_stride: 65536,
    threshold: 524288000
} (=pools.Config)
```

Using `super -f lake`, this can be rendered in the same pretty-printed form as it
would have originally appeared in the output of `super db ls`, e.g.,

```mdtest-command
super -f lake pools.jsup
```
produces
```mdtest-output
MyPool 2jTi7n3sfiU7qTgPTAE1nwTUJ0M key ts order desc
```

## Query Debugging

If you are ever stumped about how the `super` compiler is parsing your query,
you can always run `super -C` to compile and display your query in canonical form
without running it.
This can be especially handy when you are learning the language and
[its shortcuts](../language/pipeline-model.md#implied-operators).

For example, this query
```mdtest-command
super -C -c 'has(foo)'
```
is an implied [`where` operator](../language/operators/where.md), which matches values
that have a field `foo`, i.e.,
```mdtest-output
where has(foo)
```
while this query
```mdtest-command
super -C -c 'a:=x+1'
```
is an implied [`put` operator](../language/operators/put.md), which creates a new field `a`
with the value `x+1`, i.e.,
```mdtest-output
put a:=x+1
```

## Error Handling

Fatal errors like "file not found" or "file system full" are reported
as soon as they happen and cause the `super` process to exit.

On the other hand,
runtime errors resulting from the query itself
do not halt execution.  Instead, these error conditions produce
[first-class errors](../language/data-types.md#first-class-errors)
in the data output stream interleaved with any valid results.
Such errors are easily queried with the
[`is_error` function](../language/functions/is_error.md).

This approach provides a robust technique for debugging complex queries,
where errors can be wrapped in one another providing stack-trace-like debugging
output alongside the output data.  This approach has emerged as a more powerful
alternative to the traditional technique of looking through logs for errors
or trying to debug a halted query with a vague error message.

For example, this query
```mdtest-command
echo '1 2 0 3' | super -z -c '10.0/this' -
```
produces
```mdtest-output
10.
5.
error("divide by zero")
3.3333333333333335
```
and
```mdtest-command
echo '1 2 0 3' | super -c '10.0/this' - | super -z -c 'is_error(this)' -
```
produces just
```mdtest-output
error("divide by zero")
```

## Examples

As you may have noticed, many examples of the [SuperSQL language](../language/README.md)
are illustrated using this pattern
```
echo <values> | super -c <query> -
```
which is used throughout the [language documentation](../language/README.md)
and [operator reference](../language/operators/README.md).

The language documentation and [tutorials directory](../tutorials/README.md)
have many examples, but here are a few more simple `super` use cases.

_Hello, world_
```mdtest-command
super -z -c "select value 'hello, world'"
```
produces this Super JSON output
```mdtest-output
"hello, world"
```

_Some values of available [data types](../language/data-types.md)_
```mdtest-command
echo '1 1.5 [1,"foo"] |["apple","banana"]|' | super -z -
```
produces
```mdtest-output
1
1.5
[1,"foo"]
|["apple","banana"]|
```
_The types of various data_
```mdtest-command
echo '1 1.5 [1,"foo"] |["apple","banana"]|' | super -z -c 'select value typeof(this)' -
```
produces
```mdtest-output
<int64>
<float64>
<[(int64,string)]>
<|[string]|>
```
_A simple [aggregation](../language/aggregates/README.md)_
```mdtest-command
echo '{key:"foo",val:1}{key:"bar",val:2}{key:"foo",val:3}' |
  super -z -c 'sum(val) by key | sort key' -
```
produces
```mdtest-output
{key:"bar",sum:2}
{key:"foo",sum:4}
```
_Read CSV input and [cast](../language/functions/cast.md) a to an integer from default float_
```mdtest-command
printf "a,b\n1,foo\n2,bar\n" | super -z -c 'a:=int64(a)' -
```
produces
```mdtest-output
{a:1,b:"foo"}
{a:2,b:"bar"}
```
_Read JSON input and cast to an integer from default float_
```mdtest-command
echo '{"a":1,"b":"foo"}{"a":2,"b":"bar"}' | super -z -c 'a:=int64(a)' -
```
produces
```mdtest-output
{a:1,b:"foo"}
{a:2,b:"bar"}
```
_Make a schema-rigid Parquet file using fuse, then output the Parquet file as Super JSON_
```mdtest-command
echo '{a:1}{a:2}{b:3}' | super -f parquet -o tmp.parquet -c fuse -
super -z tmp.parquet
```
produces
```mdtest-output
{a:1,b:null(int64)}
{a:2,b:null(int64)}
{a:null(int64),b:3}
```

## Performance

You might think that the overhead involved in managing super-structured types
and the generality of heterogeneous data would confound the performance of
the `super` command, but it turns out that `super` can hold its own when
compared to other analytics systems.

To illustrate comparative performance, we'll present some informal performance
measurements among `super`,
[`DuckDB`](https://duckdb.org/),
[`ClickHouse`](https://clickhouse.com/), and
[`DataFusion`](https://datafusion.apache.org/).

We'll use the Parquet format to compare apples to apples
and also report results for the custom columnar database format of DuckDB
and the Super Binary format used by `super`.
We tried loading our test data into a ClickHouse table using its
[new experimental JSON type](https://clickhouse.com/blog/a-new-powerful-json-data-type-for-clickhouse)
but those attempts failed with "too many open files".

As of this writing in November 2024, we're using the latest version 1.1.3 of `duckdb`.
version 24.11.1.1393 of `clickhouse`, and v43.0.0 of `datafusion-cli`.

### The Test Data

These tests are based on the data and exemplary queries
published by the DuckDB team on their blog
[Shredding Deeply Nested JSON, One Vector at a Time](https://duckdb.org/2023/03/03/json.html).  We'll follow their script starting at the
[GitHub Archive Examples](https://duckdb.org/2023/03/03/json.html#github-archive-examples).

If you want to reproduce these results for yourself,
you can fetch the 2.2GB of gzipped JSON data:
```
wget https://data.gharchive.org/2023-02-08-0.json.gz
wget https://data.gharchive.org/2023-02-08-1.json.gz
...
wget https://data.gharchive.org/2023-02-08-23.json.gz
```
We downloaded these files into a directory called `gharchive_gz`
and created a duckdb database file called `gha.db` and a table called `gha`
using this command:
```
duckdb gha.db -c "CREATE TABLE gha AS FROM read_json('gharchive_gz/*.json.gz', union_by_name=true)"
```
To create a relational table from the input JSON, we utilized DuckDB's
`union_by_name` parameter to fuse all of the different shapes of JSON encountered
into a single monolithic schema.

We then created a Parquet file called `gha.parquet` with this command:
```
duckdb gha.db -c "COPY (from gha) TO 'gha.parquet'"
```
To create a super-structed file for the `super` command, there is no need to
fuse the data into a single schema (though `super` can still work with the fused
schema in the Parquet file), and we simply ran this command to create a Super Binary
file:
```
super gharchive_gz/*.json.gz > gha.bsup
```
This code path in `super` is not multi-threaded so not particularly performant but,
on our test machine, it takes about the same time as the `duckdb` method of creating
a schema-fused table.

Here are the resulting file sizes:
```
% du -h gha.db gha.parquet gha.bsup gharchive_gz
9.3G	gha.db
4.6G	gha.parquet
2.8G	gha.bsup
2.2G	gharchive_gz
```

### The Test Queries

The test queries involve these patterns:
* simple search (single and multicolumn)
* count-where aggregation
* count by field aggregation
* rank over union of disparate field types

We will call these tests `search`, `search+`, `count`, `agg`, and `union`, respectively

#### Search

For the search test, we'll search for the string pattern
```
    "in case you have any feedback 😊"
```
in the field `payload.pull_request.body`
and we'll just count the number of matches found.
The number of matches is small (3) so the query performance is dominated
by the search.

The SQL for this query is
```sql
SELECT count()
FROM 'gha.parquet' -- or gha
WHERE payload.pull_request.body LIKE '%in case you have any feedback 😊%'
```
SuperSQL has a function called `grep` that is similar to the SQL `LIKE` clause but
can operate over specified fields or default to all the string fields in any value.
The SuperSQL query is
```sql
SELECT count()
FROM 'gha.bsup'
WHERE grep('in case you have any feedback 😊', payload.pull_request.body)
```

#### Search+

For search across multiple columns, SQL doesn't have a `grep` function so
we must enumerate all the fields of such a query.  The SQL for a string search
over our GitHub Archive dataset involves the following fields:
```sql
SELECT count() FROM gha
WHERE id LIKE '%in case you have any feedback 😊%'
  OR type LIKE '%in case you have any feedback 😊%'
  OR actor.login LIKE '%in case you have any feedback 😊%'
  OR actor.display_login LIKE '%in case you have any feedback 😊%'
  ...
  OR payload.member.type LIKE '%in case you have any feedback 😊%'
```
There are 486 such fields.  You can review the entire query in
[docs/commands/search.sql](search.sql).

#### Count

In the `count` test, we filter the input with a WHERE clause and count the results.
We chose a random GitHub user name for the filter.
This query has the form:
```
SELECT count()
FROM 'gha.parquet' -- or gha or 'gha.bsup'
WHERE actor.login='johnbieren'"
```

#### Agg

In the `agg` test, we filter the input and count the results grouped by the field `type`
as in the DuckDB blog.
This query has the form:
```
SELECT count(),type
FROM 'gha.parquet' -- or 'gha' or 'gha.bsup'
WHERE repo.name='duckdb/duckdb'
GROUP BY type
```

#### Union

The `union` test is straight out of the DuckDB blog at the end of
[this section](https://duckdb.org/2023/03/03/json.html#handling-inconsistent-json-schemas).
This query computes the GitHub users that were assigned as a PR reviewer the most often
and returns the top 5 such users.
Because the assignees can appear in either a list of strings
or within a single string field, the relational model requires that two different
subqueries run for the two cases and the result unioned together; then,
this intermediary table can be counted using the unnested
assignee as the group-by key.
This query is:
```sql
WITH assignees AS (
  SELECT payload.pull_request.assignee.login assignee
  FROM 'gha.parquet'
  UNION ALL
  SELECT unnest(payload.pull_request.assignees).login assignee
  FROM 'gha.parquet'
)
SELECT assignee, count(*) count
FROM assignees
WHERE assignee IS NOT NULL
GROUP BY assignee
ORDER BY count DESC
LIMIT 5
```
For DataFusion, we needed to rewrite this SELECT
```sql
SELECT unnest(payload.pull_request.assignees).login
FROM 'gha.parquet'
```
as
```sql
SELECT rec.login as assignee FROM (
    SELECT unnest(payload.pull_request.assignees) rec
    FROM 'gha.parquet'
)
```
and for ClickHouse, we had to use `arrayJoin` instead of `unnest`.

SuperSQL's data model does not require these sorts of gymnastics as
everything does not have to be jammed into a table.  Instead, we can use the
`UNNEST` pipe operator combined with the spread operator applied to the array of
string fields to easily produce a stream of string values representing the
assignees.  Then we simply aggregate the assignee stream:
```
FROM 'gha.bsup'
| UNNEST [...payload.pull_request.assignees, payload.pull_request.assignee]
| WHERE this IS NOT NULL
| AGGREGATE count() BY assignee:=login
| ORDER BY count DESC
| LIMIT 5
```

### The Test Results

The following table summarizes the results of each test as a column and
each tool as a row with the speed-up factor shown in parentheses:

| tool | format | search | search+ | count | agg | union |
|--------------|---------------|---------------|---------------|----|------|-------|
| `super` | `bsup` | 3.2 (2.6X) | 6.7 (3.6X) | 3.2 (0.04X) | 3.1 (0.04X) | 3.8 (117X) |
| `super` | `parquet` | note 1 | note 1  | 0.18 (0.7X) | 0.27 (0.4X) | note 2 |
| `duckdb` | `db` | 8.2  | 24  | 0.13 | 0.12  | 446  |
| `duckdb` | `parquet` | 8.4 (1) | 23 (1X)  | 0.26 (0.5X) | 0.21 (0.6X) | 419 (1.1X) |
| `datafusion` | `parquet` | 9.1 (0.9X) | 18 (1.3X)  | 0.24 (0.5X) | 0.24 (0.5X) | 40 (11x) |
| `clickhouse` | `parquet` | 56 (0.1X) | 463 (0.1X)  | 1 (0.1X) | 0.91 (0.1X) | 66 (7X) |

_Note 1: the `super` vectorized runtime does not yet support `grep`_

_Note 2: the `super` vectorized runtime does not yet support array expressions_

Since DuckDB with its native format is overall the best performing,
we used it as the baseline for all of the speedup factors.

To summarize,
`super` with Super Binary is substantially faster than the relational systems for
the search use cases and performs on par with the others for traditional OLAP queries,
except for the union query, where the super-structured data model trounces the relational
model (by over 100X!) for stitching together disparate data types for analysis in an aggregation.

## Appendix 1: Preparing the Test Data

For our tests, we diverged a bit from the methodology in the DuckDB blog and wanted
to put all the JSON data in a single table.  It wasn't obvious how to go about this
and this section documents the difficulties we encountered trying to do so.

First, we simply tried this:
```
duckdb gha.db -c "CREATE TABLE gha AS FROM 'gharchive_gz/*.json.gz'"
```
which fails with
```
Invalid Input Error: JSON transform error in file "gharchive_gz/2023-02-08-10.json.gz", in line 4903: Object {"url":"https://api.github.com/repos/aws/aws-sam-c... has unknown key "reactions"
Try increasing 'sample_size', reducing 'maximum_depth', specifying 'columns', 'format' or 'records' manually, setting 'ignore_errors' to true, or setting 'union_by_name' to true when reading multiple files with a different structure.
```
Clearly the schema inference algorithm relies upon sampling and the sample doesn't
cover enough data to capture all of its variations.

Okay, maybe there is a reason the blog first explores the structure of
the data to specify `columns` arguments to `read_json` as suggested by the error
message above.  To this end, you can run this query:
```
SELECT json_group_structure(json)
FROM (
  SELECT *
  FROM read_ndjson_objects('gharchive_gz/*.json.gz')
  LIMIT 2048
);
```
Unfortunately, if you use the resulting structure to create the `columns` argument
then `duckdb` fails also because the first 2048 records don't have enough coverage.
So let's try removing the `LIMIT` clause:
```
SELECT json_group_structure(json)
FROM (
  SELECT *
  FROM read_ndjson_objects('gharchive_gz/*.json.gz')
);
```
Hmm, now `duckdb` runs out of memory.

We then thought we'd see if the sampling algorithm of `read_json` is more efficient,
so we ran tried this command with successively larger sample sizes:
```
duckdb scratch -c "CREATE TABLE gha AS FROM read_json('gharchive_gz/*.json.gz', sample_size=1000000)"
```
even with a million rows as the sample, `duckdb` fails with
```
Invalid Input Error: JSON transform error in file "gharchive_gz/2023-02-08-14.json.gz", in line 49745: Object {"issues":"write","metadata":"read","pull_requests... has unknown key "repository_hooks"
Try increasing 'sample_size', reducing 'maximum_depth', specifying 'columns', 'format' or 'records' manually, setting 'ignore_errors' to true, or setting 'union_by_name' to true when reading multiple files with a different structure.
```
Ok, there 4434953 JSON objects in the input so let's try this:
```
duckdb gha.db -c "CREATE TABLE gha AS FROM read_json('gharchive_gz/*.json.gz', sample_size=4434953)"
```
and again `duckdb` runs out of memory.

So we looked at the other options suggested by the error message and
`union_by_name` appeared promising.  Enabling this option causes DuckDB
to combine all the JSON objects into a single fused schema.
Maybe this would work better?

Sure enough, this works:
```
duckdb gha.db -c "CREATE TABLE gha AS FROM read_json('gharchive_gz/*.json.gz', union_by_name=true)"
```
We now have the `duckdb` database file for our GitHub Archive data called `gha.db`
containing a single table called `gha` embedded in that database.
What about the super-structured
format for the `super` command?  There is no need to futz with sample sizes,
schema inference, or union by name, just run this to create a Super Binary file:
```
super gharchive_gz/*.json.gz > gha.bsup
```

## Appendix 2: Running the Tests

This appendix provides the raw tests and output that we run on a MacBook Pro to generate
the table of results above.

### Search Test

```
; time super -c "
  SELECT count()
  FROM 'gha.bsup'
  WHERE grep('in case you have any feedback 😊', payload.pull_request.body)
"
{count:2(uint64)}
super -c   12.70s user 0.69s system 415% cpu 3.223 total

time duckdb gha.db -c "
  SELECT count()
  FROM gha
  WHERE payload.pull_request.body LIKE '%in case you have any feedback 😊%'
"
┌──────────────┐
│ count_star() │
│    int64     │
├──────────────┤
│            2 │
└──────────────┘
duckdb gha.db -c   26.66s user 6.90s system 406% cpu 8.266 total

; time duckdb -c "
  SELECT count()
  FROM gha.parquet
  WHERE payload.pull_request.body LIKE '%in case you have any feedback 😊%'
"
┌──────────────┐
│ count_star() │
│    int64     │
├──────────────┤
│            2 │
└──────────────┘
duckdb -c   42.71s user 6.06s system 582% cpu 8.380 total

; time datafusion-cli -c "
  SELECT count()
  FROM 'gha.parquet'
  WHERE payload.pull_request.body LIKE '%in case you have any feedback 😊%'
"
DataFusion CLI v43.0.0
+---------+
| count() |
+---------+
| 2       |
+---------+
1 row(s) fetched.
Elapsed 8.819 seconds.

datafusion-cli -c   40.75s user 6.72s system 521% cpu 9.106 total

; time clickhouse -q "
  SELECT count()
  FROM 'gha.parquet'
  WHERE payload.pull_request.body LIKE '%in case you have any feedback 😊%'
"
2
clickhouse -q   50.81s user 1.83s system 94% cpu 55.994 total
```

### Search+ Test

```
; time super -c "
  SELECT count()
  FROM 'gha.bsup'
  WHERE grep('in case you have any feedback 😊')
"
{count:3(uint64)}
super -c   43.80s user 0.71s system 669% cpu 6.653 total

; time duckdb gha.db < search.sql
┌──────────────┐
│ count_star() │
│    int64     │
├──────────────┤
│            3 │
└──────────────┘
duckdb gha.db < search.sql  73.60s user 33.29s system 435% cpu 24.563 total

; time duckdb < search-parquet.sql
┌──────────────┐
│ count_star() │
│    int64     │
├──────────────┤
│            3 │
└──────────────┘
duckdb < search-parquet.sql  89.57s user 29.21s system 513% cpu 23.113 total

; time datafusion-cli -f search-parquet.sql
DataFusion CLI v43.0.0
+---------+
| count() |
+---------+
| 3       |
+---------+
1 row(s) fetched.
Elapsed 18.184 seconds.
datafusion-cli -f search-parquet.sql  83.84s user 11.13s system 513% cpu 18.494 total

; time clickhouse --queries-file search-parquet.sql
3
clickhouse --queries-file search-parquet.sql  515.68s user 5.50s system 112% cpu 7:43.37 total
```
### Count Test

```
; time super -c "
  SELECT count()
  FROM 'gha.bsup'
  WHERE actor.login='johnbieren'
"
{count:879(uint64)}
super -c   13.81s user 0.71s system 449% cpu 3.233 total

; time SUPER_VAM=1 super -c "
  SELECT count()
  FROM 'gha.parquet'
  WHERE actor.login='johnbieren'
"
{count:879(uint64)}
SUPER_VAM=1 super -c   0.43s user 0.08s system 277% cpu 0.182 total

; time duckdb gha.db -c "
  SELECT count()
  FROM gha
  WHERE actor.login='johnbieren'
"
┌──────────────┐
│ count_star() │
│    int64     │
├──────────────┤
│          879 │
└──────────────┘
duckdb gha.db -c   0.64s user 0.06s system 517% cpu 0.134 total

; time duckdb -c "
  SELECT count()
  FROM 'gha.parquet'
  WHERE actor.login='johnbieren'
"
┌──────────────┐
│ count_star() │
│    int64     │
├──────────────┤
│          879 │
└──────────────┘
duckdb gha.db -c   1.14s user 0.14s system 490% cpu 0.261 total

DataFusion CLI v43.0.0
+---------+
| count() |
+---------+
| 879     |
+---------+
1 row(s) fetched.
Elapsed 0.203 seconds.

datafusion-cli -c   0.93s user 0.15s system 453% cpu 0.238 total

; time clickhouse -q "
  SELECT count()
  FROM 'gha.parquet'
  WHERE actor.login='johnbieren'
"
879
clickhouse -q   0.86s user 0.07s system 93% cpu 1.001 total
```

### Agg Test

```
; time super -c "
  SELECT count(),type
  FROM 'gha.bsup'
  WHERE repo.name='duckdb/duckdb'
  GROUP BY type
"
{type:"PullRequestReviewEvent",count:14(uint64)}
{type:"IssueCommentEvent",count:30(uint64)}
{type:"WatchEvent",count:29(uint64)}
{type:"PullRequestEvent",count:35(uint64)}
{type:"PushEvent",count:15(uint64)}
{type:"IssuesEvent",count:9(uint64)}
{type:"ForkEvent",count:3(uint64)}
{type:"PullRequestReviewCommentEvent",count:7(uint64)}
super -c   12.24s user 0.68s system 413% cpu 3.129 total

; time SUPER_VAM=1 super -c "
  SELECT count(),type
  FROM 'gha.parquet'
  WHERE repo.name='duckdb/duckdb'
  GROUP BY type
"
{type:"IssueCommentEvent",count:30(uint64)}
{type:"PullRequestEvent",count:35(uint64)}
{type:"PushEvent",count:15(uint64)}
{type:"WatchEvent",count:29(uint64)}
{type:"PullRequestReviewEvent",count:14(uint64)}
{type:"ForkEvent",count:3(uint64)}
{type:"PullRequestReviewCommentEvent",count:7(uint64)}
{type:"IssuesEvent",count:9(uint64)}
SUPER_VAM=1 super -c   1.01s user 0.13s system 421% cpu 0.271 total

; time duckdb gha.db -c "
  SELECT count(),type
  FROM gha
  WHERE repo.name='duckdb/duckdb'
  GROUP BY type
"
┌──────────────┬───────────────────────────────┐
│ count_star() │             type              │
│    int64     │            varchar            │
├──────────────┼───────────────────────────────┤
│            3 │ ForkEvent                     │
│           35 │ PullRequestEvent              │
│           29 │ WatchEvent                    │
│            7 │ PullRequestReviewCommentEvent │
│           15 │ PushEvent                     │
│            9 │ IssuesEvent                   │
│           14 │ PullRequestReviewEvent        │
│           30 │ IssueCommentEvent             │
└──────────────┴───────────────────────────────┘
duckdb gha.db -c   0.49s user 0.06s system 466% cpu 0.119 total

; time duckdb -c "
  SELECT count(),type
  FROM 'gha.parquet'
  WHERE repo.name='duckdb/duckdb'
  GROUP BY type
"
┌──────────────┬───────────────────────────────┐
│ count_star() │             type              │
│    int64     │            varchar            │
├──────────────┼───────────────────────────────┤
│            9 │ IssuesEvent                   │
│            7 │ PullRequestReviewCommentEvent │
│           15 │ PushEvent                     │
│           14 │ PullRequestReviewEvent        │
│            3 │ ForkEvent                     │
│           29 │ WatchEvent                    │
│           35 │ PullRequestEvent              │
│           30 │ IssueCommentEvent             │
└──────────────┴───────────────────────────────┘
duckdb -c   0.73s user 0.14s system 413% cpu 0.211 total

; time datafusion-cli -c "
  SELECT count(),type
  FROM 'gha.parquet'
  WHERE repo.name='duckdb/duckdb'
  GROUP BY type
"
DataFusion CLI v43.0.0
+---------+-------------------------------+
| count() | type                          |
+---------+-------------------------------+
| 15      | PushEvent                     |
| 35      | PullRequestEvent              |
| 7       | PullRequestReviewCommentEvent |
| 14      | PullRequestReviewEvent        |
| 30      | IssueCommentEvent             |
| 9       | IssuesEvent                   |
| 29      | WatchEvent                    |
| 3       | ForkEvent                     |
+---------+-------------------------------+
8 row(s) fetched.
Elapsed 0.200 seconds.

datafusion-cli -c   0.80s user 0.15s system 398% cpu 0.238 total

; time clickhouse -q "
  SELECT count(),type
  FROM 'gha.parquet'
  WHERE repo.name='duckdb/duckdb'
  GROUP BY type
"
30	IssueCommentEvent
14	PullRequestReviewEvent
15	PushEvent
29	WatchEvent
9	IssuesEvent
7	PullRequestReviewCommentEvent
3	ForkEvent
35	PullRequestEvent
clickhouse -q   0.77s user 0.11s system 97% cpu 0.908 total
```

### Union Test

```
time super -c "
  FROM 'gha.bsup'
  | SELECT VALUE payload.pull_request
  | WHERE this IS NOT NULL
  | UNNEST [...assignees, assignee]
  | WHERE this IS NOT NULL
  | AGGREGATE count() BY assignee:=login
  | ORDER BY count DESC
  | LIMIT 5
"
{assignee:"poad",count:1966(uint64)}
{assignee:"vinayakkulkarni",count:508(uint64)}
{assignee:"tmtmtmtm",count:356(uint64)}
{assignee:"AMatutat",count:260(uint64)}
{assignee:"danwinship",count:208(uint64)}
super -c   12.39s user 0.95s system 351% cpu 3.797 total

; time duckdb gha.db -c "
  WITH assignees AS (
    SELECT payload.pull_request.assignee.login assignee
    FROM gha
    UNION ALL
    SELECT unnest(payload.pull_request.assignees).login assignee
    FROM gha
  )
  SELECT assignee, count(*) count
  FROM assignees
  WHERE assignee NOT NULL
  GROUP BY assignee
  ORDER BY count DESC
  LIMIT 5
"
┌─────────────────┬───────┐
│    assignee     │ count │
│     varchar     │ int64 │
├─────────────────┼───────┤
│ poad            │  1966 │
│ vinayakkulkarni │   508 │
│ tmtmtmtm        │   356 │
│ AMatutat        │   260 │
│ danwinship      │   208 │
└─────────────────┴───────┘
duckdb gha.db -c   3119.93s user 90.86s system 719% cpu 7:26.22 total

time duckdb -c "
  WITH assignees AS (
    SELECT payload.pull_request.assignee.login assignee
    FROM 'gha.parquet'
    UNION ALL
    SELECT unnest(payload.pull_request.assignees).login assignee
    FROM 'gha.parquet'
  )
  SELECT assignee, count(*) count
  FROM assignees
  WHERE assignee NOT NULL
  GROUP BY assignee
  ORDER BY count DESC
  LIMIT 5
"
┌─────────────────┬───────┐
│    assignee     │ count │
│     varchar     │ int64 │
├─────────────────┼───────┤
│ poad            │  1966 │
│ vinayakkulkarni │   508 │
│ tmtmtmtm        │   356 │
│ AMatutat        │   260 │
│ danwinship      │   208 │
└─────────────────┴───────┘
duckdb -c   2914.72s user 107.15s system 721% cpu 6:58.68 total

time datafusion-cli -c "
  WITH assignees AS (
    SELECT payload.pull_request.assignee.login assignee
    FROM 'gha.parquet'
    UNION ALL
    SELECT object.login as assignee FROM (
      SELECT unnest(payload.pull_request.assignees) object
      FROM 'gha.parquet'
    )
  )
  SELECT assignee, count() count
  FROM assignees
  WHERE assignee IS NOT NULL
  GROUP BY assignee
  ORDER BY count DESC
  LIMIT 5
"
DataFusion CLI v43.0.0
+-----------------+-------+
| assignee        | count |
+-----------------+-------+
| poad            | 1966  |
| vinayakkulkarni | 508   |
| tmtmtmtm        | 356   |
| AMatutat        | 260   |
| danwinship      | 208   |
+-----------------+-------+
5 row(s) fetched.
Elapsed 39.012 seconds.

datafusion-cli -c   116.97s user 44.50s system 408% cpu 39.533 total

; time clickhouse -q "
 WITH assignees AS (
    SELECT payload.pull_request.assignee.login assignee
    FROM 'gha.parquet'
    UNION ALL
    SELECT arrayJoin(payload.pull_request.assignees).login assignee
    FROM 'gha.parquet'
  )
  SELECT assignee, count(*) count
  FROM assignees
  WHERE assignee IS NOT NULL
  GROUP BY assignee
  ORDER BY count DESC
  LIMIT 5
"
poad	1966
vinayakkulkarni	508
tmtmtmtm	356
AMatutat	260
danwinship	208
clickhouse -q   105.49s user 6.54s system 169% cpu 1:06.27 total
```
