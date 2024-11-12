### Function

&emsp; **isnotnull** &mdash; test if a value is not null

### Synopsis

```
isnotnull(val: any) -> bool
```

### Description

The _isnotnull_ function returns true if the argument is not a null
value.

### Examples

A simple value is not null:
```mdtest-command
echo '"foo" null(string)' | super -z -c 'yield isnotnull(this)' -
```
=>
```mdtest-output
true
false
```
