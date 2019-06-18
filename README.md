# :children_crossing: crossing-go

## :children_crossing: Description

Crossing is a utility for storing objects in S3 while taking advantage of client side envelope encryption with KMS.  The native AWS CLI tool does not have an easy way to client-side-encrypted-upload's into S3.

This utility allows you to do client side encrypted uploads to S3 from the command line, allowing you to quickly upload files to S3 securely. It is a golang implementation of crossing, a Ruby utility.

## :children_crossing: AWS Profile
Please note that crossing-go requires that your AWS credentials is properfly configured.
[Configuration and Credentials Files](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

## :children_crossing: Go Modules

Crossing-go has moved to go modules for dependency management
Prerequisites:
    *[Install latest Go 1.11 release](https://golang.org/dl/)
    
    ** Using go modules **
    When starting a new terminal session you can set an enverionment variable:
    export GO111MODULE=on

    ** or **
    GO111MODULE=on go [command]


## :children_crossing: Build

    git clone git@github.com:stelligent/crossing-go.git
    cd crossing-go
    go build

## :children_crossing: Verify Build
Change into the directory where crossing-go was built

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

## :children_crossing: Contributing
Contributing to crossing-go will require that unit tests pass. To run unit tests in a go module environment please follow 
the instructions under the go modules heading first.

Tests are located in each submodules directory:
├── cmd
│   ├── env.go
│   ├── get.go
│   ├── get_test.go
│   ├── put.go
│   ├── put_test.go
│   ├── root.go
│   ├── root_test.go
│   ├── util.go
│   └── util_test.go
├── crypto
│   ├── aes_cbc_content_cipher.go
│   ├── aes_cbc_content_cipher_test.go
│   ├── aes_cbc.go
│   ├── aes_cbc_test.go
│   ├── cipher_util.go
│   └── pkcs5_padder.go

Running tests for cmd module using the temp AWS profile:
    cd cmd
    AWS_PROFILE=temp go test ./...
    ok      github.com/stelligent/crossing-go/cmd
## CAVEATS / KNOWN BUGS

The "env" subcommand does not correctly escape shell strings

## :children_crossing: License

Refer to [LICENSE.md](LICENSE.md)
