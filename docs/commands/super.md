---
sidebar_position: 1
sidebar_label: super
---

# `super`

> **TL;DR** `super` is a command-line tool that uses [SuperSQL](../language/README.md)
> to query a variety of data formats in files, over HTTP, or in [S3](../integrations/amazon-s3.md)
> storage. It is particularly fast when operating on data in binary formats such as
> [Super Binary](../formats/bsup.md), [Super Columnar](../formats/csup.md), and
> [Parquet](https://github.com/apache/parquet-format).
>
> The `super` design philosophy blends the command-line, embedded database
> approach of SQLite and DuckDB with the query/search-tool approach
> of `jq`, `awk`, and `grep`.

## Usage

```
super [ options ] [ -c query ] input [ input ... ]
```

`super` is a command-line tool for processing data in diverse input
formats, powering data wrangling, search, analytics, and extensive transformations
using the [SuperSQL language](../language/README.md). A SuperSQL query may be extended with
[pipe syntax](https://research.google/pubs/sql-has-problems-we-can-fix-them-pipe-syntax-in-sql/)
to apply Boolean logic or keyword search to filter the input, transform, and/or analyze
the filtered stream.  Output is written to one or more files or to
standard output.

Each `input` argument must be a file path, an HTTP or HTTPS URL,
an S3 URL, or standard input specified with `-`.

For built-in command help and a listing of all available options,
simply run `super` with no arguments.

`super` supports a number of [input](#input-formats) and [output](#output-formats) formats, but [Super Binary](../formats/bsup.md)
tends to be the most space-efficient and most performant.  Super Binary has efficiency similar to
[Avro](https://avro.apache.org)
and [Protocol Buffers](https://developers.google.com/protocol-buffers)
but its comprehensive [type system](../formats/zed.md) obviates
the need for schema specification or registries.
Also, the [Super JSON](../formats/jsup.md) format is human-readable and entirely one-to-one with Super Binary
so there is no need to represent non-readable formats like Avro or Protocol Buffers
in a clunky JSON encapsulated form.  

`super` typically operates on Super Binary-encoded data and when you want to inspect
human-readable bits of output, you merely format it as Super JSON, which is the
default format when output is directed to the terminal.  Super Binary is the default
when redirecting to a non-terminal output like a file or pipe.

When run with input arguments, each input's format is [automatically inferred](#auto-detection)
and each input is scanned
in the order appearing on the command line forming the input stream.

By invoking the `-c` option, a query expressed in the [SuperSQL language](../language/README.md)
may be specified and applied to the input stream.

If no query is specified, the inputs are scanned without modification
and output in the desired format as [described below](#input-formats).  This latter approach
provides a convenient means to convert files from one format to another.

When `super` is run with a query and no input arguments, then the query must
begin with
* a [`from`, `file`, or `get` operator](../language/operators/from.md), or
* an explicit or implied [`yield` operator](../language/operators/yield.md).

In the case of a `yield` with no inputs, the query is run with
a single input value of `null`.  This provides a convenient means to run in a
"calculator mode" where input is produced by the `yield` and can be operated upon
by the query, e.g.,
```mdtest-command
super -z -c '1+1'
```
emits
```mdtest-output
2
```
Note here that the query `1+1` [implies](../language/pipeline-model.md#implied-operators)
`yield 1+1`.

## Input Formats

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

Since [Super JSON](../formats/jsup.md) is a superset of plain JSON, `super` must be careful in whether it
interprets input as either format.  While you can always clarify your intent
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

## Output Formats

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
super -f bsup -c 'yield 1,[1,2,3]' > a.bsup
super -f bsup -c 'yield {s:"hello"},{s:"world"}' > b.bsup
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
super -f bsup -c 'yield 1,[1,2,3]' > a.bsup
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
[`super db` sub-command](zed.md) outputs.  Because it's `super db`'s default output format,
it's rare to request it explicitly via `-f`.  However, since it's possible for
`super db` to [generate output in any supported format](zed.md#zed-commands),
the `lake` format is useful to reverse this.

For example, imagine you'd executed a [meta-query](zed.md#meta-queries) via
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
echo '"hello, world"' | super -z -c 'yield this' -
```
produces this Super JSON output
```mdtest-output
"hello, world"
```

_Some values of available [data types](../language/data-types.md)_
```mdtest-command
echo '1 1.5 [1,"foo"] |["apple","banana"]|' | super -z -c 'yield this' -
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
echo '1 1.5 [1,"foo"] |["apple","banana"]|' | super -z -c 'yield typeof(this)' -
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

Your mileage may vary, but many new users of `super` are surprised by its speed
compared to tools like `jq`, `grep`, `awk`, or `sqlite` especially when running
`super` over files in the Super Binary format.

### Fast Pattern Matching

One important technique that helps `super` run fast is to take advantage of queries
that involve fine-grained searches.

When a query begins with a logical expression containing either a search
or a predicate match with a constant value, and presuming the input data format
is Super Binary, then the runtime optimizes the query by performing an efficient,
byte-oriented "pre-search" of the values required in the predicate.  This pre-search
scans the bytes that comprise a large buffer of values and looks for these values
and, if they are not present, the entire buffer is discarded knowing no individual
value in that buffer could match because the required serialized
values were not present in the buffer.

For example, if the query is
```
"http error" and ipsrc==10.0.0.1 | count()
```
then the pre-search would look for the string "http error" and the encoding
of the IP address 10.0.0.1 and unless both those values are present, then the
buffer is discarded.

Moreover, Super Binary data is compressed and arranged into frames that can be decompressed
and processed in parallel.  This allows the decompression and pre-search to
run in parallel very efficiently across a large number of threads.  When searching
for sparse results, many frames are discarded without their uncompressed bytes
having to be processed any further.

### Efficient JSON Processing

While processing data in the Super Binary format is far more efficient than JSON,
there is substantial JSON data in the world and it is important for JSON
input to perform well.

This proved a challenge as `super` is written in [Go](https://go.dev/) and Go's JSON package
is not particularly performant.  To this end, `super` has its own lean and simple
[JSON tokenizer](https://pkg.go.dev/github.com/brimdata/super/pkg/jsonlexer),
which performs quite well,
and is
[integrated tightly](https://github.com/brimdata/super/blob/main/zio/jsonio/reader.go)
with SuperDB's internal data representation.
Moreover, like `jq`,
`super`'s JSON parser does not require objects to be newline delimited and can
incrementally parse the input to minimize memory overhead and improve
processor cache performance.

The net effect is a JSON parser that is typically a bit faster than the
native C implementation in `jq`.

### Performance Comparisons

To provide a rough sense of the performance tradeoffs between `super` and
other tooling, this section provides results of a few simple speed tests.

#### Test Data

These tests are easy to reproduce.  The input data comes from a
[repository of sample security log data](https://github.com/brimdata/zed-sample-data),
where we used a semi-structured Zeek "conn" log from the `zeek-default` directory.

It is easy to convert the Zeek logs to a local Super Binary file using
`super`'s built-in [`get` operator](../language/operators/get.md):
```
super -o conn.bsup -c 'get https://raw.githubusercontent.com/brimdata/zed-sample-data/main/zeek-default/conn.log.gz'
```
This creates a new file `conn.bsup` from the Zeek log file fetched from GitHub.

Note that this data is a gzip'd file in the Zeek format and `super`'s auto-detector
figures out both that it is gzip'd and that the uncompressed format is Zeek.
There's no need to specify flags for this.

Next, a JSON file can be converted from Super Binary using:
```
super -f json conn.bsup > conn.json
```
Note here that we lose information in this conversion because the rich data types
of the [super data model](../formats/zed.md) (that were [translated from the Zeek format](../integrations/zeek/data-type-compatibility.md)) are lost.

We'll also make a SQLite database in the file `conn.db` as the table named `conn`.
One easy way to do this is to install
[sqlite-utils](https://sqlite-utils.datasette.io/en/stable/)
and run
```
sqlite-utils insert conn.db conn conn.json --nl
```
(If you need a cup of coffee, a good time to get it would be when
loading the JSON into SQLite.)

#### File Sizes

Note the resulting file sizes:
```
% du -h conn.json conn.db conn.bsup
416M	conn.json
192M	conn.db
 38M	conn.bsup
```
Much of the performance of Super Binary derives from an efficient, parallelizable
structure where frames of data are compressed
(currently with [LZ4](http://lz4.github.io/lz4/) though the
specification supports multiple algorithms) and the sequence of values
can be processed with only partial deserialization.

That said, there are quite a few more opportunities to further improve
the performance of `super` and the SuperDB system and we have a number of projects
forthcoming on this front.

#### Tests

We ran three styles of tests on a Mac quad-core 2.3GHz i7:
* `count` - compute the number of values present
* `search` - find a value in a field
* `agg` - sum a field grouped by another field

Each test was run for `jq`, `super` on JSON, `sqlite3`, and `super` on Super Binary.

We used the Bash `time` command to measure elapsed time.

The command lines for the `count` test were:
```
jq -s length conn.json
sqlite3 conn.db 'select count(*) from conn'
super -c 'count()' conn.bsup
super -c 'count()' conn.json
```
The command lines for the `search` test were:
```
jq 'select(.id.orig_h=="10.47.23.5")' conn.json
sqlite3 conn.db 'select * from conn where json_extract(id, "$.orig_h")=="10.47.23.5"'
super -c 'id.orig_h==10.47.23.5' conn.bsup
super -c 'id.orig_h==10.47.23.5' conn.json
```
Here, we look for an IP address (10.47.23.5) in a specific
field `id.orig_h` in the semi-structured data.  Note when using Super Binary,
the IP is a native type whereas for `jq` and SQLite it is a string.
Note that `sqlite` must use its `json_extract` function since nested JSON objects
are stored as minified JSON text.

The command lines for the `agg` test were:
```
jq -n -f agg.jq conn.json
sqlite3 conn.db 'select sum(orig_bytes),json_extract(id, "$.orig_h") as orig_h from conn group by orig_h'
super -c "sum(orig_bytes) by id.orig_h" conn.bsup
super -c "sum(orig_bytes) by id.orig_h" conn.json
```
where the `agg.jq` script is:
```
def adder(stream):
  reduce stream as $s ({}; .[$s.key] += $s.val);
adder(inputs | {key:.id.orig_h,val:.orig_bytes})
| to_entries[]
| {orig_h: (.key), sum: .value}
```

#### Results

The following table summarizes the results of each test as a column and
each tool as a row with the speed-up factor (relative to `jq`)
shown in parentheses:

|  | `count` | `search` | `agg` |
|------|---------------|---------------|---------------|
| `jq` | 11,540ms (1X) | 10,730ms (1X) | 20,175ms (1X) |
| `super-json` | 7,150ms (1.6X) | 7,230ms (1.5X)  | 7,390ms (2.7X) |
| `sqlite` | 100ms (115X) | 620ms (17X) | 1,475ms (14X) |
| `super-bsup` | 110ms (105X) | 135ms (80X) | 475ms (42X) |

To summarize, `super` with Super Binary is consistently fastest though `sqlite`
was a bit faster counting rows.

In particular, `super` is substantially faster (40-100X) than `jq` with the efficient
Super Binary format but more modestly faster (50-170%) when processing the bulky JSON input.
This is expected because parsing JSON becomes the bottleneck.

While SQLite is much faster than `jq`, it is not as fast as `super`.  The primary
reason for this is that SQLite stores its semi-structured columns as minified JSON text,
so it must scan and parse the JSON when executing the _where_ clause above
as well as the aggregated fields.

Also, note that the inferior performance of `sqlite` is in areas where databases
perform extraordinarily well if you do the work to
(1) transform semi-structured columns to relational columns by flattening
nested JSON objects (which are not indexable by `sqlite`) and
(2) configuring database indexes.

In fact, if you implement these changes, `sqlite` performs better than `super` on these tests.

However, the benefit of SuperDB is that no flattening is required.  And unlike `sqlite`,
`super` is not intended to be a database.  That said, there is no reason why database
performance techniques cannot be applied to the super data model and this is precisely what the
open-source SuperDB project intends to do.

Stay tuned!
