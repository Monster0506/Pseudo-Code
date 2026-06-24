package main

import "strings"

type OpCode int

const (
	ASN OpCode = iota
	CAL
	AOP
	COM
	IDX
	RET
	SKP
	SKPT
	JMP
	PRT
)

func (o OpCode) String() string {
	switch o {
	case ASN:
		return "ASN"
	case CAL:
		return "CAL"
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
	case SKPT:
		return "SKPT"
	case JMP:
		return "JMP"
	case PRT:
		return "PRT"
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
