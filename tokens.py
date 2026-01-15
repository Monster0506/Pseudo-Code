from enum import StrEnum, auto


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
    ",",
    "end",
    "do",
    "then",
}
