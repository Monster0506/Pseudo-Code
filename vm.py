from opcodes import OpCode


class VM:
    """Virtual Machine to execute generated opcodes"""

    def __init__(self, instructions):
        self.instructions = instructions
        self.pc = 0
        self.variables = {}
        self.last_cmp = False
        self.return_value = None
        self.inputs = self._detect_inputs()

    def _detect_inputs(self):
        """Detect which variables are read before being written (i.e., inputs)"""
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
                if "[" not in target:
                    written.add(target.split("[")[0] if "[" in target else target)

            elif opcode == OpCode.AOP:

                for operand in [operands[1], operands[2]]:
                    op = str(operand)
                    if op.isalpha() and op not in written and not op.startswith("t"):
                        inputs[op] = None

                written.add(str(operands[3]))

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
        """Get value - either literal or variable"""
        operand = str(operand)
        try:
            return int(operand)
        except (ValueError, TypeError):
            return self.variables.get(operand, 0)

    def set_indexed(self, target, value):
        """Set array element: A[i] = value"""
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
        """Execute single instruction, return PC delta (0 for normal, n for jumps)"""
        opcode = instr.opcode
        operands = instr.operands

        if opcode == OpCode.ASN:

            target, value = operands[0], self.get_value(operands[1])
            self.set_indexed(target, value)

        elif opcode == OpCode.AOP:

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
            elif op == "=":

                self.variables[result] = right_val

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
            else:
                cmp_result = False

            self.variables[result] = cmp_result
            self.last_cmp = cmp_result

        elif opcode == OpCode.IDX:

            array_name, index, result = operands
            index_val = self.get_value(index)
            arr = self.variables.get(array_name, [])
            self.variables[result] = arr[index_val] if index_val < len(arr) else 0

        elif opcode == OpCode.SKP:

            n = int(operands[0])
            if not self.last_cmp:
                return n

        elif opcode == OpCode.JMP:

            target = int(operands[0])
            return target - self.pc - 1

        elif opcode == OpCode.RET:

            if operands:
                self.return_value = self.get_value(operands[0])
            else:
                self.return_value = None
            return len(self.instructions)

        return 0

    def run(self, **initial_vars):
        """Run program with initial variables"""
        self.variables.update(initial_vars)
        self.pc = 0
        self.last_cmp = False
        self.return_value = None

        while self.pc < len(self.instructions):
            pc_delta = self.execute(self.instructions[self.pc])
            self.pc += 1 + pc_delta

        return self.return_value
