package main

import (
	"regexp"
	"strings"
)

type TokenType int

const (
	IDENTIFIER  TokenType = iota
	LITERAL
	OPERATOR
	KEYWORD
	PUNCTUATION
)

type Token struct {
	Value string
	Type  TokenType
}

var keywords = map[string]bool{
	"if": true, "else": true, "while": true, "for": true,
	"return": true, "Algorithm": true, "repeat": true,
	"print": true, "output": true,
}

var operators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true, "=": true,
	"<-": true, "!=": true, "<": true, ">": true, "<=": true, ">=": true,
	"and": true, "or": true, "not": true, "mod": true,
}

var punctuation = map[string]bool{
	"(": true, ")": true, "[": true, "]": true, ",": true,
	"end": true, "do": true, "then": true, "to": true, "until": true,
	"downto": true,
}

var tokenRe = regexp.MustCompile(`"[^"]*"|'[^']*'|-?\d+|\w+|<-|[+\-*/=<>!]+|[(),\[\]]`)

func Tokenize(text string) []Token {
	var result []Token
	for _, line := range strings.Split(text, "\n") {
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for _, raw := range tokenRe.FindAllString(line, -1) {
			result = append(result, classify(raw))
		}
	}
	return result
}

func classify(raw string) Token {
	if keywords[raw] {
		return Token{raw, KEYWORD}
	}
	if operators[raw] {
		return Token{raw, OPERATOR}
	}
	if punctuation[raw] {
		return Token{raw, PUNCTUATION}
	}
	if isIntLiteral(raw) {
		return Token{raw, LITERAL}
	}
	if isStringLiteral(raw) {
		return Token{raw, LITERAL}
	}
	return Token{raw, IDENTIFIER}
}

func isIntLiteral(s string) bool {
	if len(s) == 0 {
		return false
	}
	start := 0
	if s[0] == '-' {
		start = 1
	}
	if start == len(s) {
		return false
	}
	for _, c := range s[start:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isStringLiteral(s string) bool {
	return (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
		(strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`))
}
