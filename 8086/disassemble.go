package main

import (
	"fmt"
	"log"
	"strings"
)

func disassemble(data []byte) []Instruction {
	for _, d := range data {
		fmt.Printf("%08b\n", d)
	}

	di := 0

	// read next byte
	next := func() byte {
		b := data[di]
		di++
		return b
	}

	// read an 8-bit immediate
	imm8 := func() uint16 {
		return uint16(next())
	}

	// read full 16-bit immediate
	imm16 := func() uint16 {
		return imm8() | (imm8() << 8)
	}

	handleModRM := func(mod, rm, w byte) Operand {
		var o Operand
		switch mod {
		case 0b00: // Memory mode, no displacement *
			o.Ptr = true
			if rm == 0b110 { // * special case
				o.Imm = imm16()
			} else {
				ea := effectiveAddr[rm]
				o.Reg1, o.Reg2 = ea[0], ea[1]
				o.Ptr = true
			}
		case 0b01: // Memory mode, 8-bit displacement
			ea := effectiveAddr[rm]
			o.Reg1, o.Reg2 = ea[0], ea[1]
			o.Imm = imm8()
			o.Ptr = true
		case 0b10: // Memory mode, 16-bit displacement
			ea := effectiveAddr[rm]
			o.Reg1, o.Reg2 = ea[0], ea[1]
			o.Imm = imm16()
			o.Ptr = true
		case 0b11: // Register mode (no displacement)
			o.Reg1 = register(rm, w)
		}
		return o
	}

	for di < len(data) {
		b := next()
		bi := 0

		read := func(length int) byte {
			if bi == 8 {
				b = next()
				bi = 0
			}
			d := (b >> (8 - (bi + length))) & mask(length)
			bi += length
			return d
		}

		encs := encoder.Decode(b)
		if len(encs) == 0 {
			log.Fatalf("unable to decode %08b at pos %d", b, di)
		}

		enc := encs[0]

		in := Instruction{
			Name:   enc.Name,
			Opcode: enc.Opcode.Opcode,
		}

		hasMod := false
		hasData := false
		hasReg := false
		hasW := false
		for _, b := range enc.Bytes {
			for _, p := range b {
				switch p.Name {
				case "S":
					in.S = read(p.Len)
				case "D":
					in.D = read(p.Len)
				case "W":
					in.W = read(p.Len)
					hasW = true
				case "MOD":
					in.Mod = read(p.Len)
					hasMod = true
				case "REG":
					hasReg = true
					in.Reg = read(p.Len)
					if hasW {
						in.Operand1.Reg1 = register(in.Reg, in.W)
					} else {
						in.Operand1.Reg1 = register(in.Reg, 1)
					}
				case "RM":
					in.RM = read(p.Len)
					if hasReg {
						in.Operand1, in.Operand2 = in.Operand2, in.Operand1
					}
				case "SR":
					switch read(p.Len) {
					case 0b00:
						in.Operand1.Reg1 = ES
					case 0b01:
						in.Operand1.Reg1 = CS
					case 0b10:
						in.Operand1.Reg1 = SS
					case 0b011:
						in.Operand1.Reg1 = DS
					}
				case "DATA":
					hasData = true
				case "DISPLO", "DISPHI", "DATAW":
					// IGNORE
				default:
					// Constant
					c := read(p.Len)
					if c != p.Const {
						fmt.Println(encs)
						fmt.Println(b)
						fmt.Println(p, c, p.Const)
						log.Fatalf("CONSTANT NOT CORRECT, FIX ME")
					}

				}
			}
		}
		if hasMod {
			in.Operand1 = handleModRM(in.Mod, in.RM, in.W)
		}
		if hasData {
			if in.W > 0 && in.S == 0 {
				in.Operand2.Imm = imm16()
			} else {
				in.Operand2.Imm = imm8()
			}
		}
		if in.D > 0 {
			in.Operand1, in.Operand2 = in.Operand2, in.Operand1
		}
		fmt.Println(in)
	}

	return nil
}

var effectiveAddr = map[byte][2]Reg{
	0b000: [2]Reg{BX, SI},
	0b001: [2]Reg{BX, DI},
	0b010: [2]Reg{BP, SI},
	0b011: [2]Reg{BP, DI},
	0b100: [2]Reg{SI, NoReg},
	0b101: [2]Reg{DI, NoReg},
	0b110: [2]Reg{BP, NoReg},
	0b111: [2]Reg{BX, NoReg},
}

func register(r byte, w byte) Reg {
	return Reg((w << 4) | (1 << 3) | r)
}

type Reg byte

const (
	NoReg Reg = 0b000

	ES Reg = 0b1
	CS Reg = 0b10
	SS Reg = 0b11
	DS Reg = 0b100

	// W = 0
	AL Reg = 0b1000
	CL Reg = 0b1001
	DL Reg = 0b1010
	BL Reg = 0b1011
	AH Reg = 0b1100
	CH Reg = 0b1101
	DH Reg = 0b1110
	BH Reg = 0b1111

	// W = 1
	AX Reg = 0b11000
	CX Reg = 0b11001
	DX Reg = 0b11010
	BX Reg = 0b11011
	SP Reg = 0b11100
	BP Reg = 0b11101
	SI Reg = 0b11110
	DI Reg = 0b11111
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
	case ES:
		return "es"
	case CS:
		return "cs"
	case SS:
		return "ss"
	case DS:
		return "ds"
	}
	return "INVALID REG"
}

type Instruction struct {
	Name   string
	Opcode byte
	D      byte
	W      byte
	S      byte
	Mod    byte
	Reg    byte
	RM     byte

	Operand1 Operand
	Operand2 Operand
}

func (i Instruction) String() string {
	// TODO: do this better
	switch i.Name {
	case "PUSH", "POP", "INC", "DEC":
		return fmt.Sprintf("%s %s", i.Name, i.Operand1)
	}
	return fmt.Sprintf("%s %s, %s", i.Name, i.Operand1, i.Operand2)
}

type Operand struct {
	Reg1 Reg
	Reg2 Reg
	Imm  uint16
	Size size
	Ptr  bool
}

func (o Operand) String() string {
	var sb strings.Builder
	if o.Ptr {
		sb.WriteString("[")
		written := false
		if o.Reg1 != NoReg {
			written = true
			sb.WriteString(formatReg(o.Reg1))
		}
		if o.Reg2 != NoReg {
			written = true
			sb.WriteString(fmt.Sprintf(" + %s", formatReg(o.Reg2)))
		}
		if o.Imm > 0 {
			if written {
				sb.WriteString(" + ")
			}
			sb.WriteString(fmt.Sprintf("0x%x", o.Imm))
		}
		sb.WriteString(fmt.Sprintf("]"))
	} else if o.Imm > 0 {
		if o.Size != "" {
			sb.WriteString(fmt.Sprintf("%s ", o.Size))
		}
		sb.WriteString(fmt.Sprintf("0x%x", o.Imm))
	} else {
		sb.WriteString(formatReg(o.Reg1))
	}
	return sb.String()
}

type size string

const (
	sizeByte = "byte"
	sizeWord = "word"
)
