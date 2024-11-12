### Function

&emsp; **isnull** &mdash; test if a value is null

### Synopsis

```
isnull(val: any) -> bool
```

### Description

The _isnull_ function returns true if the argument is a null value.

### Examples

A simple value is null:
```mdtest-command
echo 'null(int64) 1' | super -z -c 'yield isnull(this)' -
```
=>
```mdtest-output
true
false
```
