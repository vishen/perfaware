package main

import (
	"flag"
	"log"
	"os"
)

var (
	inputFileFlag = flag.String("input", "", "8086 binary file to read")
	debugFlag     = flag.Bool("d", false, "debug output")
)

func main() {
	os.Exit(main1())
}

func main1() int {
	flag.Parse()

	data, err := os.ReadFile(*inputFileFlag)
	if err != nil {
		log.Fatal(err)
	}
	disassemble(data)

	return 0
}
