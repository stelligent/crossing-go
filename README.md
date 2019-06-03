# :children_crossing: crossing-go

## :children_crossing: Description

Crossing is a utility for storing objects in S3 while taking advantage of client side envelope encryption with KMS.  The native AWS CLI tool does not have an easy way to client-side-encrypted-upload's into S3.

This utility allows you to do client side encrypted uploads to S3 from the command line, allowing you to quickly upload files to S3 securely. It is a golang implementation of crossing, a Ruby utility.

## :children_crossing: Installation

    git clone git@github.com:stelligent/crossing-go.git
    cd crossings-go
    go build

## :children_crossing: Verify Installation
Change into the directory where crossings-go was built

    ./crossing-go --version
    crossing-go version 0.0.5


## :children_crossing: Usage

Crossing is designed to be simple to use. To upload, you just need to provide a filepath, bucket location, region and which KMS key to use.

    crossing-go put \
      --kms-key-id abcde-12345-abcde-12345 \
      sourcefile destinationlocation

or with a KMS alias
    crossing-go put \
      --kms-key-id 'alias/foo' \
      sourcefile destinationlocation

Downloading is basically the same:

    crossing-go get \
      sourcelocation destinationfile

Where destinationlocation and sourcelocation are of the form s3://bucketname/objectprefix/object . If destinationfile is ommitted, the last part of the sourcelocation key is used as the filename. That is, s3://foo/bar/baz.txt would be written to baz.txt . If destinationlocation is a bare bucket or ends in "/", a destination object is created with the same name as the sourcefile. If destinationfile is a directory, a file is created with the object key as the filename.

A special feature is the ability to download a YAML-compatible file and print all of its values as environment variable exports, suitable for sourcing in a shell.

    s3://foo/test.yml
    ---
    a: foo
    b: bar
    c:
      d: cat
      e: dog
    f:
      - 1
      - 2

    crossing-go env s3://foo/test.yml

outputs

    export a="foo"
    export b="bar"
    export c__d="cat"
    export c__e="dog"
    export f__0="1"
    export f__1="2"

additionally a "-p" / "--prefix" option lets you specify a prefix

    crossing-go env -p testyml s3://foo/test.yml

outputs

    export testyml__a="foo"
    export testyml__b="bar"
    export testyml__c__d="cat"
    export testyml__c__e="dog"
    export testyml__f__0="1"
    export testyml__f__1="2"

## CAVEATS / KNOWN BUGS

The "env" subcommand does not correctly escape shell strings

## :children_crossing: License

Refer to [LICENSE.md](LICENSE.md)
