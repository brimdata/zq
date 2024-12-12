---
title: Commands
weight: 2
---

The [`super` command](super) is used to execute command-line queries on
inputs from files, HTTP URLs, or [S3](../integrations/amazon-s3).

The [`super db` sub-commands](super-db) are for creating, configuring, ingesting
into, querying, and orchestrating SuperDB data lakes. These sub-commands are
organized into further subcommands like the familiar command patterns
of `docker` or `kubectl`.

All operations with these commands utilize the [super data model](../formats)
and provide querying via [SuperSQL](../language).

Built-in help for `super` and all sub-commands is always accessible with the `-h` flag.
