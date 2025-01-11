### Aggregate Function

&emsp; **collect** &mdash; aggregate values into array

### Synopsis
```
collect(any) -> [any]
```

### Description

The _collect_ aggregate function organizes its input into an array.
If the input values vary in type, the return type will be an array
of union of the types encountered.

### Examples

Simple sequence collected into an array:
```mdtest-command
echo '1 2 3 4' | super -z -c 'collect(this)' -
```

```mdtest-output
[1,2,3,4]
```

Continuous collection over a simple sequence:
```mdtest-command
echo '1 2 3 4' | super -z -c 'yield collect(this)' -
```

```mdtest-output
[1]
[1,2]
[1,2,3]
[1,2,3,4]
```

Mixed types create a union type for the array elements:
```mdtest-command
echo '1 2 3 4 "foo"' | super -z -c 'collect(this)' -
```

```mdtest-output
[1,2,3,4,"foo"]
```

Create arrays of values bucketed by key:
```mdtest-command
echo '{a:1,k:1} {a:2,k:1} {a:3,k:2} {a:4,k:2}' |
  super -z -c 'collect(a) by k |> sort' -
```

```mdtest-output
{k:1,collect:[1,2]}
{k:2,collect:[3,4]}
```
