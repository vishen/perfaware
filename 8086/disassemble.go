package main

import (
	"fmt"
	"strings"
)

type disassembler struct {
	data []byte
	di   int

	curByte byte
	cbi     int
}

func (d *disassembler) nextInstruction() Instruction {
	start := d.di
	b := d.next()

	var flags InstructionFlags

	// Look for any prefix instructions
	found := true
	for found {
		switch b {
		case 0b11110010:
			flags |= FlagRepeat
		case 0b11110011:
			flags |= FlagRepeatZ
		case 0b00100110:
			flags |= FlagESOverride
		case 0b00101110:
			flags |= FlagCSOverride
		case 0b00110110:
			flags |= FlagSSOverride
		case 0b00111110:
			flags |= FlagDSOverride
		case 0b11110000:
			flags |= FlagLock
		default:
			found = false
		}
		if !found {
			break
		}
		b = d.next()
	}

	encs := encoder.Decode(b)
	if len(encs) == 0 {
		panic(fmt.Sprintf("unable to decode %08b at pos %d", b, d.di))
	}

	for _, enc := range encs {
		if in, ok := d.parse(enc); ok {
			in.Length = d.di - start
			in.Flags = flags
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
			case "V":
				in.V = d.read(p.Len)
			case "Z":
				in.Z = d.read(p.Len)
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
			case "JUMP":
				in.JumpTarget = d.signedImm8()
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

type InstructionFlags int

const (
	FlagRepeat InstructionFlags = 1 << iota
	FlagRepeatZ
	FlagLock
	FlagESOverride
	FlagCSOverride
	FlagSSOverride
	FlagDSOverride
)

type Instruction struct {
	Name           string
	Type           string
	Opcode         byte
	D              byte
	W              byte
	S              byte
	V              byte
	Z              byte
	Mod            byte
	Reg            byte
	RM             byte
	SR             byte
	Data           uint16
	Displacement8  int8
	Displacement16 int16
	JumpTarget     int8
	Flags          InstructionFlags
	Length         int // Length of this instruction in bytes
}

func (i Instruction) FlagSet(f InstructionFlags) bool {
	return i.Flags&f == f
}

type Operand struct {
	Reg1         string
	Reg2         string
	SR           string
	Imm          uint16
	ImmSet       bool // TODO: This is stupid. Need a better way to know if a zero-value was set.
	Displacement int16
	JumpTarget   int8
	Ptr          bool
	UnknownSize  bool
}

func (i Instruction) Operands() []Operand {
	var ops []Operand
	for _, typ := range strings.Split(i.Type, "__") {
		switch typ {
		case "REG":
			ops = append(ops, i.operandReg())
		case "RM":
			ops = append(ops, i.operandRM())
		case "IMM", "DATA":
			ops = append(ops, Operand{Imm: i.Data, ImmSet: true})
		case "JUMP":
			ops = append(ops, Operand{JumpTarget: i.JumpTarget})
		case "MEM":
			ops = append(ops, Operand{Imm: i.Data, ImmSet: true, Ptr: true})
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
		case "V":
			if i.V == 0 {
				ops = append(ops, Operand{Imm: 1, UnknownSize: true})
			} else {
				ops = append(ops, Operand{Reg1: "cl", UnknownSize: true})
			}
		case "":
			// Do nothing
		default:
			panic(fmt.Sprintf("type %s not implemented", typ))
		}
	}
	if i.D > 0 {
		ops[0], ops[1] = ops[1], ops[0]
	}
	if *debugFlag {
		fmt.Printf("inst=%#v\n", i)
		for _, o := range ops {
			fmt.Printf("> op1: %#v\n", o)
		}
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

	if i.FlagSet(FlagRepeat) || i.FlagSet(FlagRepeatZ) {
		sb.WriteString("rep ")
	}
	if i.FlagSet(FlagLock) {
		sb.WriteString("lock ")
	}

	sb.WriteString(i.Name)

	ops := i.Operands()

	knownSize := false
	for _, o := range ops {
		if !o.Ptr && o.Reg1 != "" && !o.UnknownSize {
			knownSize = true
		}
	}

	for oi, o := range ops {
		if oi > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(" ")
		if o.Ptr {
			if !knownSize {
				if i.W > 0 {
					sb.WriteString("word ")
				} else {
					sb.WriteString("byte ")
				}
			}
			switch {
			case i.FlagSet(FlagESOverride):
				sb.WriteString("es:")
			case i.FlagSet(FlagCSOverride):
				sb.WriteString("cs:")
			case i.FlagSet(FlagSSOverride):
				sb.WriteString("ss:")
			case i.FlagSet(FlagDSOverride):
				sb.WriteString("ds:")
			}
			sb.WriteString("[")
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
				sb.WriteString(fmt.Sprintf(" - %d", -o.Displacement))
			} else {
				// TODO: Don't know if this is correct, but don't think a displacement on its own can
				// ever be negative?
				sb.WriteString(fmt.Sprintf("%d", uint16(o.Displacement)))
			}
		}

		if o.ImmSet {
			sb.WriteString(fmt.Sprintf("%d", o.Imm))
		}

		if o.JumpTarget > 0 {
			sb.WriteString(fmt.Sprintf("$+%d", o.JumpTarget))
		} else if o.JumpTarget < 0 {
			sb.WriteString(fmt.Sprintf("$%d", o.JumpTarget))
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
