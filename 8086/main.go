package main

import (
	"flag"
	"fmt"
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
	d := &disassembler{data: data}
	for d.di < len(data) {
		start := d.di
		in := d.nextInstruction()
		fmt.Print(in)
		if *debugFlag {
			fmt.Printf(" (")
			for i := start; i < d.di; i++ {
				fmt.Printf(" %08b", data[i])
			}
			fmt.Printf(" )")
		}
		fmt.Println()
	}

	return 0
}
