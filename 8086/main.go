package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	inputFileFlag = flag.String("input", "", "8086 binary file to read")
	debugFlag     = flag.Bool("debug", false, "debug output")
	execFlag      = flag.Bool("exec", false, "execute instructions")
)

func main() {
	os.Exit(main1())
}

func main1() int {
	flag.Parse()

	/*
		fmt.Printf("0x%x\n", (0x7788&0xff00)>>8)
		fmt.Printf("0x%x\n", (0x6655&0xff00)|0x77)
		return 0
	*/

	data, err := os.ReadFile(*inputFileFlag)
	if err != nil {
		log.Fatal(err)
	}

	s := &simulator{}

	d := &disassembler{data: data}
	for d.di < len(data) {
		start := d.di
		in := d.nextInstruction()
		fmt.Print(in)

		// Simulate instructions
		if *execFlag {
			s.exec(d.di, in)
			fmt.Printf(" ; ip=%d, ", s.ip)
			for _, r := range s.regs {
				fmt.Printf("0x%x ", r)
			}
		}

		// Print debug info
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
