from opcodes import OpCode, Instruction
from parser import (
    ASTNode,
    Literal,
    Identifier,
    ArrayLiteral,
    ArrayAccess,
    BinaryOp,
    UnaryOp,
    Assignment,
    IfStatement,
    WhileLoop,
    ForLoop,
    ReturnStatement,
    Block,
    FunctionStatement,
)


class Generator:
    """Generates machine-level opcodes from AST"""

    def __init__(self):
        self.instructions = []
        self.temp_counter = 0
        self.label_counter = 0

    def new_temp(self) -> str:
        """Allocate a new temporary register"""
        temp = f"t{self.temp_counter}"
        self.temp_counter += 1
        return temp

    def new_label(self) -> str:
        """Generate a new label for jumps"""
        label = f"L{self.label_counter}"
        self.label_counter += 1
        return label

    def emit(self, opcode: OpCode, *operands):
        """Emit a new instruction"""
        self.instructions.append(Instruction(opcode, *operands))

    def generate(self, ast: Block) -> list[Instruction]:
        """Generate opcodes from AST"""
        self.instructions = []
        self.temp_counter = 0
        self.label_counter = 0

        self.visit_block(ast)
        return self.instructions

    def visit_block(self, node: Block):
        """Visit a block of statements"""
        for stmt in node.statements:
            self.visit(stmt)

    def visit(self, node: ASTNode) -> str | None:
        """Visit an AST node and return the result location (temp/var name)"""
        if isinstance(node, Literal):
            return self.visit_literal(node)
        elif isinstance(node, Identifier):
            return self.visit_identifier(node)
        elif isinstance(node, ArrayAccess):
            return self.visit_array_access(node)
        elif isinstance(node, BinaryOp):
            return self.visit_binary_op(node)
        elif isinstance(node, UnaryOp):
            return self.visit_unary_op(node)
        elif isinstance(node, Assignment):
            return self.visit_assignment(node)
        elif isinstance(node, IfStatement):
            return self.visit_if_statement(node)
        elif isinstance(node, WhileLoop):
            return self.visit_while_loop(node)
        elif isinstance(node, ForLoop):
            return self.visit_for_loop(node)
        elif isinstance(node, ReturnStatement):
            return self.visit_return_statement(node)
        elif isinstance(node, FunctionStatement):
            return self.visit_function_statement(node)
        elif isinstance(node, Block):
            self.visit_block(node)
            return None
        else:
            raise ValueError(f"Unknown AST node type: {type(node)}")

    def visit_literal(self, node: Literal) -> str:
        """Visit a literal - return the literal value directly"""
        return str(node.value)

    def visit_identifier(self, node: Identifier) -> str:
        """Visit an identifier - return the variable name directly"""
        return node.name

    def visit_array_access(self, node: ArrayAccess) -> str:
        """Visit array access - emit IDX opcode"""
        array_name = (
            node.array.name
            if isinstance(node.array, Identifier)
            else self.visit(node.array)
        )
        index = self.visit(node.index)
        temp = self.new_temp()
        self.emit(OpCode.IDX, array_name, index, temp)
        return temp

    def visit_binary_op(self, node: BinaryOp) -> str:
        """Visit binary operation - emit AOP or COM"""
        left = self.visit(node.left)
        right = self.visit(node.right)
        temp = self.new_temp()

        if node.operator in ["+", "-", "*", "/"]:
            self.emit(OpCode.AOP, node.operator, left, right, temp)
        elif node.operator in ["=", "<", ">", "<=", ">=", "!="]:
            self.emit(OpCode.COM, node.operator, left, right, temp)
        elif node.operator in ["and", "or"]:

            self.emit(OpCode.COM, node.operator, left, right, temp)
        else:
            raise ValueError(f"Unknown operator: {node.operator}")

        return temp

    def visit_unary_op(self, node: UnaryOp) -> str:
        """Visit unary operation (e.g., 'not')"""
        operand = self.visit(node.operand)
        temp = self.new_temp()
        self.emit(OpCode.AOP, node.operator, operand, temp)
        return temp

    def visit_assignment(self, node: Assignment) -> None:
        """Visit assignment - emit ASN"""
        value = self.visit(node.value)

        if isinstance(node.target, Identifier):
            self.emit(OpCode.ASN, node.target.name, value)
        elif isinstance(node.target, ArrayAccess):

            array_name = (
                node.target.array.name
                if isinstance(node.target.array, Identifier)
                else self.visit(node.target.array)
            )
            index = self.visit(node.target.index)
            self.emit(OpCode.ASN, f"{array_name}[{index}]", value)
        else:
            raise ValueError(f"Invalid assignment target: {type(node.target)}")

    def visit_if_statement(self, node: IfStatement) -> None:
        """Visit if statement - emit COM + SKP to skip else block"""

        condition = self.visit(node.condition)

        skp_idx = len(self.instructions)
        self.emit(OpCode.SKP, 0)

        then_start = len(self.instructions)
        self.visit(node.then_block)

        if node.else_block:

            jmp_idx = len(self.instructions)
            self.emit(OpCode.JMP, 0)

            else_start = len(self.instructions)
            self.visit(node.else_block)

            then_size = jmp_idx - then_start + 1
            self.instructions[skp_idx] = Instruction(OpCode.SKP, then_size)

            end_idx = len(self.instructions)
            self.instructions[jmp_idx] = Instruction(OpCode.JMP, end_idx)
        else:

            then_size = len(self.instructions) - then_start
            self.instructions[skp_idx] = Instruction(OpCode.SKP, then_size)

    def visit_while_loop(self, node: WhileLoop) -> None:
        """Visit while loop - emit condition check with SKP and JMP back"""
        loop_start = len(self.instructions)

        condition = self.visit(node.condition)

        skp_idx = len(self.instructions)
        self.emit(OpCode.SKP, 0)

        body_start = len(self.instructions)
        self.visit(node.body)

        self.emit(OpCode.JMP, loop_start)

        body_size = len(self.instructions) - body_start
        self.instructions[skp_idx] = Instruction(OpCode.SKP, body_size)

    def visit_for_loop(self, node: ForLoop) -> None:
        """Visit for loop - emit initialization, condition with SKP, body, increment, JMP back"""

        self.visit(node.assignment)

        loop_start = len(self.instructions)

        loop_var = (
            node.assignment.target.name
            if isinstance(node.assignment.target, Identifier)
            else str(node.assignment.target)
        )
        end_value = self.visit(node.end)

        temp = self.new_temp()
        self.emit(OpCode.COM, "<=", loop_var, end_value, temp)

        skp_idx = len(self.instructions)
        self.emit(OpCode.SKP, 0)

        body_start = len(self.instructions)
        self.visit(node.body)

        temp_inc = self.new_temp()
        self.emit(OpCode.AOP, "+", loop_var, "1", temp_inc)
        self.emit(OpCode.ASN, loop_var, temp_inc)

        self.emit(OpCode.JMP, loop_start)

        body_size = len(self.instructions) - body_start
        self.instructions[skp_idx] = Instruction(OpCode.SKP, body_size)

    def visit_return_statement(self, node: ReturnStatement) -> None:
        """Visit return statement - emit RET"""
        if node.value:
            value = self.visit(node.value)
            self.emit(OpCode.RET, value)
        else:
            self.emit(OpCode.RET)

    def visit_function_statement(self, node: FunctionStatement) -> None:
        """Visit function definition - emit function body"""

        self.visit(node.body)

        self.emit(OpCode.RET)
