import math
from opcodes import OpCode


# Built-in functions callable from pseudocode
_BUILTINS = {
    "length": lambda args: len(args[0]) if isinstance(args[0], list) else 1,
    "floor": lambda args: math.floor(args[0]),
    "ceil": lambda args: math.ceil(args[0]),
    "abs": lambda args: abs(args[0]),
    "sqrt": lambda args: math.sqrt(args[0]),
    "min": lambda args: min(args),
    "max": lambda args: max(args),
}


class VM:
    """Virtual Machine to execute generated opcodes"""

    def __init__(self, instructions, function_table=None):
        self.instructions = instructions
        self.pc = 0
        self.variables = {}
        self.last_cmp = False
        self.return_value = None
        self.call_stack = []  # (return_pc, saved_vars, result_target)
        self.function_table = function_table or {}
        self.inputs = self._detect_inputs()

    def _detect_inputs(self):
        """Detect variables that are read before being written (i.e., inputs)."""
        written = set()
        inputs = {}

        for instr in self.instructions:
            opcode = instr.opcode
            operands = instr.operands

            if opcode == OpCode.ASN:
                value = str(operands[1])
                if (
                    value.isalpha()
                    and value not in written
                    and not value.startswith("t")
                ):
                    inputs[value] = None

                target = str(operands[0])
                written.add(target.split("[")[0])

            elif opcode == OpCode.AOP:
                # Unary (3 operands) or binary (4 operands)
                read_ops = (
                    [operands[1]] if len(operands) == 3 else [operands[1], operands[2]]
                )
                for operand in read_ops:
                    op = str(operand)
                    if op.isalpha() and op not in written and not op.startswith("t"):
                        inputs[op] = None
                written.add(str(operands[-1]))

            elif opcode == OpCode.COM:
                for operand in [operands[1], operands[2]]:
                    op = str(operand)
                    if op.isalpha() and op not in written and not op.startswith("t"):
                        inputs[op] = None
                written.add(str(operands[3]))

            elif opcode == OpCode.IDX:
                for operand in [operands[0], operands[1]]:
                    op = str(operand)
                    if op.isalpha() and op not in written and not op.startswith("t"):
                        inputs[op] = None
                written.add(str(operands[2]))

        return inputs

    def get_value(self, operand):
        """Resolve an operand to its runtime value."""
        operand = str(operand)
        if operand == "true":
            return True
        if operand == "false":
            return False
        if operand in ("NIL", "null"):
            return None
        if operand == "infinity":
            return float("inf")
        # Strip string literal quotes
        if (operand.startswith('"') and operand.endswith('"')) or (
            operand.startswith("'") and operand.endswith("'")
        ):
            return operand[1:-1]
        try:
            return int(operand)
        except (ValueError, TypeError):
            return self.variables.get(operand, 0)

    def set_indexed(self, target, value):
        """Write to a plain variable or an array element (A[i])."""
        if "[" in target:
            arr_name = target.split("[")[0]
            index_str = target.split("[")[1].rstrip("]")
            index = self.get_value(index_str)
            if arr_name not in self.variables:
                self.variables[arr_name] = []
            while len(self.variables[arr_name]) <= index:
                self.variables[arr_name].append(0)
            self.variables[arr_name][index] = value
        else:
            self.variables[target] = value

    def execute(self, instr):
        """Execute one instruction; return PC delta (0 = advance by 1)."""
        opcode = instr.opcode
        operands = instr.operands

        if opcode == OpCode.ASN:
            target, value = operands[0], self.get_value(operands[1])
            self.set_indexed(target, value)

        elif opcode == OpCode.AOP:
            op = operands[0]
            if len(operands) == 3:
                # Unary operation (currently only "not")
                val = self.get_value(operands[1])
                result = str(operands[2])
                if op == "not":
                    result_val = not bool(val)
                    self.variables[result] = result_val
                    self.last_cmp = result_val  # keep last_cmp in sync for control flow
                else:
                    raise ValueError(f"Unknown unary op: {op}")
            else:
                # Binary arithmetic
                op, left, right, result = operands
                left_val = self.get_value(left)
                right_val = self.get_value(right)

                if op == "+":
                    self.variables[result] = left_val + right_val
                elif op == "-":
                    self.variables[result] = left_val - right_val
                elif op == "*":
                    self.variables[result] = left_val * right_val
                elif op == "/":
                    self.variables[result] = left_val // right_val
                elif op == "mod":
                    self.variables[result] = left_val % right_val
                else:
                    raise ValueError(f"Unknown arithmetic op: {op}")

        elif opcode == OpCode.COM:
            op, left, right, result = operands
            left_val = self.get_value(left)
            right_val = self.get_value(right)

            if op == "<":
                cmp_result = left_val < right_val
            elif op == ">":
                cmp_result = left_val > right_val
            elif op == "<=":
                cmp_result = left_val <= right_val
            elif op == ">=":
                cmp_result = left_val >= right_val
            elif op == "=":
                cmp_result = left_val == right_val
            elif op == "!=":
                cmp_result = left_val != right_val
            elif op == "and":
                cmp_result = bool(left_val) and bool(right_val)
            elif op == "or":
                cmp_result = bool(left_val) or bool(right_val)
            else:
                cmp_result = False

            self.variables[result] = cmp_result
            self.last_cmp = cmp_result

        elif opcode == OpCode.IDX:
            array_name, index, result = operands
            index_val = self.get_value(index)
            arr = self.variables.get(array_name, [])
            self.variables[result] = arr[index_val] if index_val < len(arr) else 0

        elif opcode == OpCode.CAL:
            func_name = str(operands[0])
            arg_operands = operands[1:-1]
            result_target = str(operands[-1])
            arg_values = [self.get_value(a) for a in arg_operands]

            if func_name in _BUILTINS:
                result = _BUILTINS[func_name](arg_values)
                self.variables[result_target] = result
            else:
                if func_name not in self.function_table:
                    raise RuntimeError(f"Unknown function: {func_name!r}")
                start_idx, param_names = self.function_table[func_name]
                # Push caller frame
                self.call_stack.append(
                    (self.pc + 1, dict(self.variables), result_target)
                )
                # New scope: only function parameters
                self.variables = {
                    name: val for name, val in zip(param_names, arg_values)
                }
                return start_idx - self.pc - 1

        elif opcode == OpCode.PRT:
            print(self.get_value(operands[0]))

        elif opcode == OpCode.SKP:
            n = int(operands[0])
            if not self.last_cmp:
                return n

        elif opcode == OpCode.SKPT:
            n = int(operands[0])
            if self.last_cmp:
                return n

        elif opcode == OpCode.JMP:
            target = int(operands[0])
            return target - self.pc - 1

        elif opcode == OpCode.RET:
            ret_val = self.get_value(operands[0]) if operands else None

            if self.call_stack:
                return_pc, saved_vars, result_target = self.call_stack.pop()
                self.variables = saved_vars
                if result_target and ret_val is not None:
                    self.variables[result_target] = ret_val
                return return_pc - self.pc - 1
            else:
                self.return_value = ret_val
                return len(self.instructions)  # halt

        return 0

    def run(self, **initial_vars):
        """Run program with initial variable bindings."""
        self.variables.update(initial_vars)
        self.pc = 0
        self.last_cmp = False
        self.return_value = None

        while self.pc < len(self.instructions):
            pc_delta = self.execute(self.instructions[self.pc])
            self.pc += 1 + pc_delta

        return self.return_value
