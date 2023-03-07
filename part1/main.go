package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	read := func() byte {
		b := data[di]
		di++
		return b
	}

	rImm := func() uint16 {
		return uint16(read())
	}

	for {
		if di >= len(data) {
			break
		}

		var in inst

		b1 := read()
		switch {
		case b1>>1 == 0b1010000:
			_, w := dw(b1)
			in.name = "mov"
			in.field1.reg1 = Reg(w<<3 | byte(AL))
			in.field2.imm = rImm() | (rImm() << 8)
			in.field2.ptr = true
			in.field2.reg1, in.field2.reg2 = NoReg, NoReg
			fmt.Println(in.String())
		case b1>>1 == 0b1010001:
			_, w := dw(b1)
			in.name = "mov"
			in.field2.reg1 = Reg(w<<3 | byte(AL))
			in.field1.imm = rImm() | (rImm() << 8)
			in.field1.ptr = true
			in.field1.reg1, in.field1.reg2 = NoReg, NoReg
			fmt.Println(in.String())
		case b1>>1 == 0b1100011:
			_, w := dw(b1)
			mod, _, rm := modrm(read())

			in.name = "mov"
			switch mod {
			case 0b00: // Memory mode, no displacement *
				if rm == 0b110 {
					panic("* unhandled special case; direct address")
				}
				ea := effectiveAddr[rm]
				in.field1.reg1, in.field1.reg2 = ea[0], ea[1]
				in.field1.ptr = true
			case 0b01: // Memory mode, 8-bit displacement
				ea := effectiveAddr[rm]
				in.field1.reg1, in.field1.reg2 = ea[0], ea[1]
				in.field1.imm = rImm()
				in.field1.ptr = true
			case 0b10: // Memory mode, 16-bit displacement
				// TODO
				ea := effectiveAddr[rm]
				in.field1.reg1, in.field1.reg2 = ea[0], ea[1]
				in.field1.imm = rImm() | (rImm() << 8)
				in.field1.ptr = true
			case 0b11: // Register mode (no displacement)
				in.field1.reg1 = register(rm, w)
			}
			in.field2.imm = rImm()
			in.field2.size = sizeByte
			if w > 0 {
				in.field2.imm |= (rImm() << 8)
				in.field2.size = sizeWord
			}
			fmt.Println(in.String())
		case b1>>2 == 0b100010:
			d, w := dw(b1)
			mod, reg, rm := modrm(read())
			dst := register(reg, w)

			in.name = "mov"
			in.field2.reg1 = dst

			switch mod {
			case 0b00: // Memory mode, no displacement *
				in.field1.ptr = true
				if rm == 0b110 {
					in.field1.reg1, in.field1.reg2 = NoReg, NoReg
					in.field1.imm = rImm() | (rImm() << 8)
				} else {
					ea := effectiveAddr[rm]
					in.field1.reg1, in.field1.reg2 = ea[0], ea[1]
					in.field1.ptr = true
				}
			case 0b01: // Memory mode, 8-bit displacement
				ea := effectiveAddr[rm]
				in.field1.reg1, in.field1.reg2 = ea[0], ea[1]
				in.field1.imm = rImm()
				in.field1.ptr = true
			case 0b10: // Memory mode, 16-bit displacement
				// TODO
				ea := effectiveAddr[rm]
				in.field1.reg1, in.field1.reg2 = ea[0], ea[1]
				in.field1.imm = rImm() | (rImm() << 8)
				in.field1.ptr = true
			case 0b11: // Register mode (no displacement)
				in.field1.reg1 = register(rm, w)
			}
			if d > 0 {
				in.field1, in.field2 = in.field2, in.field1
			}
			fmt.Println(in.String())
		case b1>>4 == 0b1011:
			w := (b1 >> 3) & 0b1
			reg := b1 & 0b111
			src := register(reg, w)
			imm := rImm()
			if w == 1 {
				imm2 := rImm()
				imm |= (imm2 << 8)
			}
			fmt.Printf("mov %s, %d\n", formatReg(src), imm)
		default:
			panic("unsupported opcode")
		}
	}
}

type inst struct {
	name           string
	field1, field2 field
}

func (i inst) String() string {
	return fmt.Sprintf("%s %s, %s", i.name, i.field1, i.field2)
}

type field struct {
	reg1 Reg
	reg2 Reg
	imm  uint16
	size size
	ptr  bool
}

func (f field) String() string {
	var sb strings.Builder
	if f.ptr {
		sb.WriteString("[")
		written := false
		if f.reg1 != NoReg {
			written = true
			sb.WriteString(formatReg(f.reg1))
		}
		if f.reg2 != NoReg {
			written = true
			sb.WriteString(fmt.Sprintf(" + %s", formatReg(f.reg2)))
		}
		if f.imm > 0 {
			if written {
				sb.WriteString(" + ")
			}
			sb.WriteString(fmt.Sprintf("%d", f.imm))
		}
		sb.WriteString(fmt.Sprintf("]"))
	} else if f.imm > 0 {
		if f.size != "" {
			sb.WriteString(fmt.Sprintf("%s ", f.size))
		}
		sb.WriteString(fmt.Sprintf("%d", f.imm))
	} else {
		sb.WriteString(formatReg(f.reg1))
	}
	return sb.String()
}

type size string

const (
	sizeByte = "byte"
	sizeWord = "word"
)

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

func modrm(b byte) (byte, byte, byte) {
	mod := (b >> 6) & 0b11
	reg := (b >> 3) & 0b111
	rm := b & 0b111
	return mod, reg, rm
}

func dw(b byte) (byte, byte) {
	d := b & 0b00000010 // 0: inst src specified in REG, 1: inst dst specified in REG
	w := b & 0b00000001 // 0: inst operates on byte, 1: inst operates on word
	return d, w
}

func register(r byte, w byte) Reg {
	return Reg((w << 3) | r)
}

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

	NoReg = 0b11111
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
