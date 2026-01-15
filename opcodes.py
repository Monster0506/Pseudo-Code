from enum import Enum, auto


class OpCode(Enum):
    """Machine-level operation codes"""
    ASN = auto()
    CAL = auto()
    AOP = auto()
    COM = auto()
    IDX = auto()
    DRF = auto()
    RET = auto()
    SKP = auto()
    JMP = auto()


class Instruction:
    """Represents a single instruction with opcode and operands"""
    
    def __init__(self, opcode: OpCode, *operands):
        self.opcode = opcode
        self.operands = operands
    
    def __repr__(self):
        operands_str = " ".join(str(op) for op in self.operands)
        return f"{self.opcode.name} {operands_str}".strip()
