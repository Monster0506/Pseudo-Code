package main

import "strings"

type OpCode int

const (
	ASN OpCode = iota
	AOP
	COM
	IDX
	RET
	SKP
	JMP
)

func (o OpCode) String() string {
	switch o {
	case ASN:
		return "ASN"
	case AOP:
		return "AOP"
	case COM:
		return "COM"
	case IDX:
		return "IDX"
	case RET:
		return "RET"
	case SKP:
		return "SKP"
	case JMP:
		return "JMP"
	}
	return "???"
}

type Instruction struct {
	Opcode   OpCode
	Operands []string
}

func (i Instruction) String() string {
	if len(i.Operands) == 0 {
		return i.Opcode.String()
	}
	return i.Opcode.String() + " " + strings.Join(i.Operands, " ")
}
