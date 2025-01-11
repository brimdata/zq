### Operator

&emsp; **load** &mdash; add and commit data to a pool

### Synopsis

```
load <pool>[@<branch>] [author <author>] [message <message>] [meta <meta>]
```

{{% tip "Note" %}}

The `load` operator is exclusively for working with pools in a
[SuperDB data lake](../../commands/super-db.md) and is not available for use in
[`super`](../../commands/super.md).

{{% /tip %}}

### Description

The `load` operator populates the specified `<pool>` with the values it
receives as input. Much like how [`super db load`](../../commands/super-db.md#load)
is used at the command line to populate a pool with data from files, streams,
and URIs, the `load` operator is used to save query results from your SuperPipe
query to a pool in the same SuperDB data lake. `<pool>` is a string indicating the
[name or ID](../../commands/super-db.md#data-pools) of the destination pool.
If the optional `@<branch>` string is included then the data will be committed
to an existing branch of that name, otherwise the `main` branch is assumed.
The `author`, `message`, and `meta` strings may also be provided to further
describe the committed data, similar to the same `super db load` options.

### Input Data

Examples below assume the existence of the SuperDB data lake created and populated
by the following commands:

```mdtest-command
export SUPER_DB_LAKE=example
super db -q init
super db -q create -orderby flip:asc coinflips
super db branch -q -use coinflips onlytails
echo '{flip:1,result:"heads"} {flip:2,result:"tails"}' |
  super db load -q -use coinflips -
super db -q create -orderby flip:asc bigflips
super db query -f text '
  from :branches
  |> yield pool.name + "@" + branch.name
  |> sort'
```

The lake then contains the two pools:

```mdtest-output
bigflips@main
coinflips@main
coinflips@onlytails
```

### Examples

_Modify some values, load them into the `main` branch of our empty `bigflips` pool, and see what was loaded_
```mdtest-command
super db -lake example query '
  from coinflips
  |> result:=upper(result)
  |> load bigflips
' > /dev/null

super db -lake example query -z 'from bigflips'
```

```mdtest-output
{flip:1,result:"HEADS"}
{flip:2,result:"TAILS"}
```

_Add a filtered subset of records to our `onlytails` branch, while also adding metadata_
```mdtest-command
super db -lake example query '
  from coinflips
  |> result=="tails"
  |> load coinflips@onlytails
      author "Steve"
      message "A subset"
      meta "\"Additional metadata\""
' > /dev/null

super db -lake example query -z 'from coinflips@onlytails'
```

```mdtest-output
{flip:2,result:"tails"}
```
