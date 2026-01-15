from sys import argv

from parser import Parser
from tokenizer import tokenize, validate_syntax


def get_code(filename):
    with open(filename, "r") as f:
        contents = f.read()
    return contents


tokens = tokenize(get_code(argv[1]))
if not validate_syntax(tokens)[0]:
    raise SyntaxError(validate_syntax(tokens)[1]) 

try:
    parser = Parser(tokens)
    ast = parser.parse()
    print("AST:", ast)
    print("Parsed statements:")
    for stmt in ast.statements:
        print(f"  {stmt}")
except SyntaxError as e:
    print(f"Parse error: {e}")
