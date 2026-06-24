package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const sep = "============================================================"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: pseudo <file.psu>")
		os.Exit(1)
	}

	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading file:", err)
		os.Exit(1)
	}

	tokens := Tokenize(string(src))

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(1)
	}

	var allAlgos []*AlgoNode
	for _, s := range ast.Stmts {
		if a, ok := s.(*AlgoNode); ok {
			allAlgos = append(allAlgos, a)
		}
	}

	gen := NewGenerator()
	instrs, funcTable, entryParams, err := gen.Generate(ast)
	if err != nil {
		fmt.Fprintln(os.Stderr, "codegen error:", err)
		os.Exit(1)
	}

	sa := &StaticAnalyzer{}
	for _, algo := range allAlgos {
		PrintComplexityReport(sa.Analyze(algo))
	}

	fmt.Println(sep)
	fmt.Fprintf(os.Stderr, "Required inputs: %v\n", entryParams)
	initial := make(map[string]Value, len(entryParams))
	scanner := bufio.NewScanner(os.Stdin)
	for _, param := range entryParams {
		fmt.Fprintf(os.Stderr, "%s = ", param)
		if scanner.Scan() {
			initial[param] = parseInput(scanner.Text())
		}
	}
	fmt.Println(sep)

	vm := NewVM(instrs, funcTable)
	if vm.Counters != nil {
		for k, v := range initial {
			switch v.Kind {
			case TypeInt:
				vm.Counters.InputSizes[k] = int64(v.IntVal)
			case TypeArray:
				vm.Counters.InputSizes[k] = int64(len(v.ArrVal))
			}
		}
	}
	result, err := vm.Run(initial)
	if err != nil {
		fmt.Fprintln(os.Stderr, "runtime error:", err)
		os.Exit(1)
	}

	width := len(strconv.Itoa(len(instrs))) + 1
	for i, instr := range instrs {
		fmt.Printf("%*d: %s\n", width, i, instr.String())
	}

	fmt.Println(sep)
	fmt.Printf("Return value: %s\n", result.Format())
	fmt.Println(sep)
}

func parseInput(s string) Value {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := s[1 : len(s)-1]
		if inner == "" {
			return ArrayVal(nil)
		}
		parts := strings.Split(inner, ",")
		elems := make([]Value, len(parts))
		for i, p := range parts {
			p = strings.TrimSpace(p)
			if n, err := strconv.Atoi(p); err == nil {
				elems[i] = IntVal(n)
			} else {
				elems[i] = StrVal(p)
			}
		}
		return ArrayVal(elems)
	}
	if n, err := strconv.Atoi(s); err == nil {
		return IntVal(n)
	}
	return StrVal(s)
}
