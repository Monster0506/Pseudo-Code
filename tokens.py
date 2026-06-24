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
    "true",
    "false",
    "NIL",
    "null",
    "infinity",
    "print",
    "output",
    "repeat",
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
    "mod",
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
    "to",
    "downto",
    "by",
    "until",
}
