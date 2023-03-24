package main

import (
	"log"
	"os"
)

func main() {
	data, err := os.ReadFile("./test")
	if err != nil {
		log.Fatal(err)
	}
	disassemble(data)
}
