from tokenizer import tokenize, validate_syntax

code_line = input("Enter your code: ")

tokens = tokenize(code_line)
print(validate_syntax(tokens))
