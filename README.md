# :children_crossing: crossing-go

### :children_crossing: Description
Crossing is a utility for storing objects in S3 while taking advantage of client side envelope encryption with KMS.  The native AWS CLI tool does not have an easy way to client-side-encrypted-upload's into S3.

This utility allows you to do client side encrypted uploads to S3 from the command line, allowing you to quickly upload files to S3 securely. It is a golang implementation of crossing, a Ruby utility.

### :children_crossing: Installation

### :children_crossing: Usage
Crossing is designed to be simple to use. To upload, you just need to provide a filepath, bucket location, region and which KMS key to use.

    crossing-go put \
      --kms-key-id abcde-12345-abcde-12345 \
      sourcefile destinationlocation

Downloading is basically the same:

    crossing-go get \
      sourcelocation destinationfile

Where destinationlocation and sourcelocation are of the form s3://bucketname/objectprefix/object . If destinationfile is ommitted, the last part of the sourcelocation key is used as the filename. That is, s3://foo/bar/baz.txt would be written to baz.txt

### :children_crossing: License

Refer to [LICENSE.md](LICENSE.md)
