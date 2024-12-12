---
title: Functions
---

Functions appear in [expression](../expressions) context and
take Zed values as arguments and produce a value as a result. In addition to
the built-in functions listed below, Zed also allows for the creation of
[user-defined functions](../statements#func-statements).

A function-style syntax is also available for converting values to each of
Zed's [primitive types](../../formats/zed#1-primitive-types), e.g.,
`uint8()`, `time()`, etc. For details and examples, read about the
[`cast` function](cast) and how it is [used in expressions](../expressions#casts).

* [abs](abs) - absolute value of a number
* [base64](base64) - encode/decode base64 strings
* [bucket](bucket) - quantize a time or duration value into buckets of equal widths
* [cast](cast) - coerce a value to a different type
* [ceil](ceil) - ceiling of a number
* [cidr_match](cidr_match) - test if IP is in a network
* [compare](compare) - return an int comparing two values
* [coalesce](coalesce) - return first value that is not null, a "missing" error, or a "quiet" error
* [crop](crop) - remove fields from a value that are missing in a specified type
* [error](error) - wrap a value as an error
* [every](every) - bucket `ts` using a duration
* [fields](fields) - return the flattened path names of a record
* [fill](fill) - add null values for missing record fields
* [flatten](flatten) - transform a record into a flattened map
* [floor](floor) - floor of a number
* [grep](grep) - search strings inside of values
* [grok](grok) - parse a string into a structured record
* [has](has) - test existence of values
* [hex](hex) - encode/decode hexadecimal strings
* [has_error](has_error) - test if a value has an error
* [is](is) - test a value's type
* [is_error](is_error) - test if a value is an error
* [join](join) - concatenate array of strings with a separator
* [kind](kind) - return a value's type category
* [ksuid](ksuid) - encode/decode KSUID-style unique identifiers
* [len](len) - the type-dependent length of a value
* [levenshtein](levenshtein) Levenshtein distance
* [log](log) - natural logarithm
* [lower](lower) - convert a string to lower case
* [map](map) - apply a function to each element of an array or set
* [missing](missing) - test for the "missing" error
* [nameof](nameof) - the name of a named type
* [nest_dotted](nest_dotted) - transform fields in a record with dotted names to nested records
* [network_of](network_of) - the network of an IP
* [now](now) - the current time
* [order](order) - reorder record fields
* [parse_uri](parse_uri) - parse a string URI into a structured record
* [parse_zson](parse_zson) - parse ZSON text into a Zed value
* [pow](pow) - exponential function of any base
* [quiet](quiet) - quiet "missing" errors
* [regexp](regexp) - perform a regular expression search on a string
* [regexp_replace](regexp_replace) - replace regular expression matches in a string
* [replace](replace) - replace one string for another
* [round](round) - round a number
* [rune_len](rune_len) - length of a string in Unicode code points
* [shape](shape) - apply cast, fill, and order
* [split](split) - slice a string into an array of strings
* [sqrt](sqrt) - square root of a number
* [strftime](strftime) - format time values
* [trim](trim) - strip leading and trailing whitespace
* [typename](typename) - look up and return a named type
* [typeof](typeof) - the type of a value
* [typeunder](typeunder) - the underlying type of a value
* [under](under) - the underlying value
* [unflatten](unflatten) - transform a record with dotted names to a nested record
* [upper](upper) - convert a string to upper case
