from sys import argv

from parser import Parser
from tokenizer import tokenize, validate_syntax
from generator import Generator


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

    generator = Generator()
    instructions = generator.generate(ast)
    lines = [
        f"{i:{len(str(len(instructions)))+1}d}: {instr}"
        for i, instr in enumerate(instructions)
    ]
    print("\n".join(lines))

except SyntaxError as e:
    print(f"Parse error: {e}")
