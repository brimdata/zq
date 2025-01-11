### Function

&emsp; **len** &mdash; the type-dependent length of a value

### Synopsis

```
len(v: record|array|set|map|type|bytes|string|ip|net|error) -> int64
```

### Description

The _len_ function returns the length of its argument `val`.
The semantics of this length depend on the value's type.

Supported types include:
- record
- array
- set
- map
- error
- bytes
- string
- ip
- net
- type

#### Example:

Take the length of various types:

```mdtest-command
echo '[1,2,3] |["hello"]| {a:1,b:2} "hello" 10.0.0.1 1' |
  super -z -c 'yield {this,len:len(this)}' -
```

```mdtest-output
{this:[1,2,3],len:3}
{this:|["hello"]|,len:1}
{this:{a:1,b:2},len:2}
{this:"hello",len:5}
{this:10.0.0.1,len:4}
{this:1,len:error({message:"len: bad type",on:1})}
```
