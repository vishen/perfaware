package main

import (
	"fmt"
	"strings"
)

/*
func disassemble(data []byte) []Instruction {
	d := &disassembler{data: data}
	for d.di < len(data) {
		start := d.di
		b := d.next()
		encs := encoder.Decode(b)
		if len(encs) == 0 {
			log.Fatalf("unable to decode %08b at pos %d", b, d.di)
		}

		found := false
		for _, enc := range encs {
			in, ok := d.parse(enc)
			if !ok {
				continue
			}
			found = true
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
		if !found {
			log.Fatalf("unable to find instruction encoding for %08b at index %d", b, d.di-1)
		}
	}

	return nil
}
*/

type disassembler struct {
	data []byte
	di   int

	curByte byte
	cbi     int
}

func (d *disassembler) nextInstruction() Instruction {
	b := d.next()
	encs := encoder.Decode(b)
	if len(encs) == 0 {
		panic(fmt.Sprintf("unable to decode %08b at pos %d", b, d.di))
	}

	for _, enc := range encs {
		if in, ok := d.parse(enc); ok {
			return in
		}
	}
	panic(fmt.Sprintf("unable to find instruction encoding for %08b at index %d", b, d.di-1))
}

// read a portion of the current byte
func (d *disassembler) read(length int) byte {
	if d.cbi == 8 {
		d.curByte = d.next()
		d.cbi = 0
	}
	bd := (d.curByte >> (8 - (d.cbi + length))) & mask(length)
	d.cbi += length
	return bd
}

// next returns the next byte in the input
func (d *disassembler) next() byte {
	b := d.data[d.di]
	d.di++
	d.curByte = b
	d.cbi = 0
	return b
}

// read an 8-bit immediate
func (d *disassembler) imm8() uint16 {
	return uint16(d.next())
}

// signed 8-bit immediate
func (d *disassembler) signedImm8() int8 {
	return int8(d.next())
}

// read full 16-bit immediate
func (d *disassembler) imm16() uint16 {
	return d.imm8() | (d.imm8() << 8)
}

func (d *disassembler) parse(enc Encoding) (Instruction, bool) {
	di, curByte, cbi := d.di, d.curByte, d.cbi
	rollback := func() {
		d.di, d.curByte, d.cbi = di, curByte, cbi
	}

	in := Instruction{
		Name:   enc.Name,
		Type:   enc.Type,
		Opcode: enc.Opcode.Opcode,
		W:      1,
	}

	for _, b := range enc.Bytes {
		for _, p := range b {
			switch pname := p.Name; pname {
			case "S":
				in.S = d.read(p.Len)
			case "D":
				in.D = d.read(p.Len)
			case "W":
				in.W = d.read(p.Len)
			case "MOD":
				in.Mod = d.read(p.Len)
			case "REG":
				in.Reg = d.read(p.Len)
			case "RM":
				in.RM = d.read(p.Len)
				switch in.Mod {
				case 0b00: // Memory mode, no displacement *
					if in.RM == 0b110 { // * special case
						in.Displacement16 = int16(d.imm16())
					}
				case 0b01: // Memory mode, 8-bit displacement
					// If the displacement it 1 byte, then if needs to be sign-extended to 16-bit.
					in.Displacement8 = d.signedImm8()
				case 0b10: // Memory mode, 16-bit displacement
					in.Displacement16 = int16(d.imm16())
				}
			case "SR":
				in.SR = d.read(p.Len)
			case "DATAW":
				if in.W > 0 && in.S == 0 {
					in.Data = d.imm16()
				} else {
					in.Data = d.imm8()
				}
			case "DATA":
				in.Data = d.imm8()
			case "ADDR":
				if in.W > 0 {
					in.Data = d.imm16()
				} else {
					in.Data = d.imm8()
				}
			case "DISP":
				// Ignore
			default:
				if !p.IsConst {
					panic("p is not a constant")
				}

				// Constant
				c := d.read(p.Len)

				// Constant not correct, likely not the correct instruction
				if c != p.Const {
					// fmt.Printf("error: found constant %v that didn't match %0b\n", p, c)
					// Rollback to the original index in the data if we don't find a matching
					// instruction
					rollback()
					return Instruction{}, false
				}
			}
		}
	}
	return in, true
}

type Instruction struct {
	Name           string
	Type           string
	Opcode         byte
	D              byte
	W              byte
	S              byte
	Mod            byte
	Reg            byte
	RM             byte
	SR             byte
	Data           uint16
	Displacement8  int8
	Displacement16 int16
}

type Operand struct {
	Reg1         string
	Reg2         string
	SR           string
	Imm          uint16
	Displacement int16
	Ptr          bool
}

