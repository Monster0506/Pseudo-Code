from parser import Parser
from tokenizer import tokenize, validate_syntax

code_line = input("Enter your code: ")

tokens = tokenize(code_line)
print(validate_syntax(tokens))

try:
    parser = Parser(tokens)
    ast = parser.parse()
    print("AST:", ast)
    print("Parsed statements:")
    for stmt in ast.statements:
        print(f"  {stmt}")
except SyntaxError as e:
    print(f"Parse error: {e}")
