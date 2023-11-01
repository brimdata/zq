### Function

&emsp; **map** &mdash; calls a function on each element of an array and returns the results

### Synopsis

```
map(v: array|set, f: function) -> array|set
```

### Description

The _map_ function applies function f to every element in array/set v and
returns the augmented array/set. Function f must be a function that takes
only one argument. A user defined function can be used for f.

### Examples

Upper case each element of an array:

```mdtest-command
echo '["foo","bar","baz"]' | zq -z 'yield map(this, upper)' -
```
=>
```mdtest-output
["FOO","BAR","BAZ"]
```

Using a user defined function to convert an epoch float to a time:

```mdtest-input udf.zed
func floatToTime(x): ( cast(x*1000000000, <time>) )
yield map(this, floatToTime)
```

```mdtest-command
echo '[1697151533.41415,1697151540.716529]' | zq -z -I udf.zed -
```
=>
```mdtest-output
[2023-10-12T22:58:53.414149888Z,2023-10-12T22:59:00.716528896Z]
```