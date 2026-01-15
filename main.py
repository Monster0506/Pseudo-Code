from sys import argv

from parser import Parser
from tokenizer import tokenize, validate_syntax
from generator import Generator
from vm import VM


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

    vm = VM(instructions)
    print("=" * 60)

    print("Required inputs:", list(vm.inputs.keys()))
    for var in vm.inputs.keys():
        user_input = input(f"{var} = ")

        if user_input.startswith("["):
            vm.inputs[var] = eval(user_input)
        else:
            try:
                vm.inputs[var] = int(user_input)
            except ValueError:
                vm.inputs[var] = user_input
    print("=" * 60)
    result = vm.run(**vm.inputs)

    print("\n".join(lines))

    print("=" * 60)
    print(f"Return value: {result}")
    print("=" * 60)

except SyntaxError as e:
    print(f"Parse error: {e}")
