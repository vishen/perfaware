package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	_ "embed"
)

var (
	//go:embed instruction_encodings.txt
	instructionEncodings string

	encoder = NewEncoder(instructionEncodings)
)

var sizes = map[string]int{
	"D":      1,
	"W":      1,
	"S":      1,
	"MOD":    2,
	"SR":     2,
	"REG":    3,
	"RM":     3,
	"DISPLO": 8,
	"DISPHI": 8,
	"DATA":   8,
	"DATAW":  8,
	"ADDRLO": 8,
	"ADDRHI": 8,
	// TODO: These sizes should be zero, but the parser doesn't currently handle it
	"ACCDST": 8,
	"ACCSRC": 8,
}

func sizeOf(val string) int {
	if size, ok := sizes[val]; ok {
		return size
	} else {
		return len(val)
	}
}

func nameOf(val string) string {
	if _, ok := sizes[val]; ok {
		return val
	}
	return "const_" + val
}

func isConst(val string) bool {
	if _, ok := sizes[val]; !ok {
		return true
	}
	return false
}

type Encoder struct {
	rawEncoding string

	encodings []Encoding
}

type Encoding struct {
	Orig string

	Name   string
	Opcode Opcode

	Bytes [][]Part
}

type Part struct {
	Name  string
	Start int
	Len   int

	IsConst bool
	Const   byte
}

type Opcode struct {
	Opcode byte
	Len    int
}

func (o Opcode) String() string {
	return fmt.Sprintf("%d: %0b", o.Len, o.Opcode)
}

func NewEncoder(instructionEncodings string) Encoder {
	e := Encoder{rawEncoding: instructionEncodings}
	for i, encoding := range strings.Split(instructionEncodings, "\n") {
		if len(encoding) == 0 {
			continue
		}
		if encoding[0] == '#' {
			continue
		}
		enc := Encoding{Orig: encoding}

		encoding := strings.Split(encoding, " ")
		enc.Name = encoding[0]
		for _, part := range encoding[1:] {
			if len(part) == 0 {
				continue
			}

			var parts []Part
			total := 8

			pi := 0
			for _, p := range strings.Split(part, "_") {
				if enc.Opcode.Len == 0 {
					enc.Opcode.Opcode = convert(p)
					enc.Opcode.Len = len(p)
				}
				size := sizeOf(p)

				part := Part{
					Name:  nameOf(p),
					Start: pi,
					Len:   size,
				}
				if isConst(p) {
					part.IsConst = true
					part.Const = convert(p)
				}
				parts = append(parts, part)

				total -= size
				pi += size
			}
			if total != 0 {
				log.Fatalf("line %d (%s): invalid part %q doesn't equal 8 bits", i+1, encoding, part)
			}
			enc.Bytes = append(enc.Bytes, parts)
		}
		e.encodings = append(e.encodings, enc)
	}
	return e
}

func (e Encoder) Decode(b byte) []Encoding {
	var found []Encoding
	for _, e := range encoder.encodings {
		if b>>(8-e.Opcode.Len)&mask(e.Opcode.Len) == e.Opcode.Opcode {
			found = append(found, e)
		}
	}
	sort.Slice(found, func(i, j int) bool {
		return found[i].Opcode.Len > found[j].Opcode.Len
	})
	return found
}

func mask(length int) byte {
	m := byte(1)
	for i := 1; i < length; i++ {
		m = m << 1
		m |= 1
	}
	return m
}

func convert(v string) byte {
	b := byte(0)
	for i, c := range v {
		if c == '1' {
			b |= 1 << (len(v) - 1 - i)
		}
	}
	return b
}
