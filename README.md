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
## :children_crossing: Contributing
Crossing-go has move to go modules for dependency management. Unit tests can be ran locally via the go test command.
    ~/crossing-go/cmd$ go test
    crossing-go implements get/put to S3 using KMS envelope
    client-side encryption with the AWS SDK. It is intended to be object
    compatible with the Ruby crossing utility.

    Usage:
    crossing-go [command]

    Available Commands:
    env         Retrieve an object from S3 and export as environment variables
    get         Retrieve an object from S3
    help        Help about any command
    put         Upload a file to S3

    Flags:
        --config string   config file (default is $HOME/.crossing-go.yaml)
    -h, --help            help for crossing-go
        --version         version for crossing-go

    Use "crossing-go [command] --help" for more information about a command.
    PASS
    ok      github.com/stelligent/crossing-go/cmd   0.006s

Integration tests are ran to ensure the application can make proper API calls to AWS. This means that an authentication token is required or that AWS CLI must be configured with access key id and access key. The below example uses an aws-vault setup with mfa.
    ~/crossing-go/cmd$ aws-vault exec home -- go test -all
    wrote 5 bytes
    Waiting for bucket "crossinggojqvzsgmyj" to be created...
    Bucket "crossinggojqvzsgmyj" successfully created
    Successfully configured versioning "{\n\n}"{
    KeyMetadata: {
        AWSAccountId: "324320755747",
        Arn: "arn:aws:kms:us-east-2:324320755747:key/54348bc1-6e3b-4cda-8b18-c6033ca7d328",
        CreationDate: 2019-07-12 18:23:13 +0000 UTC,
        Description: "",
        Enabled: true,
        KeyId: "54348bc1-6e3b-4cda-8b18-c6033ca7d328",
        KeyManager: "CUSTOMER",
        KeyState: "Enabled",
        KeyUsage: "ENCRYPT_DECRYPT",
        Origin: "AWS_KMS"
    }
    }
    Returning key: "54348bc1-6e3b-4cda-8b18-c6033ca7d328"{

    }
    { "VersionId": "IqAC410t2VPQON6xubtS0BO_JPMeESP8" }
    PASS
    Deleted:  {
    VersionId: "IqAC410t2VPQON6xubtS0BO_JPMeESP8"
    }
    Waiting for object to be deleted: "dat1", Id: "IqAC410t2VPQON6xubtS0BO_JPMeESP8"Delete was successful {

    }
    Key deletion scheduled:  {
    DeletionDate: 2019-07-20 00:00:00 +0000 UTC,
    KeyId: "arn:aws:kms:us-east-2:324320755747:key/54348bc1-6e3b-4cda-8b18-c6033ca7d328"
    }
    ok      github.com/stelligent/crossing-go/integration   2.963s
## CAVEATS / KNOWN BUGS

The "env" subcommand does not correctly escape shell strings

## :children_crossing: License

Refer to [LICENSE.md](LICENSE.md)
