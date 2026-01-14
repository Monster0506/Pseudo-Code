from tokenizer import Token


class ASTNode:
    pass


class Literal(ASTNode):
    def __init__(self, value):
        self.value = value

    def __repr__(self):
        return f"Literal({self.value})"


class Identifier(ASTNode):
    def __init__(self, name):
        self.name = name

    def __repr__(self):
        return f"Identifier({self.name})"


class ArrayLiteral(ASTNode):
    def __init__(self, elements):
        self.elements = elements

    def __repr__(self):
        return f"ArrayLiteral({self.elements})"


class ArrayAccess(ASTNode):
    def __init__(self, array, index):
        self.array = array
        self.index = index

    def __repr__(self):
        return f"ArrayAccess({self.array}, {self.index})"


class BinaryOp(ASTNode):
    def __init__(self, left, operator, right):
        self.left = left
        self.operator = operator
        self.right = right

    def __repr__(self):
        return f"BinaryOp({self.left}, {self.operator}, {self.right})"


class UnaryOp(ASTNode):
    def __init__(self, operator, operand):
        self.operator = operator
        self.operand = operand

    def __repr__(self):
        return f"UnaryOp({self.operator}, {self.operand})"


class Assignment(ASTNode):
    def __init__(self, target, value):
        self.target = target
        self.value = value

    def __repr__(self):
        return f"Assignment({self.target}, {self.value})"


class IfStatement(ASTNode):
    def __init__(self, condition, then_block, else_block=None):
        self.condition = condition
        self.then_block = then_block
        self.else_block = else_block

    def __repr__(self):
        return f"IfStatement({self.condition}, {self.then_block}, {self.else_block})"


class WhileLoop(ASTNode):
    def __init__(self, condition, body):
        self.condition = condition
        self.body = body

    def __repr__(self):
        return f"WhileLoop({self.condition}, {self.body})"


class ForLoop(ASTNode):
    def __init__(self, assignment, end, body):
        self.assignment = assignment
        self.end = end
        self.body = body

    def __repr__(self):
        return f"ForLoop({self.assignment}, {self.end}, {self.body})"


class ReturnStatement(ASTNode):
    def __init__(self, value):
        self.value = value

    def __repr__(self):
        return f"ReturnStatement({self.value})"


class Block(ASTNode):
    def __init__(self, statements):
        self.statements = statements

    def __repr__(self):
        return f"Block({self.statements})"


