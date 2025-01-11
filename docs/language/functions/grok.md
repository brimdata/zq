### Function

&emsp; **grok** &mdash; parse a string using a Grok pattern

### Synopsis

```
grok(p: string, s: string) -> record
grok(p: string, s: string, definitions: string) -> record
```

### Description

The _grok_ function parses a string `s` using Grok pattern `p` and returns
a record containing the parsed fields. The syntax for pattern `p`
is `%{pattern:field_name}` where _pattern_ is the name of the pattern
to match in `s` and _field_name_ is the resultant field name of the capture
value.

When provided with three arguments, `definitions` is a string
of named patterns in the format `PATTERN_NAME PATTERN` each separated by
newlines (`\n`). The named patterns can then be referenced in argument `p`.

### Included Patterns

The `grok` function by default includes a set of built-in named patterns
that can be referenced in any pattern. The included named patterns can be seen
[here](https://raw.githubusercontent.com/brimdata/super/main/pkg/grok/base.go).

### Comparison to Other Implementations

Although Grok functionality appears in many open source tools, it lacks a
formal specification. As a result, example parsing configurations found via
web searches may not all plug seamlessly into SuperPipe's `grok` function without
modification.

[Logstash](https://www.elastic.co/logstash) was the first tool to widely
promote the approach via its
[Grok filter plugin](https://www.elastic.co/guide/en/logstash/current/plugins-filters-grok.html),
so it serves as the de facto reference implementation. Many articles have
been published by Elastic and others that provide helpful guidance on becoming
proficient in Grok. To help you adapt what you learn from these resources to
the use of the `grok` function, review the tips below.

{{% tip "Note" %}}

As these represent areas of possible future SuperPipe enhancement, links to open
issues are provided. If you find a functional gap significantly impacts your
ability to use the `grok` function, please add a comment to the relevant
issue describing your use case.

{{% /tip %}}

1. Logstash's Grok offers an optional data type conversion syntax,
   e.g.,
   ```
   %{NUMBER:num:int}
   ```
  to store `num` as an integer type instead of as a
  string. SuperPipe currently accepts this trailing `:type` syntax but effectively
  ignores it and stores all parsed values as strings. Downstream use of the
  [`cast` function](cast.md) can be used instead for data type conversion.
  ([super/4928](https://github.com/brimdata/super/issues/4928))

2. Some Logstash Grok examples use an optional square bracket syntax for
   storing a parsed value in a nested field, e.g.,
   ```
   %{GREEDYDATA:[nested][field]}
   ```
   to store a value into `{"nested": {"field": ... }}`. In SuperPipe the more common
   dot-separated field naming convention `nested.field` can be combined
   with the downstream use of the [`nest_dotted` function](nest_dotted.md) to
   store values in nested fields.
   ([super/4929](https://github.com/brimdata/super/issues/4929))

3. SuperPipe's regular expressions syntax does not currently support the
   "named capture" syntax shown in the
   [Logstash docs](https://www.elastic.co/guide/en/logstash/current/plugins-filters-grok.html#_custom_patterns).
   ([super/4899](https://github.com/brimdata/super/issues/4899))

   Instead use the the approach shown later in that section of the Logstash
   docs by including a custom pattern in the `definitions` argument, e.g.,

   ```mdtest-command
   echo '"Jan  1 06:25:43 mailserver14 postfix/cleanup[21403]: BEF25A72965: message-id=<20130101142543.5828399CCAF@mailserver14.example.com>"' |
     super -Z -c 'yield grok("%{SYSLOGBASE} %{POSTFIX_QUEUEID:queue_id}: %{GREEDYDATA:syslog_message}",
                       this,
                       "POSTFIX_QUEUEID [0-9A-F]{10,11}")' -
   ```

   produces

   ```mdtest-output
   {
       timestamp: "Jan  1 06:25:43",
       logsource: "mailserver14",
       program: "postfix/cleanup",
       pid: "21403",
       queue_id: "BEF25A72965",
       syslog_message: "message-id=<20130101142543.5828399CCAF@mailserver14.example.com>"
   }
   ```

4. The Grok implementation for Logstash uses the
   [Oniguruma](https://github.com/kkos/oniguruma) regular expressions library
   while SuperPipe's `grok` uses Go's [regexp](https://pkg.go.dev/regexp) and
   [RE2 syntax](https://github.com/google/re2/wiki/Syntax). These
   implementations share the same basic syntax which should suffice for most
   parsing needs. But per a detailed
   [comparison](https://en.wikipedia.org/wiki/Comparison_of_regular_expression_engines),
   Oniguruma does provide some advanced syntax not available in RE2,
   such as recursion, look-ahead, look-behind, and backreferences. To
   avoid compatibility issues, we recommend building configurations starting
   from the RE2-based [included patterns](#included-patterns).

{{% tip "Note" %}}

If you absolutely require features of Logstash's Grok that are not currently
present in SuperPipe, you can create a Logstash-based preprocessing
pipeline that uses its
[Grok filter plugin](https://www.elastic.co/guide/en/logstash/current/plugins-filters-grok.html)
and send its output as JSON to SuperPipe. Issue
[super/3151](https://github.com/brimdata/super/issues/3151) provides some tips for
getting started. If you pursue this approach, please add a comment to the
issue describing your use case or come talk to us on
[community Slack](https://www.brimdata.io/join-slack/).

{{% /tip %}}

### Debugging

Much like creating complex regular expressions, building sophisticated Grok
configurations can be frustrating because single-character mistakes can make
the difference between perfect parsing and total failure.

A recommended workflow is to start by successfully parsing a small/simple
portion of your target data and
[incrementally](https://www.elastic.co/blog/slow-and-steady-how-to-build-custom-grok-patterns-incrementally)
adding more parsing logic and re-testing at each step.

To aid in this workflow, you may find an
[interactive Grok debugger](https://grokdebugger.com/) helpful. However, note
that these have their own
[differences and limitations](https://github.com/cjslack/grok-debugger).
If you devise a working Grok config in such a tool be sure to incrementally
test it with SuperPipe's `grok`. Be mindful of necessary adjustments such as those
described [above](#comparison-to-other-implementations) and in the [examples](#examples).

### Need Help?

If you have difficulty with your Grok configurations, please come talk to us
on the [community Slack](https://www.brimdata.io/join-slack/).

### Examples

Parsing a simple log line using the built-in named patterns:
```mdtest-command
echo '"2020-09-16T04:20:42.45+01:00 DEBUG This is a sample debug log message"' |
  super -Z -c 'yield grok("%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{GREEDYDATA:message}",
                    this)' -
```

```mdtest-output
{
    timestamp: "2020-09-16T04:20:42.45+01:00",
    level: "DEBUG",
    message: "This is a sample debug log message"
}
```

As with any [string literal](../expressions.md#literals), the
leading backslash in escape sequences in string arguments must be doubled,
such as changing the `\d` to `\\d` if we repurpose the
[included pattern](#included-patterns) for `NUMTZ` as a `definitions` argument:

```mdtest-command
echo '"+7000"' |
  super -z -c 'yield grok("%{MY_NUMTZ:tz}",
                    this,
                    "MY_NUMTZ [+-]\\d{4}")' -
```

```mdtest-output
{tz:"+7000"}
```

In addition to using `\n` newline escapes to separate multiple named patterns
in the `definitions` argument, string concatenation via `+` may further enhance
readability.

```mdtest-command
echo '"(555)-1212"' |
  super -z -c 'yield grok("\\(%{PH_PREFIX:prefix}\\)-%{PH_LINE_NUM:line_number}",
                    this,
                    "PH_PREFIX \\d{3}\n" +
                    "PH_LINE_NUM \\d{4}")' -
```

```mdtest-output
{prefix:"555",line_number:"1212"}
```