func (i Instruction) Operands() []Operand {
	if *debugFlag {
		fmt.Printf("inst=%#v\n", i)
	}
	var ops []Operand
	for _, typ := range strings.Split(i.Type, "__") {
		switch typ {
		case "REG":
			ops = append(ops, i.operandReg())
		case "RM":
			ops = append(ops, i.operandRM())
		case "IMM", "DATA":
			ops = append(ops, Operand{Imm: i.Data})
		case "MEM":
			ops = append(ops, Operand{Imm: i.Data, Ptr: true})
		case "ACC":
			ops = append(ops, i.operandAcc())
		case "DX":
			ops = append(ops, Operand{Reg1: "dx"})
		case "SR":
			sr := "es"
			switch i.SR {
			case 0b01:
				sr = "cs"
			case 0b10:
				sr = "ss"
			case 0b11:
				sr = "ds"
			}
			ops = append(ops, Operand{SR: sr})
		case "":
			// Do nothing
		default:
			panic(fmt.Sprintf("type %s not implemented", typ))
		}
	}
	if i.D > 0 {
		ops[0], ops[1] = ops[1], ops[0]
	}
	return ops
}

func (i Instruction) operandReg() Operand {
	return Operand{Reg1: formatReg(i.Reg, i.W)}
}

func (i Instruction) operandRM() Operand {
	r1, r2, disp, ptr := formatRM(i.Mod, i.RM, i.W, i.Displacement8, i.Displacement16)
	return Operand{
		Reg1:         r1,
		Reg2:         r2,
		Displacement: disp,
		Ptr:          ptr,
	}
}

func (i Instruction) operandAcc() Operand {
	if i.W > 0 {
		return Operand{Reg1: "ax"}
	}
	return Operand{Reg1: "al"}
}

func (i Instruction) String() string {
	var sb strings.Builder

	sb.WriteString(i.Name)

	wasPtr := false
	ops := i.Operands()
	for oi, o := range ops {
		if len(ops) == 1 && o.Ptr {
			if i.W > 0 {
				sb.WriteString(" word")
			} else {
				sb.WriteString(" byte")
			}
		}
		if oi > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(" ")
		if o.Ptr {
			sb.WriteString("[")
			wasPtr = true
		}
		if o.SR != "" {
			sb.WriteString(o.SR)
		}
		if o.Reg1 != "" {
			sb.WriteString(o.Reg1)
		}
		if o.Reg2 != "" {
			sb.WriteString(" + ")
			sb.WriteString(o.Reg2)
		}
		if o.Displacement > 0 {
			if o.Reg1 != "" {
				sb.WriteString(" + ")
			}
			sb.WriteString(fmt.Sprintf("%d", o.Displacement))
		} else if o.Displacement < 0 {
			if o.Reg1 != "" {
				sb.WriteString(" - ")
			}
			sb.WriteString(fmt.Sprintf("%d", -o.Displacement))
		}

		if o.Imm > 0 {
			if wasPtr && !o.Ptr {
				if i.W > 0 {
					sb.WriteString("word ")
				} else {
					sb.WriteString("byte ")
				}
			}
			sb.WriteString(fmt.Sprintf("%d", o.Imm))
		}

		if o.Ptr {
			sb.WriteString("]")
		}
	}
	return sb.String()
}

func formatRM(mod, rm, w byte, disp8 int8, disp16 int16) (string, string, int16, bool) {
	if mod == 0b11 {
		return formatReg(rm, w), "", 0, false
	}
	if mod == 0b00 && rm == 0b110 {
		return "", "", disp16, true
	}

	var (
		r1   string
		r2   string
		disp int16
	)

	switch rm {
	case 0b000:
		r1 = "bx"
		r2 = "si"
	case 0b001:
		r1 = "bx"
		r2 = "di"
	case 0b010:
		r1 = "bp"
		r2 = "si"
	case 0b011:
		r1 = "bp"
		r2 = "di"
	case 0b100:
		r1 = "si"
	case 0b101:
		r1 = "di"
	case 0b110:
		r1 = "bp"
	case 0b111:
		r1 = "bx"
	}

	switch mod {
	case 0b01:
		disp = int16(disp8)
	case 0b10:
		disp = disp16
	}

	return r1, r2, disp, true
}

func formatReg(r, w byte) string {
	format := func(r, rw string) string {
		if w == 0 {
			return r
		}
		return rw
	}
	switch r {
	case 0b000:
		return format("al", "ax")
	case 0b001:
		return format("cl", "cx")
	case 0b010:
		return format("dl", "dx")
	case 0b011:
		return format("bl", "bx")
	case 0b100:
		return format("ah", "sp")
	case 0b101:
		return format("ch", "bp")
	case 0b110:
		return format("dh", "si")
	case 0b111:
		return format("bh", "di")
	}
	return ""
}
