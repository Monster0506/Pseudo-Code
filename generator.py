from opcodes import OpCode, Instruction
from parser import (
    ASTNode,
    Literal,
    Identifier,
    ArrayLiteral,
    ArrayAccess,
    FunctionCall,
    BinaryOp,
    UnaryOp,
    Assignment,
    IfStatement,
    WhileLoop,
    ForLoop,
    RepeatUntilLoop,
    PrintStatement,
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
        self.function_table = {}  # name -> (start_addr, [param_names])
        self.entry_params = []  # params of the first / entry function

    def new_temp(self) -> str:
        temp = f"t{self.temp_counter}"
        self.temp_counter += 1
        return temp

    def new_label(self) -> str:
        label = f"L{self.label_counter}"
        self.label_counter += 1
        return label

    def emit(self, opcode: OpCode, *operands):
        self.instructions.append(Instruction(opcode, *operands))

    def generate(self, ast: Block) -> list[Instruction]:
        self.instructions = []
        self.temp_counter = 0
        self.label_counter = 0
        self.function_table = {}
        self.entry_params = []

        functions = []
        bare_stmts = []
        for stmt in ast.statements:
            if isinstance(stmt, FunctionStatement):
                functions.append(stmt)
            else:
                bare_stmts.append(stmt)

        if functions:
            # First function is the entry point; compile its body first
            entry = functions[0]
            self.entry_params = [p.name for p in entry.param_ids]
            entry_start = len(self.instructions)
            self.function_table[entry.name.name] = (entry_start, self.entry_params)
            self.visit(entry.body)
            self.emit(OpCode.RET)

            # Helper functions follow in the instruction stream
            for func in functions[1:]:
                func_start = len(self.instructions)
                params = [p.name for p in func.param_ids]
                self.function_table[func.name.name] = (func_start, params)
                self.visit(func.body)
                self.emit(OpCode.RET)

        # Bare top-level statements (no Algorithm wrapper)
        for stmt in bare_stmts:
            self.visit(stmt)

        return self.instructions

    def visit_block(self, node: Block):
        for stmt in node.statements:
            self.visit(stmt)

    def visit(self, node: ASTNode) -> str | None:
        if isinstance(node, Literal):
            return self.visit_literal(node)
        elif isinstance(node, Identifier):
            return self.visit_identifier(node)
        elif isinstance(node, ArrayAccess):
            return self.visit_array_access(node)
        elif isinstance(node, FunctionCall):
            return self.visit_function_call(node)
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
        elif isinstance(node, RepeatUntilLoop):
            return self.visit_repeat_until_loop(node)
        elif isinstance(node, PrintStatement):
            return self.visit_print_statement(node)
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
        return str(node.value)

    def visit_identifier(self, node: Identifier) -> str:
        return node.name

    def visit_array_access(self, node: ArrayAccess) -> str:
        array_name = (
            node.array.name
            if isinstance(node.array, Identifier)
            else self.visit(node.array)
        )
        index = self.visit(node.index)
        temp = self.new_temp()
        self.emit(OpCode.IDX, array_name, index, temp)
        return temp

    def visit_function_call(self, node: FunctionCall) -> str:
        arg_vals = [self.visit(arg) for arg in node.args]
        temp = self.new_temp()
        self.emit(OpCode.CAL, node.name, *arg_vals, temp)
        return temp

    def visit_binary_op(self, node: BinaryOp) -> str:
        left = self.visit(node.left)
        right = self.visit(node.right)
        temp = self.new_temp()

        if node.operator in ["+", "-", "*", "/", "mod"]:
            self.emit(OpCode.AOP, node.operator, left, right, temp)
        elif node.operator in ["=", "<", ">", "<=", ">=", "!="]:
            self.emit(OpCode.COM, node.operator, left, right, temp)
        elif node.operator in ["and", "or"]:
            self.emit(OpCode.COM, node.operator, left, right, temp)
        else:
            raise ValueError(f"Unknown operator: {node.operator}")

        return temp

    def visit_unary_op(self, node: UnaryOp) -> str:
        operand = self.visit(node.operand)
        temp = self.new_temp()
        # 3-operand AOP signals unary operation to the VM
        self.emit(OpCode.AOP, node.operator, operand, temp)
        return temp

    def visit_assignment(self, node: Assignment) -> None:
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
        self.visit(node.condition)

        skp_idx = len(self.instructions)
        self.emit(OpCode.SKP, 0)

        then_start = len(self.instructions)
        self.visit(node.then_block)

        if node.else_block:
            jmp_idx = len(self.instructions)
            self.emit(OpCode.JMP, 0)

            self.visit(node.else_block)

            then_size = jmp_idx - then_start + 1
            self.instructions[skp_idx] = Instruction(OpCode.SKP, then_size)

            end_idx = len(self.instructions)
            self.instructions[jmp_idx] = Instruction(OpCode.JMP, end_idx)
        else:
            then_size = len(self.instructions) - then_start
            self.instructions[skp_idx] = Instruction(OpCode.SKP, then_size)

    def visit_while_loop(self, node: WhileLoop) -> None:
        loop_start = len(self.instructions)

        self.visit(node.condition)

        skp_idx = len(self.instructions)
        self.emit(OpCode.SKP, 0)

        body_start = len(self.instructions)
        self.visit(node.body)

        self.emit(OpCode.JMP, loop_start)

        body_size = len(self.instructions) - body_start
        self.instructions[skp_idx] = Instruction(OpCode.SKP, body_size)

    def visit_for_loop(self, node: ForLoop) -> None:
        """for VAR <- INIT (to|downto) END [by STEP] do BODY end"""
        self.visit(node.assignment)

        loop_var = (
            node.assignment.target.name
            if isinstance(node.assignment.target, Identifier)
            else str(node.assignment.target)
        )

        loop_start = len(self.instructions)
        end_value = self.visit(node.end)

        cmp_op = "<=" if node.direction == "to" else ">="
        temp_cmp = self.new_temp()
        self.emit(OpCode.COM, cmp_op, loop_var, end_value, temp_cmp)

        skp_idx = len(self.instructions)
        self.emit(OpCode.SKP, 0)

        body_start = len(self.instructions)
        self.visit(node.body)

        # Increment / decrement
        temp_inc = self.new_temp()
        if node.step is not None:
            step_val = self.visit(node.step)
            self.emit(OpCode.AOP, "+", loop_var, step_val, temp_inc)
        elif node.direction == "to":
            self.emit(OpCode.AOP, "+", loop_var, "1", temp_inc)
        else:
            self.emit(OpCode.AOP, "-", loop_var, "1", temp_inc)

        self.emit(OpCode.ASN, loop_var, temp_inc)
        self.emit(OpCode.JMP, loop_start)

        body_size = len(self.instructions) - body_start
        self.instructions[skp_idx] = Instruction(OpCode.SKP, body_size)

    def visit_repeat_until_loop(self, node: RepeatUntilLoop) -> None:
        """repeat BODY until CONDITION: loop while condition is false"""
        body_start = len(self.instructions)
        self.visit(node.body)

        # Evaluate condition; last_cmp is set by the COM/AOP it generates
        self.visit(node.condition)

        # SKPT 1: skip the JMP if condition is TRUE (exit loop)
        self.emit(OpCode.SKPT, 1)
        # JMP back to body_start if condition is FALSE (keep looping)
        self.emit(OpCode.JMP, body_start)

    def visit_print_statement(self, node: PrintStatement) -> None:
        value = self.visit(node.expr)
        self.emit(OpCode.PRT, value)

    def visit_return_statement(self, node: ReturnStatement) -> None:
        if node.value:
            value = self.visit(node.value)
            self.emit(OpCode.RET, value)
        else:
            self.emit(OpCode.RET)

    def visit_function_statement(self, node: FunctionStatement) -> None:
        # Called when a FunctionStatement appears inside another body (unusual).
        # Just compile the inner body inline.
        self.visit(node.body)
        self.emit(OpCode.RET)
