from enum import Enum, auto


class OpCode(Enum):
    """Machine-level operation codes"""

    ASN = auto()  # assign: ASN target value
    CAL = auto()  # call:   CAL name arg... result
    AOP = auto()  # arith:  AOP op left right result  (binary)
    #          AOP op operand result     (unary: not)
    COM = auto()  # compare: COM op left right result
    IDX = auto()  # index:  IDX array index result
    DRF = auto()  # deref (reserved)
    RET = auto()  # return: RET [value]
    SKP = auto()  # skip n instructions if last_cmp is FALSE
    SKPT = auto()  # skip n instructions if last_cmp is TRUE
    JMP = auto()  # jump:   JMP absolute_address
    PRT = auto()  # print:  PRT value


class Instruction:
    """Represents a single instruction with opcode and operands"""

    def __init__(self, opcode: OpCode, *operands):
        self.opcode = opcode
        self.operands = operands

    def __repr__(self):
        operands_str = " ".join(str(op) for op in self.operands)
        return f"{self.opcode.name} {operands_str}".strip()
