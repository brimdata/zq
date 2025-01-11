### Operator

&emsp; **where** &mdash; select values based on a Boolean expression

### Synopsis
```
[where] <expr>
```
### Description

The `where` operator filters its input by applying a Boolean expression `<expr>`
to each input value and dropping each value for which the expression evaluates
to `false` or to an error.

The `where` keyword is optional since it is an
[implied operator](../pipeline-model.md#implied-operators).

The "where" keyword requires a boolean-valued expression and does not support
[search expressions](../search-expressions.md).  Use the
[search operator](search.md) if you want search syntax.

When SuperPipe queries are run interactively, it is highly convenient to be able to omit
the "where" keyword, but when where filters appear in query source files,
it is good practice to include the optional keyword.

### Examples

_An arithmetic comparison_
```mdtest-command
echo '1 2 3' | super -z -c 'where this >= 2' -
```

```mdtest-output
2
3
```
_The "where" keyword may be dropped_
```mdtest-command
echo '1 2 3' | super -z -c 'this >= 2' -
```

```mdtest-output
2
3
```
_A filter with Boolean logic_
```mdtest-command
echo '1 2 3' | super -z -c 'where this >= 2 AND this <= 2' -
```

```mdtest-output
2
```
_A filter with array containment logic_
```mdtest-command
echo '1 2 3 4' | super -z -c 'where this in [1,4]' -
```

```mdtest-output
1
4
```
_A filter with inverse containment logic_
```mdtest-command
echo '1 2 3 4' | super -z -c 'where ! (this in [1,4])' -
```

```mdtest-output
2
3
```
_Boolean functions may be called_
```mdtest-command
echo '1 "foo" 10.0.0.1' | super -z -c 'where is(<int64>)' -
```

```mdtest-output
1
```
_Boolean functions with Boolean logic_
```mdtest-command
echo '1 "foo" 10.0.0.1' | super -z -c 'where is(<int64>) or is(<ip>)' -
```

```mdtest-output
1
10.0.0.1
```
