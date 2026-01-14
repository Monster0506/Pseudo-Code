import re
from enum import StrEnum, auto

code_line = input("Enter your code: ")


class Token(StrEnum):
    IDENTIFIER = auto()
    LITERAL = auto()
    OPERATOR = auto()
    KEYWORD = auto()
    PUNCTUATION = auto()


keywords = {
    "if",
    "else",
    "while",
    "for",
    "return",
    "Algorithm",
}

operators = {
    "+",
    "-",
    "*",
    "/",
    "=",
    "<-",
    "!=",
    "<",
    ">",
    "<=",
    ">=",
    "and",
    "or",
    "not",
}

punctuation = {
    "(",
    ")",
    "[",
    "]",
    "end",
    "do",
    "then",
}


def tokenize(line: str) -> list[tuple[str, Token]]:
    tokens = re.findall(r"\w+|<-|[+\-*/=<>!]+|[(){}\[\]]|and|or|not", line)

    result: list[tuple[str, Token]] = []
    for token in tokens:
        if token in keywords:
            result.append((token, Token.KEYWORD))
        elif token in operators:
            result.append((token, Token.OPERATOR))
        elif token in punctuation:
            result.append((token, Token.PUNCTUATION))
        elif re.match(r"^\d+$", token):
            result.append((token, Token.LITERAL))
        elif re.match(r'^".*"$|^\'.*\'$', token):
            result.append((token, Token.LITERAL))
        elif token == ",":
            continue
        else:
            result.append((token, Token.IDENTIFIER))
    return result


def validate_syntax(tokens):
    matching_pairs = {
        "(": ")",
        "do": "end",
        "[": "]",
        "then": "end",
    }

    stack = []
    for i, (value, token_type) in enumerate(tokens):
        if value in matching_pairs:
            stack.append(value)
        elif value in matching_pairs.values():
            if not stack:
                return False, f"Unexpected '{value}' without matching opening"

            last_open = stack[-1]
            if matching_pairs[last_open] == value:
                stack.pop()
            else:
                return False, f"'{value}' doesn't match '{last_open}'"

    if stack:
        return False, f"Unclosed delimiters: {stack}"

    return True, "Valid syntax"


tokens = tokenize(code_line)
print(tokens)
print(validate_syntax(tokens))
