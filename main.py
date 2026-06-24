from sys import argv

from parser import Parser
from tokenizer import tokenize, validate_syntax
from generator import Generator
from vm import VM


def get_code(filename):
    with open(filename, "r") as f:
        return f.read()


tokens = tokenize(get_code(argv[1]))
valid, msg = validate_syntax(tokens)
if not valid:
    raise SyntaxError(msg)

try:
    parser = Parser(tokens)
    ast = parser.parse()

    generator = Generator()
    instructions = generator.generate(ast)

    lines = [
        f"{i:{len(str(len(instructions)))+1}d}: {instr}"
        for i, instr in enumerate(instructions)
    ]

    vm = VM(instructions, generator.function_table)

    # Use the entry function's declared parameter names when available
    input_names = generator.entry_params or list(vm.inputs.keys())

    print("=" * 60)
    print("Required inputs:", input_names)
    initial_vars = {}
    for var in input_names:
        user_input = input(f"{var} = ")
        if user_input.startswith("["):
            initial_vars[var] = eval(user_input)
        else:
            try:
                initial_vars[var] = int(user_input)
            except ValueError:
                initial_vars[var] = user_input
    print("=" * 60)

    result = vm.run(**initial_vars)

    print("\n".join(lines))

    print("=" * 60)
    print(f"Return value: {result}")
    print("=" * 60)

except SyntaxError as e:
    print(f"Parse error: {e}")