class Parser:
    def __init__(self, tokens: list[tuple[str, Token]]):
        self.tokens = tokens
        self.pos = 0

    def current_token(self) -> tuple[str, Token] | None:
        if self.pos < len(self.tokens):
            return self.tokens[self.pos]
        return None

    def peek_token(self, offset=1) -> tuple[str, Token] | None:
        if self.pos + offset < len(self.tokens):
            return self.tokens[self.pos + offset]
        return None

    def advance(self):
        self.pos += 1

    def consume(self, expected_value=None) -> tuple[str, Token]:
        token = self.current_token()
        if token is None:
            raise SyntaxError("Unexpected end of input")
        if expected_value is not None and token[0] != expected_value:
            raise SyntaxError(f"Expected '{expected_value}', got '{token[0]}'")
        self.advance()
        return token

    def parse(self) -> Block:
        statements = []
        while self.current_token() is not None:
            stmt = self.parse_statement()
            if stmt:
                statements.append(stmt)
        return Block(statements)

    def parse_statement(self) -> ASTNode | None:
        token = self.current_token()
        if token is None:
            return None

        value, token_type = token

        if value == "if":
            return self.parse_if_statement()
        elif value == "while":
            return self.parse_while_loop()
        elif value == "for":
            return self.parse_for_loop()
        elif value == "return":
            return self.parse_return_statement()
        elif token_type == Token.IDENTIFIER:
            return self.parse_assignment_or_expression()
        else:
            raise SyntaxError(f"Unexpected token: {value}")

    def parse_if_statement(self) -> IfStatement:
        self.consume("if")
        condition = self.parse_expression()
        self.consume("then")

        then_block = []
        while self.current_token() and self.current_token()[0] not in [
            "else",
            "end",
        ]:
            stmt = self.parse_statement()
            if stmt:
                then_block.append(stmt)

        else_block = None
        if self.current_token() and self.current_token()[0] == "else":
            self.consume("else")
            else_block = []
            while self.current_token() and self.current_token()[0] != "end":
                stmt = self.parse_statement()
                if stmt:
                    else_block.append(stmt)

        self.consume("end")
        return IfStatement(
            condition, Block(then_block), Block(else_block) if else_block else None
        )

    def parse_while_loop(self) -> WhileLoop:
        self.consume("while")
        condition = self.parse_expression()
        self.consume("do")

        body = []
        while self.current_token() and self.current_token()[0] != "end":
            stmt = self.parse_statement()
            if stmt:
                body.append(stmt)

        self.consume("end")
        return WhileLoop(condition, Block(body))

    def parse_for_loop(self) -> ForLoop:
        """This should be formatted as for ASSIGN to EXPR do BODY end"""
        self.consume("for")

        assignment: Assignment = self.parse_assignment_or_expression()

        self.consume("to")
        expression = self.parse_expression()
        self.consume("do")

        body = []
        while self.current_token() and self.current_token()[0] != "end":
            stmt = self.parse_statement()
            if stmt:
                body.append(stmt)

        self.consume("end")
        return ForLoop(assignment, expression, Block(body))

    def parse_return_statement(self) -> ReturnStatement:
        self.consume("return")
        value = self.parse_expression()
        return ReturnStatement(value)

    def parse_assignment_or_expression(self) -> ASTNode:
        expr = self.parse_expression()

        if self.current_token() and self.current_token()[0] == "<-":
            self.consume("<-")
            value = self.parse_expression()
            return Assignment(expr, value)

        return expr

    def parse_expression(self) -> ASTNode:
        return self.parse_or_expression()

    def parse_or_expression(self) -> ASTNode:
        left = self.parse_and_expression()

        while self.current_token() and self.current_token()[0] == "or":
            op_token = self.consume()
            right = self.parse_and_expression()
            left = BinaryOp(left, op_token[0], right)

        return left

    def parse_and_expression(self) -> ASTNode:
        left = self.parse_comparison()

        while self.current_token() and self.current_token()[0] == "and":
            op_token = self.consume()
            right = self.parse_comparison()
            left = BinaryOp(left, op_token[0], right)

        return left

    def parse_comparison(self) -> ASTNode:
        left = self.parse_additive()

        while self.current_token() and self.current_token()[0] in [
            "<",
            ">",
            "<=",
            ">=",
            "!=",
            "=",
        ]:
            op_token = self.consume()
            right = self.parse_additive()
            left = BinaryOp(left, op_token[0], right)

        return left

    def parse_additive(self) -> ASTNode:
        left = self.parse_multiplicative()

        while self.current_token() and self.current_token()[0] in ["+", "-"]:
            op_token = self.consume()
            right = self.parse_multiplicative()
            left = BinaryOp(left, op_token[0], right)

        return left

    def parse_multiplicative(self) -> ASTNode:
        left = self.parse_unary()

        while self.current_token() and self.current_token()[0] in ["*", "/"]:
            op_token = self.consume()
            right = self.parse_unary()
            left = BinaryOp(left, op_token[0], right)

        return left

    def parse_unary(self) -> ASTNode:
        if self.current_token() and self.current_token()[0] == "not":
            op_token = self.consume()
            operand = self.parse_unary()
            return UnaryOp(op_token[0], operand)

        return self.parse_postfix()

    def parse_postfix(self) -> ASTNode:
        expr = self.parse_primary()

        while self.current_token():
            token = self.current_token()
            if token[0] == "[":
                self.consume("[")
                index = self.parse_expression()
                self.consume("]")
                expr = ArrayAccess(expr, index)
            else:
                break

        return expr

    def parse_primary(self) -> ASTNode:
        token = self.current_token()
        if token is None:
            raise SyntaxError("Unexpected end of input")

        value, token_type = token

        if token_type == Token.LITERAL:
            self.advance()
            return Literal(value)

        elif token_type == Token.IDENTIFIER:
            self.advance()
            return Identifier(value)

        elif value == "(":
            self.consume("(")
            expr = self.parse_expression()
            self.consume(")")
            return expr

        elif value == "[":
            return self.parse_array_literal()

        else:
            raise SyntaxError(f"Unexpected token: {value}")

    def parse_array_literal(self) -> ArrayLiteral:
        self.consume("[")
        elements = []

        if self.current_token() and self.current_token()[0] != "]":
            elements.append(self.parse_expression())
            while self.current_token() and self.current_token()[0] == ",":
                self.consume(",")
                if self.current_token()[0] != "]":
                    elements.append(self.parse_expression())

        self.consume("]")
        return ArrayLiteral(elements)
