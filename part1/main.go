package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var debugFlag = flag.Bool("debug", false, "debug output")

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	disassemble(data)
}

func disassemble(data []byte) {
	di := 0
	for {
		if di >= len(data) {
			break
		}

		b1 := data[di]
		b2 := data[di+1]
		di += 2
		debug("b1=%b b2=%b\n", b1, b2)

		opcode := (b1 >> 2) & 0b111111
		d := b1 & 0b00000010 // 0: inst src specified in REG, 1: inst dst specified in REG
		w := b1 & 0b00000001 // 0: inst operates on byte, 1: inst operates on word
		mod := (b2 >> 6) & 0b11
		reg := (b2 >> 3) & 0b111
		rm := b2 & 0b111
		debug("opcode=%b d=%b w=%b mod=%b reg=%b rm=%b\n", opcode, d, w, mod, reg, rm)

		assert(mod == 0b11, "mod 0b11 is currently only recognised")

		dst := Reg((w << 3) | reg)
		src := Reg((w << 3) | rm)
		if d == 1 {
			src, dst = dst, src
		}

		switch Opcode(opcode) {
		case mov:
			fmt.Printf("mov %s, %s\n", formatReg(src), formatReg(dst))
		default:
			panic("unsupported opcode")
		}
	}
}

type Opcode int

const (
	mov Opcode = 0b100010
)

type Reg int

const (
	// W = 0
	AL Reg = 0b000
	CL Reg = 0b001
	DL Reg = 0b010
	BL Reg = 0b011
	AH Reg = 0b100
	CH Reg = 0b101
	DH Reg = 0b110
	BH Reg = 0b111

	// W = 1
	AX Reg = 0b1000
	CX Reg = 0b1001
	DX Reg = 0b1010
	BX Reg = 0b1011
	SP Reg = 0b1100
	BP Reg = 0b1101
	SI Reg = 0b1110
	DI Reg = 0b1111
)

func formatReg(r Reg) string {
	switch r {
	case AL:
		return "al"
	case CL:
		return "cl"
	case DL:
		return "dl"
	case BL:
		return "bl"
	case AH:
		return "ah"
	case CH:
		return "ch"
	case DH:
		return "dh"
	case BH:
		return "bh"
	case AX:
		return "ax"
	case CX:
		return "cx"
	case DX:
		return "dx"
	case BX:
		return "bx"
	case SP:
		return "sp"
	case BP:
		return "bp"
	case SI:
		return "si"
	case DI:
		return "di"
	}
	return "INVALID REG"
}

func assert(cond bool, msg string, args ...any) {
	if !cond {
		panic(fmt.Sprintf(msg, args...))
	}
}

func debug(msg string, args ...any) {
	if *debugFlag {
		fmt.Printf(msg, args...)
	}
}
