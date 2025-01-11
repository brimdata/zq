---
weight: 1
title: super
heading: super Tutorial
---

This tour provides new users of `super` an overview of the tool and
the [SuperPipe language](../language/_index.md)
by walking through a number of examples on the command-line.
This should get you started without having to read through all the gory details
of the [SuperPipe language](../language/_index.md) or [`super` command-line usage](../commands/super.md).

We'll start with some simple one-liners on the command line where we feed
some data to `super` with `echo` and specify `-` for `super` input to indicate
that standard input should be used, e.g.,
```
echo '"hello, world"' | super -
```
Then, toward the end of the tour, we'll experiment with some real-world GitHub data
pulled from the GitHub API.

If you want to follow along on the command line,
just make sure the `super` command [is installed](../install.md)
as well as [`jq`](https://stedolan.github.io/jq/).

## But JSON

While `super` is based on a new type of [data model](../formats/zed.md),
Zed just so happens to be a superset of JSON.

So if all you ever use `zq` for is manipulating JSON data,
it can serve you well as a handy, go-to tool.  In this way, `zq` is kind of
like `jq`.  As you probably know, `jq`
is a popular command-line tool for taking a sequence of JSON values as input,
doing interesting things on that input, and emitting results, of course, as JSON.

`jq` is awesome and powerful, but its syntax and computational model can
sometimes be daunting and difficult.  We tried to make `zq` really easy and intuitive,
and it is usually faster, sometimes [much faster](../commands/super.md#performance),
than `jq`.

To this end, if you want full JSON compatibility without having to delve into the
details of Zed, just use the `-j` option with `zq` and this will tell `zq` to
expect JSON values as input and produce JSON values as output, much like `jq`.

{{% tip "Tip" %}}

If your downstream JSON tooling expects only a single JSON value, we can use
`-j` along with [`collect()`](../language/aggregates/collect.md) to aggregate
multiple input values into an array. A `collect()` example is shown
[later in this tutorial](#running-analytics).

{{% /tip %}}

## `this` vs `.`

For example, to add 1 to some numbers with `jq`, you say:
```
echo '1 2 3' | jq '.+1'
```
and you get
```
2
3
4
```
With `zq`, the mysterious `jq` value `.` is instead called
the almost-as-mysterious value
[`this`](../language/pipeline-model.md#the-special-value-this) and you say:
```mdtest-command
echo '1 2 3' | super -z -c 'this+1' -
```
which also gives
```mdtest-output
2
3
4
```
> Note that we are using the `-z` option with `zq` in all of the examples,
> which causes `zq` to format the output as [ZSON](../formats/jsup.md).
> When running `zq` on the terminal, you do not need `-z` as it is the default,
> but we include it here for clarity and because all of these examples are
> run through automated testing, which is not attached to a terminal.

## Search vs Transformation

Unlike `jq`, which leads with _transformation_, `zq` leads with _search_ but
transformation is also pretty easy.  Let's show what we mean here with an
example.

If we run this `jq` command,
```
echo '1 2 3' | jq 2
```
we get
```
2
2
2
```
Hmm, that's a little odd, but it did what we told it to do.  In `jq`, the
expression `2` is evaluated for each input value, and the value `2`
is produced each time, so three copies of `2` are emitted.

In `zq` however, `2` by itself is interpreted as a search and is
[shorthand for](../language/pipeline-model.md#implied-operators) `search 2` so the command
```mdtest-command
echo '1 2 3' | super -z -c '? 2' -
```
produces this "search result":
```mdtest-output
2
```
In fact, this search syntax generalizes, and if we search over a more complex
input:
```mdtest-command
echo '1 2 [1,2,3] [4,5,6] {r:{x:1,y:2}} {r:{x:3,y:4}} "hello" "Number 2"' |
  super -z -c '? 2' -
```
we naturally find all the 2's whether as a value, inside a value, or inside a string:
```mdtest-output
2
[1,2,3]
{r:{x:1,y:2}}
"Number 2"
```
You can also do keyword-text search, e.g.,
```mdtest-command
echo '1 2 [1,2,3] [4,5,6] {r:{x:1,y:2}} {r:{x:3,y:4}} "hello" "Number 2"' |
  super -z -c '? hello or Number' -
```
produces
```mdtest-output
"hello"
"Number 2"
```
Doing searches like this in `jq` would be hard.

That said, we can emulate the `jq` transformation stance by explicitly
indicating that we want to [yield](../language/operators/yield.md)
the result of the expression evaluated for each input value, e.g.,
```mdtest-command
echo '1 2 3' | super -z -c 'yield 2' -
```
now gives the same answer as `jq`:
```mdtest-output
2
2
2
```
Cool, but doesn't it seem like search is a better disposition for
shorthand syntax?  What do you think?

## On to ZSON

JSON is super easy and ubiquitous, but it can be limiting and frustrating when
trying to do high-precision stuff with data.

When using `zq`, it's handy to operate in the
domain of Zed data and only output to JSON when needed.

The human-readable format of Zed is called [ZSON](../formats/jsup.md)
(and yes, that's a play on the acronym JSON).

ZSON is nice because it has a comprehensive type system and you can
go from ZSON to an efficient binary row format ([Super Binary](../formats/bsup.md))
and columnar ([Super Columnar](../formats/csup.md)) --- and vice versa ---
with complete fidelity and no loss of information.  In this tour,
we'll stick to ZSON (though for large data sets,
[Super Binary is much faster](../commands/super.md#performance)).

The first thing you'll notice about ZSON is that you don't need
quotations around field names.  We can see this by taking some JSON
as input (the JSON format is auto-detected by `zq`) and formatting
it as pretty-printed ZSON with `-Z`:
```mdtest-command
echo '{"s":"hello","val":1,"a":[1,2],"b":true}' | super -Z -
```
which gives
```mdtest-output
{
    s: "hello",
    val: 1,
    a: [
        1,
        2
    ],
    b: true
}
```
`s`, `val`, `a`, and `b` all appear as unquoted identifiers here.
Of course if you have funny characters in a field name, ZSON can handle
it with quotes just like JSON:
```mdtest-command
echo '{"funny@name":1}' | super -z -
```
produces
```mdtest-output
{"funny@name":1}
```
Moreover, ZSON is fully compatible with all of JSON's corner cases like empty string
as a field name and empty object as a value, e.g.,
```mdtest-command
echo '{"":{}}' | super -z -
```
produces
```mdtest-output
{"":{}}
```

## Comprehensive Types

ZSON also has a [comprehensive type system](../formats/zed.md).

For example, here is ZSON "record" with a taste of different types
of values as record fields:
```
{
    v1: 1.5,
    v2: 1,
    v3: 1 (uint8),
    v4: 2018-03-24T17:30:20.600852Z,
    v5: 2m30s,
    v6: 192.168.1.1,
    v7: 192.168.1.0/24,
    v8: [
        1,
        2,
        3
    ],
    v9: |[
        "GET",
        "PUT",
        "POST"
    ]|,
    v10: |{
        "key1": 123,
        "key2": 456
    }|,
    v11: {
        a: 1,
        r: {
            s1: "hello",
            s2: "world"
        }
    }
}
```
Here, `v1` is a 64-bit IEEE floating-point value just like JSON.

Unlike JSON, `v2` is a 64-bit integer.  And there are other integer
types as with `v3`,
which utilizes a [ZSON type decorator](../formats/jsup.md#22-type-decorators),
in this case,
to clarify its specific type of integer as unsigned 8 bits.
a
`v4` has type `time` and `v5` type `duration`.

`v6` is type `ip` and `v7` type `net`.

`v8` is an array of elements of type `int64`, which in Zed, is a type
written as `[int64]`.

`v9` is a "set of strings", which is written like an array but with the
enclosing syntax `|[` and `]|`.

`v10` is a "map" type, which in other languages is often called a "table"
or a "dictionary".  In Zed, a value of any type can be used for the key or the
value though all of the keys and all of the values must have the same type.

Finally, `v11` is a Zed "record", which is similar to a JSON "object", but the
keys are called "fields", the order of the fields is significant and
is always preserved.

## Records

As is often the case with semi-structured systems, you deal with
nested values all the time: in JSON, data is nested with objects and arrays,
while in Zed, data is nested with "records" and arrays (as well as other complex types).

[Record expressions](../language/expressions.md#record-expressions)
are rather flexible with `zq` and look a bit like JavaScript
or `jq` syntax, e.g.,
```mdtest-command
echo '1 2 3' | super -z -c 'yield {kind:"counter",val:this}' -
```
produces
```mdtest-output
{kind:"counter",val:1}
{kind:"counter",val:2}
{kind:"counter",val:3}
```
Note that like the search shortcut, you can also drop the yield keyword
here because the record literal implies the yield operator, e.g.,
```mdtest-command
echo '1 2 3' | super -z -c '{kind:"counter",val:this}' -
```
also produces
```mdtest-output
{kind:"counter",val:1}
{kind:"counter",val:2}
{kind:"counter",val:3}
```
`zq` can also use a spread operator like JavaScript, e.g.,
```mdtest-command
echo '{a:{s:"foo", val:1}}{b:{s:"bar"}}' | super -z -c '{...a,s:"baz"}' -
```
produces
```mdtest-output
{s:"baz",val:1}
{s:"baz"}
```
while
```mdtest-command
echo '{a:{s:"foo", val:1}}{b:{s:"bar"}}' | super -z -c '{d:2,...a,...b}' -
```
produces
```mdtest-output
{d:2,s:"foo",val:1}
{d:2,s:"bar"}
```

## Record Mutation

Sometimes you just want to extract or mutate certain fields of records.

Similar to the Unix `cut` command, the Zed [cut operator](../language/operators/cut.md)
extracts fields, e.g.,
```mdtest-command
echo '{s:"foo", val:1}{s:"bar"}' | super -z -c 'cut s' -
```
produces
```mdtest-output
{s:"foo"}
{s:"bar"}
```
while the [put operator](../language/operators/put.md) mutates existing fields
or adds new fields, e.g.,
```mdtest-command
echo '{s:"foo", val:1}{s:"bar"}' | super -z -c 'put val:=123,pi:=3.14' -
```
produces
```mdtest-output
{s:"foo",val:123,pi:3.14}
{s:"bar",val:123,pi:3.14}
```
Note that `put` is also an implied operator so the command with `put` omitted
```mdtest-command
echo '{s:"foo", val:1}{s:"bar"}' | super -z -c 'val:=123,pi:=3.14' -
```
produces the very same output:
```mdtest-output
{s:"foo",val:123,pi:3.14}
{s:"bar",val:123,pi:3.14}
```
Finally, it's worth mentioning that errors in Zed are
[first class](https://en.wikipedia.org/wiki/First-class_citizen).
This means they can just show up in the data as values.  In particular,
a common error is `error("missing")` which occurs most often when referencing
a field that does not exist, e.g.,
```mdtest-command
echo '{s:"foo", val:1}{s:"bar"}' | super -z -c 'cut val' -
```
produces
```mdtest-output
{val:1}
{val:error("missing")}
```
Sometimes you expect missing errors to occur sporadically and just want
to ignore them, which can you easily do with the
[quiet function](../language/functions/quiet.md), e.g.,
```mdtest-command
echo '{s:"foo", val:1}{s:"bar"}' | super -z -c 'cut quiet(val)' -
```
produces
```mdtest-output
{val:1}
```

## Union Types

One of the tricks `zq` uses to represent JSON data in its structured type system
is [union types](../language/expressions.md#union-values).
Most of the time, you don't need to worry about unions
but they show up from time to time.  Even when
they show up, Zed just tries to "do the right thing" so you usually
don't have to worry about them even when they show up.

For example, this query is perfectly happy to operate on the union values
that are implied by a mixed-type array:
```mdtest-command
echo '[1, "foo", 2, "bar"]' | super -z -c 'yield this[2],this[1]' -
```
produces
```mdtest-output
2
"foo"
```
but under the covers, the elements of the array have a union type of
`int64` and `string`, which is written `(int64,string)`, e.g,.
```mdtest-command
echo '[1, "foo", 2, "bar"]' | super -z -c 'yield typeof(this)' -
```
produces
```mdtest-output
<[(int64,string)]>
```
which is a type value representing an array of union values.

As you learn more about Zed and want to use `zq` to do data discovery and
preparation, union types are really quite powerful.  They allow records
with fields of different types or mixed-type arrays to be easily expressed
while also having a very precise type definition.  This is the essence
of Zed's new
[super-structured data model](../formats/_index.md#2-a-super-structured-pattern).

## First-class Types

Note that in the type value above, the type is wrapped in angle brackets.
This is how ZSON represents types when expressed as values.
In other words, Zed has
[first-class](https://en.wikipedia.org/wiki/First-class_citizen) types.

The type of any value in `zq` can be accessed via the
[typeof function](../language/functions/typeof.md), e.g.,
```mdtest-command
echo '1 "foo" 10.0.0.1' | super -z -c 'yield typeof(this)' -
```
produces
```mdtest-output
<int64>
<string>
<ip>
```
What's the big deal here?  We can print out the type of something.  Yawn.

Au contraire, this is really quite powerful because we can
use types as values to functions, e.g., as a dynamic argument to
the [cast function](../language/functions/cast.md):
```mdtest-command
echo '{a:0,b:"2"}{a:0,b:"3"}' | super -z -c 'yield cast(b, typeof(a))' -
```
produces
```mdtest-output
2
3
```
But more powerfully, types can be used anywhere a value can be used and
in particular, they can be group-by keys, e.g.,
```mdtest-command
echo '{x:1,y:2}{s:"foo"}{x:3,y:4}' |
  super -f table -c "count() by \`shape\`:=typeof(this) |> sort count" -
```
produces
```mdtest-output
shape               count
<{s:string}>        1
<{x:int64,y:int64}> 2
```
When run over large data sets, this gives you an insightful count of
each "shape" of data in the input.  This is a powerful building block for
data discovery.

It's worth mentioning `jq` also has a type operator, but it produces a
simple string instead of first-class types, and arrays and objects have
no detail about their structure, e,g.,
```
echo '1 true [1,2,3] {"s":"foo"}' | jq type
```
produces
```
"number"
"boolean"
"array"
"object"
```
Moreover, if we compare types of different objects
```
echo '{"a":{"s":"foo"},"b":{"x":1,"y":2}}' | jq '(.a|type)==(.b|type)'
```
we get "object" here for each type and thus the result:
```
true
```
i.e., they match even though their underlying shape is different.

With `zq` of course, these are different super-structured types so
the result is false, e.g.,
```mdtest-command
echo '{"a":{"s":"foo"},"b":{"x":1,"y":2}}' |
  super -z -c 'yield typeof(a)==typeof(b)' -
```
produces
```mdtest-output
false
```

## Sample

Sometimes you'd like to see a sample value of each shape, not its type.
This is easy to do with the [any aggregate function](../language/aggregates/any.md),
e.g,
```mdtest-command
echo '{x:1,y:2}{s:"foo"}{x:3,y:4}' |
  super -z -c 'val:=any(this) by typeof(this) |> sort val |> yield val' -
```
produces
```mdtest-output
{s:"foo"}
{x:1,y:2}
```
We like this pattern so much there is a shortcut [sample operator](../language/operators/sample.md), e.g.,
```mdtest-command
echo '{x:1,y:2}{s:"foo"}{x:3,y:4}' | super -z -c 'sample this |> sort this' -
```
emits the same result:
```mdtest-output
{s:"foo"}
{x:1,y:2}
```

## Fuse

Sometimes JSON data can get really messy with lots of variations in fields,
with null values appearing sometimes and sometimes not, and with the same
fields having different data types.  Most annoyingly, when you see a JSON object
like this in isolation:
```
{a:1,b:null}
```
you have no idea what the expected data type of `b` will be.  Maybe it's another
number?  Or maybe a string?  Or maybe an array or an embedded object?

`zq` and ZSON don't have this problem because every value (even `null`) is
comprehensively typed.  However, `zq` in fact must deal with this thorny problem
when reading JSON and converting it to Zed's super-structure.

This is where you might have to spend a little bit of time coding up
the right `zq` logic to disentangle a JSON mess. But once the data is cleaned up,
you can leave it in a Zed format and not worry again.

To do so, the [fuse operator](../language/operators/fuse.md) comes in handy.
Let's say you have this sequence of data:
```
{a:1,b:null}
{a:null,b:[2,3,4]}
```
As we said,
you can't tell by looking at either value what the types of both `a` and `b`
should be.  But if you merge the values into a common type, things begin to make
sense, e.g.,
```mdtest-command
echo '{a:1,b:null}{a:null,b:[2,3,4]}' | super -z -c fuse -
```
produces this transformed and comprehensively-typed ZSON output:
```mdtest-output
{a:1,b:null([int64])}
{a:null(int64),b:[2,3,4]}
```
Now you can see all the detail.

This turns out to be so useful, especially with large amounts of messy input data,
you will often find yourself fusing data then sampling it, e.g.,
```mdtest-command
echo '{a:1,b:null}{a:null,b:[2,3,4]}' | super -Z -c 'fuse |> sample' -
```
produces a comprehensively-typed sample:
```mdtest-output
{
    a: 1,
    b: null ([int64])
}
```
As you explore data in this fashion, you will often type various searches
to slice and dice the data as you get a feel for it all while sending
your interactive search results to `fuse |> sample`.

To appreciate all this, let's have a look next at some real-world data...

## Real-world GitHub Data

Now that we've covered the basics of `zq` and the Zed language, let's
use the query patterns from above to explore some GitHub data.

First, we need to grab the data.  You can use `curl` for this or you can
just use `zq` as `zq` can take URLs in addition to file name arguments.
This command will grab descriptions of first 30 PRs created in the
public `zed` repository and place it in a file called `prs.json`:
```
super -f json \
  https://api.github.com/repos/brimdata/super/pulls\?state\=all\&sort\=desc\&per_page=30 \
  > prs.json
```
Now that you have this JSON file on your local file system, how would you query it
with `zq`?

### Data Discovery

Before you can do anything, you need to know its structure but you generally don't
know anything after pulling some random data from an API.

So, let's poke around a bit and figure it out.  This process of data introspection
is often called _data discovery_.

You could start by using `jq` to pretty-print the JSON data,
```
jq . prs.json
```
That's 10,592 lines.  Ugh, quite a challenge to sift through.

Instead, let's start out by figuring out how many values are in the input, e.g.,
```mdtest-command dir=docs/tutorials
super -f text -c 'count()' prs.json
```
produces
```mdtest-output
1
```
Hmm, there's just one value.  It's probably a big JSON array but let's check with
the [kind function](../language/functions/kind.md), and as expected:
```mdtest-command dir=docs/tutorials
super -z -c 'kind(this)' prs.json
```
produces
```mdtest-output
"array"
```
Ok got it.  But, how many items are in the array?
```mdtest-command dir=docs/tutorials
super -z -c 'len(this)' prs.json
```
produces
```mdtest-output
30
```
Of course!  We asked GitHub to return 30 items and the API returns the
pull-request objects as elements of one array representing a single JSON value.

Let's see what sorts of things are in this array.  Here, we need to enumerate
the items from the array and do something with them.  So how about we use
the [over operator](../language/operators/over.md)
to traverse the array and count the array items by their "kind",
```mdtest-command dir=docs/tutorials
super -z -c 'over this |> count() by kind(this)' prs.json
```
produces
```mdtest-output
{kind:"record",count:30(uint64)}
```
Ok, they're all records.  Good, this should be easy!

The Zed records were all originally JSON objects.
Maybe we can just use "sample" to have a deeper look...
```
super -Z -c 'over this |> sample' prs.json
```
> Here we are using `-Z`, which is like `-z`, but instead of formatting each
> ZSON value on its own line, it pretty-prints the ZSON with vertical
> formatting like `jq` does for JSON.

Ugh, that output is still pretty big.  It's not 10k lines but it's still
more than 700 lines of pretty-printed ZSON.

Ok, maybe it's not so bad.  Let's check how many shapes there are with `sample`...
```mdtest-command dir=docs/tutorials
super -z -c 'over this |> sample |> count()' prs.json
```
produces
```mdtest-output
3(uint64)
```
All that data across the samples and only three shapes.
They must each be really big.  Let's check that out.

We can use the [len function](../language/functions/len.md) on the records to
see the size of each of the four records:
```mdtest-command dir=docs/tutorials
super -z -c 'over this |> sample |> len(this) |> sort this' prs.json
```
and we get
```mdtest-output
0
36
36
```
Ok, this isn't so bad... two shapes each have 36 fields but one is length zero?!
That outlier could only be the empty record.  Let's check:
```mdtest-command dir=docs/tutorials
super -z -c 'over this |> sample |> len(this)==0' prs.json
```
produces
```mdtest-output
{}
```
Sure enough, there it is.  We could also double check with `jq` that there are
blank records in the GitHub results, and sure enough
```
jq '.[] | select(length==0)' prs.json
```
produces
```
{}
{}
```
Try opening your editor on that JSON file to look for the empty objects.
Who knows why they are there?  No fun. Real-world data is messy.

How about we fuse the 3 shapes together and have a look at the result:
```
super -Z -c 'over this |> fuse |> sample' prs.json
```
We won't display the result here as it's still pretty big.  But you can
give it a try.  It's 379 lines.

But let's break down what's taking up all this space.

We can take the output from `fuse |> sample` and list the fields with
and their "kind".  Note that when we do an `over this` with records as
input, we get a new record value for each field structured as a key/value pair:
```mdtest-command dir=docs/tutorials
super -f table -c '
  over this
  |> fuse
  |> sample
  |> over this
  |> {field:key[0],kind:kind(value)}
' prs.json
```
produces
```mdtest-output
field               kind
url                 primitive
id                  primitive
node_id             primitive
html_url            primitive
diff_url            primitive
patch_url           primitive
issue_url           primitive
number              primitive
state               primitive
locked              primitive
title               primitive
user                record
body                primitive
created_at          primitive
updated_at          primitive
closed_at           primitive
merged_at           primitive
merge_commit_sha    primitive
assignee            primitive
assignees           array
requested_reviewers array
requested_teams     array
labels              array
milestone           primitive
draft               primitive
commits_url         primitive
review_comments_url primitive
review_comment_url  primitive
comments_url        primitive
statuses_url        primitive
head                record
base                record
_links              record
author_association  primitive
auto_merge          primitive
active_lock_reason  primitive
```
With this list of top-level fields, we can easily explore the different
pieces of their structure with sample.  Let's have a look at a few of the
record fields by giving these one-liners each a try and looking at the output:
```
super -Z -c 'over this |> sample head' prs.json
super -Z -c 'over this |> sample base' prs.json
super -Z -c 'over this |> sample _links' prs.json
```
While these fields have some useful information, we'll decide to drop them here
and focus on other top-level fields.  To do this, we can use the
[drop operator](../language/operators/drop.md) to whittle down the data:
```
super -Z -c 'over this |> fuse |> drop head,base,_link |> sample' prs.json
```
Ok, this looks more reasonable and is now only 120 lines of pretty-printed ZSON.

One more annoying detail here about JSON: time values are stored as strings,
in this case, in ISO format, e.g., we can pull this value out with
this query:
```mdtest-command dir=docs/tutorials
super -z -c 'over this |> head 1 |> yield created_at' prs.json
```
which produces this string:
```mdtest-output
"2019-11-11T19:50:46Z"
```
Since Zed has a native `time` type and we might want to do native date comparisons
on these time fields, we can easily translate the string to a time with a cast, e.g.,
```mdtest-command dir=docs/tutorials
super -z -c 'over this |> head 1 |> yield time(created_at)' prs.json
```
produces the native time value:
```mdtest-output
2019-11-11T19:50:46Z
```
To be sure, you can check any value's type with the `typeof` function, e.g.,
```mdtest-command dir=docs/tutorials
super -z -c 'over this |> head 1 |> yield time(created_at) |> typeof(this)' prs.json
```
produces the native time value:
```mdtest-output
<time>
```

### Cleaning up the Messy JSON

Okay, now that we've explored the data, we have a sense of it and can
"clean it up" with some Zed logic.  We'll do this one step at a time,
then put it all together.

First, let's get rid of the outer array and generate elements of an array
as a sequence of Zed records that have been fused and let's filter out
the empty records:
```
super -c 'over this |> len(this) != 0 |> fuse' prs.json > prs1.bsup
```
We can check that worked with count:
```
super -z -c 'count()' prs1.bsup
super -z -c 'sample |> count()' prs1.bsup
```
produces
```
{count:28(uint64)}
{count:1(uint64)}
```
Okay, good.  There are 28 values (the 30 requested less the two empty records)
and exactly one shape since the data was fused.

Now, let's drop the fields we aren't interested in:
```
super -c 'drop head,base,_links' prs1.bsup > prs2.bsup
```
Finally, let's clean up those dates.  To track down all the candidates,
we can run this Zed to group field names by their type and limit the output
to primitive types:
```
super -z -c '
  over this
  |> kind(value)=="primitive"
  |> fields:=union(key[0]) by type:=typeof(value)
' prs2.bsup
```
which gives
```
{type:<string>,fields:|["url","body","state","title","node_id","diff_url","html_url","closed_at","issue_url","merged_at","patch_url","created_at","updated_at","commits_url","comments_url","statuses_url","merge_commit_sha","author_association","review_comment_url","review_comments_url"]|}
{type:<int64>,fields:|["id","number"]|}
{type:<bool>,fields:|["draft","locked"]|}
{type:<null>,fields:|["assignee","milestone","auto_merge","active_lock_reason"]|}
```
> Note that this use of `over` traverses each record and generates a key-value pair
> for each field in each record.

Looking through the fields that are strings, the candidates for ISO dates appear
to be
* `closed_at`,
* `merged_at`,
* `created_at`, and
* `updated_at`.
You can do a quick check of the theory by running...
```
super -z -c '{closed_at,merged_at,created_at,updated_at}' prs2.bsup
```
and you will get strings that are all ISO dates:
```
{closed_at:"2019-11-11T20:00:22Z",merged_at:"2019-11-11T20:00:22Z",created_at:"2019-11-11T19:50:46Z",updated_at:"2019-11-11T20:00:25Z"}
{closed_at:"2019-11-11T21:00:15Z",merged_at:"2019-11-11T21:00:15Z",created_at:"2019-11-11T20:57:12Z",updated_at:"2019-11-11T21:00:26Z"}
...
```
To fix those strings, we simply transform the fields in place using the
(implied) [put operator](../language/operators/put.md) and redirect the final
output the ZNG file `prs.bsup`:
```
super -c '
  closed_at:=time(closed_at),
  merged_at:=time(merged_at),
  created_at:=time(created_at),
  updated_at:=time(updated_at)
' prs2.bsup > prs.bsup
```
We can check the result with our type analysis:
```mdtest-command dir=docs/tutorials
super -z -c '
  over this
  |> kind(value)=="primitive"
  |> fields:=union(key[0]) by type:=typeof(value)
  |> sort type
' prs.bsup
```
which now gives:
```mdtest-output
{type:<int64>,fields:|["id","number"]|}
{type:<time>,fields:|["closed_at","merged_at","created_at","updated_at"]|}
{type:<bool>,fields:|["draft","locked"]|}
{type:<string>,fields:|["url","body","state","title","node_id","diff_url","html_url","issue_url","patch_url","commits_url","comments_url","statuses_url","merge_commit_sha","author_association","review_comment_url","review_comments_url"]|}
{type:<null>,fields:|["assignee","milestone","auto_merge","active_lock_reason"]|}
```
and we can see that the date fields are correctly typed as type `time`!

> Note that we sorted the output values here using the [sort operator](../language/operators/sort.md)
> to produce a consistent output order since aggregations can be run in parallel
> to achieve scale and do not guarantee their output order.

## Putting It All Together

Instead of running each step above into a temporary file, we can
put all the transformations together in a single
Zed pipeline, where the Zed source text might look like this:
```
over this                      // traverse the array of objects
|> len(this) != 0               // skip empty objects
|> fuse                         // fuse objects into records of a combined type
|> drop head,base,_links        // drop fields that we don't need
|> closed_at:=time(closed_at),  // transform string dates to type time
  merged_at:=time(merged_at),
  created_at:=time(created_at),
  updated_at:=time(updated_at)
```
> Note that the `//` syntax indicates a single-line comment.

We can then put this in a file, called say `transform.zed`, and use the `-I`
argument to run all the transformations in one fell swoop:
```
super -I transform.zed prs.json > prs.bsup
```

## Running Analytics

Now that we've cleaned up our data, we can reliably and easily run analytics
on the finalized ZNG file `prs.bsup`.

Zed gives us the best of both worlds of JSON and relational tables: we have
the structure and clarity of the relational model while retaining the flexibility
of JSON's document model.  No need to create tables then issue SQL insert commands
to put your clean data into all the right places.

Let's start with something simple.  How about we output a "PR Report" listing
the title of each PR along with its PR number and creation date:
```mdtest-command dir=docs/tutorials
super -f table -c '{DATE:created_at,NUMBER:f"PR #{number}",TITLE:title}' prs.bsup
```
and you'll see this output...
```mdtest-output head
DATE                 NUMBER TITLE
2019-11-11T19:50:46Z PR #1  Make "make" work in zq
2019-11-11T20:57:12Z PR #2  fix install target
2019-11-11T23:24:00Z PR #3  import github.com/looky-cloud/lookytalk
2019-11-12T16:25:46Z PR #5  Make zq -f work
2019-11-12T16:49:07Z PR #6  a few clarifications to the zson spec
...
```
Note that we used a [formatted string literal](../language/expressions.md#formatted-string-literals)
to convert the field `number` into a string and format it with surrounding text.

Instead of old PRs, we can get the latest list of PRs using the
[tail operator](../language/operators/tail.md) since we know the data is sorted
chronologically. This command retrieves the last five PRs in the dataset:
```mdtest-command dir=docs/tutorials
super -f table -c '
  tail 5
  |> {DATE:created_at,"NUMBER":f"PR #{number}",TITLE:title}
' prs.bsup
```
and the output is:
```mdtest-output
DATE                 NUMBER TITLE
2019-11-18T22:14:08Z PR #26 ndjson writer
2019-11-18T22:43:07Z PR #27 Add reader for ndjson input
2019-11-19T00:11:46Z PR #28 fix TS_ISO8601, TS_MILLIS handling in NewRawAndTsFromJSON
2019-11-19T21:14:46Z PR #29 Return count of "dropped" fields from zson.NewRawAndTsFromJSON
2019-11-20T00:36:30Z PR #30 zval.sizeBytes incorrect
```

How about some aggregations?  We can count the number of PRs and sort by the
count highest first:
```mdtest-command dir=docs/tutorials
super -z -c "count() by user:=user.login |> sort count desc" prs.bsup
```
produces
```mdtest-output
{user:"mattnibs",count:10(uint64)}
{user:"aswan",count:7(uint64)}
{user:"mccanne",count:6(uint64)}
{user:"nwt",count:4(uint64)}
{user:"henridf",count:1(uint64)}
```
How about getting a list of all of the reviewers?  To do this, we need to
traverse the records in the `requested_reviewers` array and collect up
the login field from each record:
```mdtest-command dir=docs/tutorials
super -z -c 'over requested_reviewers |> collect(login)' prs.bsup
```
Oops, this gives us an array of the reviewer logins
with repetitions since [collect](../language/aggregates/collect.md)
collects each item that it encounters into an array:
```mdtest-output
["mccanne","nwt","henridf","mccanne","nwt","mccanne","mattnibs","henridf","mccanne","mattnibs","henridf","mccanne","mattnibs","henridf","mccanne","nwt","aswan","henridf","mccanne","nwt","aswan","philrz","mccanne","mccanne","aswan","henridf","aswan","mccanne","nwt","aswan","mikesbrown","henridf","aswan","mattnibs","henridf","mccanne","aswan","nwt","henridf","mattnibs","aswan","aswan","mattnibs","aswan","henridf","aswan","henridf","mccanne","aswan","aswan","mccanne","nwt","aswan","henridf","aswan"]
```
What we'd prefer is a set of reviewers where each reviewer appears only once.  This
is easily done with the [union](../language/aggregates/union.md) aggregate function
(not to be confused with union types) which
computes the set-wise union of its input and produces a Zed `set` type as its
output.  In this case, the output is a set of strings, written `|[string]|`
in the Zed language.  For example:
```mdtest-command dir=docs/tutorials
super -z -c 'over requested_reviewers |> reviewers:=union(login)' prs.bsup
```
produces
```mdtest-output
{reviewers:|["nwt","aswan","philrz","henridf","mccanne","mattnibs","mikesbrown"]|}
```
Ok, that's pretty neat.

Let's close with an analysis that's a bit more sophisticated.  Suppose we want
to look at the reviewers that each user tends to ask for.  We can think about
this question as a "graph problem" where the user requesting reviews is one node
in the graph and each set of reviewers is another node.

So as a first step, let's figure out how to create each edge, where an edge
is a relation between the requesting user and the set of reviewers.  We can
create this in Zed with a ["lateral subquery"](../language/lateral-subqueries.md).
Instead of computing a set-union over all the reviewers across all PRs,
we instead want to compute the set-union over the reviewers in each PR.
We can do this as follows:
```mdtest-command dir=docs/tutorials
super -z -c 'over requested_reviewers => ( reviewers:=union(login) )' prs.bsup
```
which produces an output like this:
```mdtest-output head
{reviewers:|["nwt","mccanne"]|}
{reviewers:|["nwt","henridf","mccanne"]|}
{reviewers:|["mccanne","mattnibs"]|}
{reviewers:|["henridf","mccanne","mattnibs"]|}
{reviewers:|["henridf","mccanne","mattnibs"]|}
...
```
Note that the syntax `=> ( ... )` defines a [lateral scope](../language/lateral-subqueries.md#lateral-scope) where any Zed subquery can
run in isolation over the input values created from the sequence of values
traversed by the outer `over`.

But we need a "graph edge" between the requesting user and the reviewers.
To do this, we need to reference the `user.login` from the top-level scope within the
lateral scope.  This can be done by
bringing that value into the scope using a `with` clause appended to the
`over` expression and yielding a
[record literal](../language/expressions.md#record-expressions) with the desired value:
```mdtest-command dir=docs/tutorials
super -z -c '
  over requested_reviewers with user=user.login => (
    reviewers:=union(login)
    |> {user,reviewers}
  )
  |> sort user,len(reviewers)
' prs.bsup
```
which gives us
```mdtest-output head
{user:"aswan",reviewers:|["mccanne"]|}
{user:"aswan",reviewers:|["nwt","mccanne"]|}
{user:"aswan",reviewers:|["nwt","henridf","mccanne"]|}
{user:"aswan",reviewers:|["henridf","mccanne","mattnibs"]|}
{user:"aswan",reviewers:|["henridf","mccanne","mattnibs"]|}
{user:"henridf",reviewers:|["nwt","aswan","mccanne"]|}
{user:"mattnibs",reviewers:|["aswan","mccanne"]|}
{user:"mattnibs",reviewers:|["aswan","henridf"]|}
...
```
The final step is to simply aggregate the "reviewer sets" with the `user` field
as the group-by key:
```mdtest-command dir=docs/tutorials
super -Z -c '
  over requested_reviewers with user=user.login => (
    reviewers:=union(login)
    |> {user,reviewers}
  )
  |> groups:=union(reviewers) by user
  |> sort user,len(groups)
' prs.bsup
```
and we get
```mdtest-output
{
    user: "aswan",
    groups: |[
        |[
            "mccanne"
        ]|,
        |[
            "nwt",
            "mccanne"
        ]|,
        |[
            "nwt",
            "henridf",
            "mccanne"
        ]|,
        |[
            "henridf",
            "mccanne",
            "mattnibs"
        ]|
    ]|
}
{
    user: "henridf",
    groups: |[
        |[
            "nwt",
            "aswan",
            "mccanne"
        ]|
    ]|
}
{
    user: "mattnibs",
    groups: |[
        |[
            "aswan",
            "henridf"
        ]|,
        |[
            "aswan",
            "mccanne"
        ]|,
        |[
            "aswan",
            "henridf",
            "mccanne"
        ]|,
        |[
            "nwt",
            "aswan",
            "henridf",
            "mccanne"
        ]|,
        |[
            "nwt",
            "aswan",
            "mccanne",
            "mikesbrown"
        ]|,
        |[
            "nwt",
            "aswan",
            "philrz",
            "henridf",
            "mccanne"
        ]|
    ]|
}
{
    user: "mccanne",
    groups: |[
        |[
            "nwt"
        ]|,
        |[
            "aswan"
        ]|,
        |[
            "mattnibs"
        ]|
    ]|
}
{
    user: "nwt",
    groups: |[
        |[
            "aswan"
        ]|,
        |[
            "aswan",
            "mattnibs"
        ]|,
        |[
            "henridf",
            "mattnibs"
        ]|,
        |[
            "mccanne",
            "mattnibs"
        ]|
    ]|
}
```
After a quick glance here, you can tell that `mccanne` looks for
very targeted reviews while `mattnibs` casts a wide net, at least
for the PRs from the beginning of the `zed` repo.

To quantify this concept, we can easily modify this query to compute
the average number of reviewers requested instead of the set of groups
of reviewers.  To do this, we just average the reviewer set size
with an aggregation:
```mdtest-command dir=docs/tutorials
super -z -c '
  over requested_reviewers with user=user.login => (
    reviewers:=union(login)
    |> {user,reviewers}
  )
  |> avg_reviewers:=avg(len(reviewers)) by user
  |> sort avg_reviewers
' prs.bsup
```
which produces
```mdtest-output
{user:"mccanne",avg_reviewers:1.}
{user:"nwt",avg_reviewers:1.75}
{user:"aswan",avg_reviewers:2.4}
{user:"mattnibs",avg_reviewers:2.9}
{user:"henridf",avg_reviewers:3.}
```

Of course, if you'd like the query output in JSON, you can just say `-j` and
`zq` will happily format the Zed sets as JSON arrays, e.g.,
```mdtest-command dir=docs/tutorials
super -j -c '
  over requested_reviewers with user=user.login => (
    reviewers:=union(login)
    |> {user,reviewers}
  )
  |> groups:=union(reviewers) by user
  |> sort user,len(groups)
' prs.bsup
```
produces
```mdtest-output
{"user":"aswan","groups":[["mccanne"],["nwt","mccanne"],["nwt","henridf","mccanne"],["henridf","mccanne","mattnibs"]]}
{"user":"henridf","groups":[["nwt","aswan","mccanne"]]}
{"user":"mattnibs","groups":[["aswan","henridf"],["aswan","mccanne"],["aswan","henridf","mccanne"],["nwt","aswan","henridf","mccanne"],["nwt","aswan","mccanne","mikesbrown"],["nwt","aswan","philrz","henridf","mccanne"]]}
{"user":"mccanne","groups":[["nwt"],["aswan"],["mattnibs"]]}
{"user":"nwt","groups":[["aswan"],["aswan","mattnibs"],["henridf","mattnibs"],["mccanne","mattnibs"]]}
```

## Key Takeaways

So to summarize, we gave you a tour here of `zq` and how the Zed data model
provide a powerful way do search, transformation, and analytics in a structured-like
way on data that begins its life as semi-structured JSON and is transformed
into the powerful super-structured format without having to create relational
tables and schemas.

As you can see, `zq` is a general-purpose tool that you can add to your bag
of tricks to:
* explore messy and confusing JSON data using shaping and sampling,
* transform JSON data in ad hoc ways, and
* develop transform logic for hitting APIs like the GitHub API to produce
clean data for analysis by `zq` or even export into other systems or for testing.

If you'd like to learn more, feel free to read through the
[language docs](../language/_index.md) in depth
or see how you can organize [data into a lake](../commands/super-db.md)
using a git-like commit model.
