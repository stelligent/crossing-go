package integration

import (
	"flag"
	"os"
	"testing"
)

var (
	all = flag.Bool("all", false, "run all integration tests")
	put = flag.Bool("put", false, "run put integration tests")
)

func TestMain(t *testing.M) {
	flag.Parse()

	if *all {
		*put = true
	}

	if *put {
		SetupPut()
	}

	result := t.Run()

	if *put {
		PutCleanUp()
	}
	os.Exit(result)
}
