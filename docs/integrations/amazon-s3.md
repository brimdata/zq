---
weight: 1
title: Amazon S3
---

Zed tools can access [Amazon S3](https://aws.amazon.com/s3/) and
S3-compatible storage via `s3://` URIs. Details are described below.

## Region

You must specify an AWS region via one of the following:
* The `AWS_REGION` environment variable
* The `~/.aws/config` file
* The file specified by the `AWS_CONFIG_FILE` environment variable

You can create `~/.aws/config` by installing the
[AWS CLI](https://aws.amazon.com/cli/) and running `aws configure`.

{{% tip "Note" %}}

If using S3-compatible storage that does not recognize the concept of regions,
a region must still be specified, e.g., by providing a dummy value for
`AWS_REGION`.

{{% /tip %}}

## Credentials

You must specify AWS credentials via one of the following:
* The `AWS_ACCESS_KEY_ID` and`AWS_SECRET_ACCESS_KEY` environment variables
* The `~/.aws/credentials` file
* The file specified by the `AWS_SHARED_CREDENTIALS_FILE` environment variable

You can create `~/.aws/credentials` by installing the
[AWS CLI](https://aws.amazon.com/cli/) and running `aws configure`.

## Endpoint

To use S3-compatible storage not provided by AWS, set the `AWS_S3_ENDPOINT`
environment variable to the hostname or URI of the provider.

## Wildcard Support

[Like the AWS CLI tools themselves](https://repost.aws/knowledge-center/s3-event-notification-filter-wildcard),
Zed does not currently expand UNIX-style `*` wildcards in S3 URIs. If you
find this limitation is impacting your workflow, please add your use case
details as a comment in issue [super/1994](https://github.com/brimdata/super/issues/1994)
to help us track the priority of possible enhancements in this area.
