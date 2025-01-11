### Function

&emsp; **join** &mdash; concatenate array of strings with a separator

### Synopsis

```
join(val: [string], sep: string) -> string
```

### Description

The _join_ function concatenates the elements of string array `val` to create a single
string. The string `sep` is placed between each value in the resulting string.

#### Example:

Join a symbol array of strings:
```mdtest-command
echo '["a","b","c"]' | super -z -c 'yield join(this, ",")' -
```

```mdtest-output
"a,b,c"
```

Join non-string arrays by first casting:
```mdtest-command
echo '[1,2,3] [10.0.0.1,10.0.0.2]' |
  super -z -c 'yield join(cast(this, <[string]>), "...")' -
```

```mdtest-output
"1...2...3"
"10.0.0.1...10.0.0.2"
```
