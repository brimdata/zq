### Aggregate Function

&emsp; **or** &mdash; logical OR of input values

### Synopsis
```
or(bool) -> bool
```

### Description

The _or_ aggregate function computes the logical OR over all of its input.

### Examples

Ored value of simple sequence:
```mdtest-command
echo 'false true false' | super -z -c 'or(this)' -
```

```mdtest-output
true
```

Continuous OR of simple sequence:
```mdtest-command
echo 'false true false' | super -z -c 'yield or(this)' -
```

```mdtest-output
false
true
true
```

Unrecognized types are ignored and not coerced for truthiness:
```mdtest-command
echo 'false "foo" 1 true false' | super -z -c 'yield or(this)' -
```

```mdtest-output
false
false
false
true
true
```

OR of values grouped by key:
```mdtest-command
echo '{a:true,k:1} {a:false,k:1} {a:false,k:2} {a:false,k:2}' |
  super -z -c 'or(a) by k |> sort' -
```

```mdtest-output
{k:1,or:true}
{k:2,or:false}
```
