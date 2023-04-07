package main

import (
	"math/bits"
	"strings"
)

type simulator struct {
	ip int

	regs  [12]uint16
	flags simFlags
}

func (s *simulator) exec(ip int, in Instruction) {
	s.ip = ip
	ops := in.Operands()

	reg := ops[0].Reg1
	switch {
	case ops[0].SR != "":
		reg = ops[0].SR
	}

	data := ops[1].Imm
	switch {
	case ops[1].Ptr:
		panic("don't handle ptr yet")
	case ops[1].Reg1 != "":
		data = s.getReg(ops[1].Reg1)
	case ops[1].SR != "":
		data = s.getReg(ops[1].SR)
	}

	var (
		result   uint16
		setFlags bool
	)

	switch in.Name {
	case "mov":
		s.setReg(reg, data)
	case "cmp":
		r1 := s.getReg(reg)
		result = r1 - data
		setFlags = true
	case "sub":
		r1 := s.getReg(reg)
		result = r1 - data
		setFlags = true
		s.setReg(reg, result)
	case "add":
		r1 := s.getReg(reg)
		result = r1 + data
		setFlags = true
		s.setReg(reg, result)
	}

	if setFlags {
		flags := simFlags(0)
		// ZF is set when result is zero
		if result == 0 {
			flags.set(flagZF)
		}
		// SF is set when result is signed
		// TODO: Not sure how this works with 8-bit numbers...
		if result&0xf000 > 0 {
			flags.set(flagSF)
		}
		// PF is set when the result has even parity; an even-number of
		// 1 bits.
		if bits.OnesCount16(result)&1 == 0 {
			flags.set(flagPF)
		}

		/* TODO (Page 22 of manual):
		- AF: is set when there has been a carry out of the low nibble into the high nibble.
		- CF: is set when there has been a carry out of the high-order bit result of the 8- or 16-bit
		- OF: is set when an arthimetic overflow has occured
		*/

		s.flags = flags
	}
}

func (s *simulator) getReg(reg string) uint16 {
	r := regLookup[reg]

	// TODO: This is stupid, but I can't work out how to easily
	// reverse bits in Go...
	mask := r.mask
	switch mask {
	case 0xff00:
		mask = 0x00ff
	case 0x00ff:
		mask = 0xff00
	}
	return (s.regs[r.index] & mask) >> r.shift
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
	s.setReg(reg1, s.getReg(reg2))
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

type simFlags uint16

const (
	flagCF simFlags = 2 << iota
	flagPF
	flagAF
	flagZF
	flagSF
	flagOF
	flagIF
	flagDF
	flagTF
)

func (sf *simFlags) set(flags ...simFlags) {
	for _, f := range flags {
		*sf |= f
	}
}

func (sf simFlags) isSet(flag simFlags) bool {
	return sf&flag == flag
}

func (sf simFlags) String() string {
	var sb strings.Builder

	flagNames := map[string]simFlags{
		"C": flagCF,
		"P": flagPF,
		"A": flagAF,
		"Z": flagZF,
		"S": flagSF,
		"O": flagOF,
		"I": flagIF,
		"D": flagDF,
		"T": flagTF,
	}
	for fn, flag := range flagNames {
		if sf.isSet(flag) {
			sb.WriteString(fn)
		}
	}

	return sb.String()
}
