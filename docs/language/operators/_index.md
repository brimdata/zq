---
title: Operators
---

Operators process a sequence of input values to create an output sequence
and appear as the components of a [pipeline](../pipeline-model). In addition to the built-in
operators listed below, Zed also allows for the creation of
[user-defined operators](../statements.md#operator-statements).

* [assert](assert) - evaluate an assertion
* [combine](combine) - combine parallel pipeline branches into a single output
* [cut](cut) - extract subsets of record fields into new records
* [drop](drop) - drop fields from record values
* [file](from) - source data from a file
* [fork](fork) - copy values to parallel pipeline branches
* [from](from) - source data from pools, files, or URIs
* [fuse](fuse) - coerce all input values into a merged type
* [get](from) - source data from a URI
* [head](head) - copy leading values of input sequence
* [join](join) - combine data from two inputs using a join predicate
* [load](load) - add and commit data to a pool
* [merge](merge) - combine parallel pipeline branches into a single, ordered output
* [over](over) - traverse nested values as a lateral query
* [pass](pass) - copy input values to output
* [put](put) - add or modify fields of records
* [rename](rename) - change the name of record fields
* [sample](sample) - select one value of each shape
* [search](search) - select values based on a search expression
* [sort](sort) - sort values
* [summarize](summarize) -  perform aggregations
* [switch](switch) -  route values based on cases
* [tail](tail) - copy trailing values of input sequence
* [top](top) - get top N sorted values of input sequence
* [uniq](uniq) - deduplicate adjacent values
* [where](where) - select values based on a Boolean expression
* [yield](yield) - emit values from expressions
