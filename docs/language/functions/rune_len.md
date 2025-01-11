### Function

&emsp; **rune_len** &mdash; length of a string in Unicode code points

### Synopsis

```
rune_len(s: string) -> int64
```

### Description

The _rune_len_ function returns the number of Unicode code points in
the argument string `s`.  Since Zed strings are always encoded as UTF-8,
this length is the same as the number of UTF-8 characters.

### Examples

The length in UTF-8 characters of a smiley is 1:
```mdtest-command
echo '"hello" "😎"' | super -z -c 'yield rune_len(this)' -
```

```mdtest-output
5
1
```

The length in bytes of a smiley is 4:
```mdtest-command
echo '"hello" "😎"' | super -z -c 'yield len(bytes(this))' -
```

```mdtest-output
5
4
```
