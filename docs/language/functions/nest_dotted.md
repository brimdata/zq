### Function

&emsp; **nest_dotted** &mdash; transform fields in a record with dotted names
to nested records.

### Synopsis

```
nest_dotted(val: record) -> record
```

### Description
The _nest_dotted_ function returns a copy of `val` with all dotted field names
converted into nested records. If no argument is supplied to `nest_dotted`,
`nest_dotted` operates on `this`.

### Examples

```mdtest-command
echo '{"a.b.c":"foo"}' | super -z -c 'yield nest_dotted()' -
```

```mdtest-output
{a:{b:{c:"foo"}}}
```
