package main

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"8086": main1,
	}))
}

func Test8086(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
