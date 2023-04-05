package main

type simulator struct {
	ip int

	regs  [12]uint16
	flags uint16
}

func (s *simulator) exec(ip int, in Instruction) {
	s.ip = ip
	ops := in.Operands()

	switch in.Name {
	case "mov":
		switch in.Type {
		default:
			panic("unhandled type: " + in.Type)
		case "REG__IMM":
			s.setReg(ops[0].Reg1, ops[1].Imm)
		case "RM__SR":
			s.setRegToReg(ops[0].Reg1, ops[1].SR)
		case "SR__RM":
			s.setRegToReg(ops[0].SR, ops[1].Reg1)
		case "RM__REG":
			s.setRegToReg(ops[0].Reg1, ops[1].Reg1)
		}
	}
}

func (s *simulator) setReg(reg1 string, data uint16) {
	r := regLookup[reg1]
	// TODO: This is stupid. Not sure if there is a way to handle the x{l,h} register
	// mask and shifts.
	if r.mask != 0xffff {
		s.regs[r.index] = (s.regs[r.index] & r.mask) | (data << r.shift)
	} else {
		s.regs[r.index] = data
	}
}

func (s *simulator) setRegToReg(reg1 string, reg2 string) {
	r := regLookup[reg2]

	// TODO: This is stupid, but I can't work out how to easily
	// reverse bits in Go...
	mask := r.mask
	switch mask {
	case 0xff00:
		mask = 0x00ff
	case 0x00ff:
		mask = 0xff00
	}
	// dh: 0x7788 -> 0x77
	data := (s.regs[r.index] & mask) >> r.shift
	s.setReg(reg1, data)
}

var regLookup = map[string]struct {
	index uint16
	mask  uint16
	shift uint16
}{
	"ax": {0, 0xffff, 0},
	"al": {0, 0xff00, 0},
	"ah": {0, 0x00ff, 8},
	"bx": {1, 0xffff, 0},
	"bl": {1, 0xff00, 0},
	"bh": {1, 0x00ff, 8},
	"cx": {2, 0xffff, 0},
	"cl": {2, 0xff00, 0},
	"ch": {2, 0x00ff, 8},
	"dx": {3, 0xffff, 0},
	"dl": {3, 0xff00, 0},
	"dh": {3, 0x00ff, 8},

	"sp": {4, 0xffff, 0},
	"bp": {5, 0xffff, 0},
	"si": {6, 0xffff, 0},
	"di": {7, 0xffff, 0},
	"cs": {8, 0xffff, 0},
	"ds": {9, 0xffff, 0},
	"ss": {10, 0xffff, 0},
	"es": {11, 0xffff, 0},
}
