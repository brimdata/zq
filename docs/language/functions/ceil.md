### Function

&emsp; **ceil** &mdash; ceiling of a number

### Synopsis

```
ceil(n: number) -> number
```

### Description

The _ceil_ function returns the smallest integer greater than or equal to its argument `n`,
which must be a numeric type.  The return type retains the type of the argument.

### Examples

The ceiling of a various numbers:
```mdtest-command
echo '1.5 -1.5 1(uint8) 1.5(float32)' | super -z -c 'yield ceil(this)' -
```

```mdtest-output
2.
-1.
1(uint8)
2.(float32)
```
