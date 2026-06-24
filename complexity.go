package main

import (
	"fmt"
	"sort"
	"strings"
)

type Analyzer interface {
	Analyze(algo *AlgoNode) ComplexityReport
}

type ComplexityReport struct {
	AlgoName   string
	Params     []string
	BigO       string
	Footer     string
	Derivation []DerivLine
	OpTable    []OpRow
}

type DerivLine struct {
	Indent int
	Label  string
	Note   string
}

type OpRow struct {
	Op    string
	Count string
	Note  string
}

type VMCounters struct {
	Comparisons int64
	ArrayReads  int64
	ArrayWrites int64
	LoopIters   int64
}

type symExpr map[string]int

func symDegree(s symExpr) int {
	d := 0
	for _, e := range s {
		d += e
	}
	return d
}

func mulSym(a, b symExpr) symExpr {
	r := make(symExpr, len(a)+len(b))
	for k, v := range a {
		r[k] += v
	}
	for k, v := range b {
		r[k] += v
	}
	return r
}

func fmtSym(s symExpr) string {
	if len(s) == 0 {
		return "1"
	}
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		switch s[k] {
		case 1:
		case 2:
			sb.WriteString("²")
		case 3:
			sb.WriteString("³")
		default:
			fmt.Fprintf(&sb, "^%d", s[k])
		}
	}
	return sb.String()
}

func exprStr(n Node) string {
	switch node := n.(type) {
	case *LiteralNode:
		return node.Value
	case *IdentNode:
		return node.Name
	case *BinaryOpNode:
		return exprStr(node.Left) + " " + node.Op + " " + exprStr(node.Right)
	case *UnaryOpNode:
		return node.Op + " " + exprStr(node.Operand)
	case *ArrayAccessNode:
		return exprStr(node.Array) + "[" + exprStr(node.Index) + "]"
	case *FuncCallNode:
		args := make([]string, len(node.Args))
		for i, a := range node.Args {
			args[i] = exprStr(a)
		}
		return node.Name + "(" + strings.Join(args, ", ") + ")"
	}
	return "?"
}

func paramInExpr(n Node, params map[string]bool) string {
	switch node := n.(type) {
	case *IdentNode:
		if params[node.Name] {
			return node.Name
		}
	case *BinaryOpNode:
		if p := paramInExpr(node.Left, params); p != "" {
			return p
		}
		return paramInExpr(node.Right, params)
	case *UnaryOpNode:
		return paramInExpr(node.Operand, params)
	}
	return ""
}

func isLitZero(n Node) bool {
	l, ok := n.(*LiteralNode)
	return ok && l.Value == "0"
}

func forBound(n *ForNode, params map[string]bool) (symExpr, string) {
	assign, ok := n.Init.(*AssignNode)
	if !ok {
		return symExpr{}, "?"
	}
	if n.Direction == "downto" {
		if p := paramInExpr(assign.Value, params); p != "" {
			return symExpr{p: 1}, p
		}
		return symExpr{}, exprStr(assign.Value)
	}
	if isLitZero(assign.Value) {
		if bin, ok2 := n.End.(*BinaryOpNode); ok2 && bin.Op == "-" {
			if p := paramInExpr(bin.Left, params); p != "" {
				return symExpr{p: 1}, p
			}
		}
		if p := paramInExpr(n.End, params); p != "" {
			return symExpr{p: 1}, p
		}
	}
	if p := paramInExpr(n.End, params); p != "" {
		return symExpr{p: 1}, exprStr(n.End)
	}
	return symExpr{}, exprStr(n.End)
}

func whileBound(n *WhileNode, params map[string]bool) (symExpr, string) {
	if p := paramInExpr(n.Cond, params); p != "" {
		return symExpr{p: 1}, p
	}
	return fallbackBound(params)
}

func repeatBound(n *RepeatNode, params map[string]bool) (symExpr, string) {
	if p := paramInExpr(n.Cond, params); p != "" {
		return symExpr{p: 1}, p
	}
	return fallbackBound(params)
}

func fallbackBound(params map[string]bool) (symExpr, string) {
	preferred := []string{"n", "m", "k", "N", "M"}
	for _, p := range preferred {
		if params[p] {
			return symExpr{p: 1}, p
		}
	}
	var lower, upper []string
	for k := range params {
		if len(k) == 1 && k[0] >= 'a' && k[0] <= 'z' {
			lower = append(lower, k)
		} else {
			upper = append(upper, k)
		}
	}
	sort.Strings(lower)
	sort.Strings(upper)
	if len(lower) > 0 {
		return symExpr{lower[0]: 1}, lower[0]
	}
	if len(upper) > 0 {
		return symExpr{upper[0]: 1}, upper[0]
	}
	return symExpr{}, "n"
}

func forLabel(n *ForNode) string {
	assign := n.Init.(*AssignNode)
	v := exprStr(assign.Target)
	start := exprStr(assign.Value)
	end := exprStr(n.End)
	if n.Direction == "downto" {
		return "for " + v + " in [" + start + ".." + end + "] downto"
	}
	return "for " + v + " in [" + start + ".." + end + "]"
}

func hasLoops(b *BlockNode) bool {
	for _, s := range b.Stmts {
		switch n := s.(type) {
		case *ForNode, *WhileNode, *RepeatNode:
			return true
		case *IfNode:
			if hasLoops(n.Then) || (n.Else != nil && hasLoops(n.Else)) {
				return true
			}
		}
	}
	return false
}

func bodyLabel(b *BlockNode) string {
	seen := map[string]bool{}
	for _, s := range b.Stmts {
		switch s.(type) {
		case *IfNode:
			seen["if"] = true
		case *AssignNode:
			seen["assign"] = true
		case *PrintNode:
			seen["print"] = true
		case *ReturnNode:
			seen["return"] = true
		case *FuncCallNode:
			seen["call"] = true
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return "body"
	}
	return strings.Join(keys, "/")
}

type walkResult struct {
	sym    symExpr
	bounds []string
}

func walkSym(b *BlockNode, params map[string]bool, ctx symExpr, ctxBounds []string, depth int, lines *[]DerivLine) walkResult {
	best := walkResult{ctx, ctxBounds}

	for _, stmt := range b.Stmts {
		switch n := stmt.(type) {
		case *ForNode:
			sym, bStr := forBound(n, params)
			*lines = append(*lines, DerivLine{depth, forLabel(n), bStr + " iterations"})
			inner := mulSym(ctx, sym)
			innerBounds := append(append([]string{}, ctxBounds...), bStr)
			if !hasLoops(n.Body) {
				*lines = append(*lines, DerivLine{depth + 1, bodyLabel(n.Body), "O(1)"})
			}
			r := walkSym(n.Body, params, inner, innerBounds, depth+1, lines)
			if symDegree(r.sym) > symDegree(best.sym) {
				best = r
			}

		case *WhileNode:
			sym, bStr := whileBound(n, params)
			*lines = append(*lines, DerivLine{depth, "while " + exprStr(n.Cond), bStr + " iterations (worst case)"})
			inner := mulSym(ctx, sym)
			innerBounds := append(append([]string{}, ctxBounds...), bStr)
			if !hasLoops(n.Body) {
				*lines = append(*lines, DerivLine{depth + 1, bodyLabel(n.Body), "O(1)"})
			}
			r := walkSym(n.Body, params, inner, innerBounds, depth+1, lines)
			if symDegree(r.sym) > symDegree(best.sym) {
				best = r
			}

		case *RepeatNode:
			sym, bStr := repeatBound(n, params)
			*lines = append(*lines, DerivLine{depth, "repeat/until " + exprStr(n.Cond), bStr + " iterations (worst case)"})
			inner := mulSym(ctx, sym)
			innerBounds := append(append([]string{}, ctxBounds...), bStr)
			if !hasLoops(n.Body) {
				*lines = append(*lines, DerivLine{depth + 1, bodyLabel(n.Body), "O(1)"})
			}
			r := walkSym(n.Body, params, inner, innerBounds, depth+1, lines)
			if symDegree(r.sym) > symDegree(best.sym) {
				best = r
			}

		case *IfNode:
			r := walkSym(n.Then, params, ctx, ctxBounds, depth, lines)
			if symDegree(r.sym) > symDegree(best.sym) {
				best = r
			}
			if n.Else != nil {
				r2 := walkSym(n.Else, params, ctx, ctxBounds, depth, lines)
				if symDegree(r2.sym) > symDegree(best.sym) {
					best = r2
				}
			}
		}
	}
	return best
}

type weightedOp struct {
	kind string
	sym  symExpr
}

func collectOps(n Node, params map[string]bool, ctx symExpr, out *[]weightedOp) {
	switch node := n.(type) {
	case *BinaryOpNode:
		switch node.Op {
		case "<", ">", "<=", ">=", "=", "!=":
			*out = append(*out, weightedOp{"cmp", ctx})
		}
		collectOps(node.Left, params, ctx, out)
		collectOps(node.Right, params, ctx, out)
	case *UnaryOpNode:
		collectOps(node.Operand, params, ctx, out)
	case *ArrayAccessNode:
		*out = append(*out, weightedOp{"read", ctx})
		collectOps(node.Index, params, ctx, out)
	case *AssignNode:
		if _, ok := node.Target.(*ArrayAccessNode); ok {
			*out = append(*out, weightedOp{"write", ctx})
		}
		collectOps(node.Value, params, ctx, out)
	case *IfNode:
		collectOps(node.Cond, params, ctx, out)
		for _, s := range node.Then.Stmts {
			collectOps(s, params, ctx, out)
		}
		if node.Else != nil {
			for _, s := range node.Else.Stmts {
				collectOps(s, params, ctx, out)
			}
		}
	case *ForNode:
		sym, _ := forBound(node, params)
		inner := mulSym(ctx, sym)
		for _, s := range node.Body.Stmts {
			collectOps(s, params, inner, out)
		}
	case *WhileNode:
		sym, _ := whileBound(node, params)
		inner := mulSym(ctx, sym)
		for _, s := range node.Body.Stmts {
			collectOps(s, params, inner, out)
		}
	case *RepeatNode:
		sym, _ := repeatBound(node, params)
		inner := mulSym(ctx, sym)
		for _, s := range node.Body.Stmts {
			collectOps(s, params, inner, out)
		}
	case *PrintNode:
		collectOps(node.Expr, params, ctx, out)
	case *ReturnNode:
		if node.Value != nil {
			collectOps(node.Value, params, ctx, out)
		}
	case *FuncCallNode:
		for _, a := range node.Args {
			collectOps(a, params, ctx, out)
		}
	case *BlockNode:
		for _, s := range node.Stmts {
			collectOps(s, params, ctx, out)
		}
	}
}

func buildOpTable(b *BlockNode, params map[string]bool) []OpRow {
	var all []weightedOp
	for _, s := range b.Stmts {
		collectOps(s, params, symExpr{}, &all)
	}

	type grp struct {
		count int
		sym   symExpr
	}
	m := map[string]*grp{}
	for _, op := range all {
		if g, ok := m[op.kind]; ok {
			if symDegree(op.sym) > symDegree(g.sym) {
				g.count = 1
				g.sym = op.sym
			} else if symDegree(op.sym) == symDegree(g.sym) {
				g.count++
			}
		} else {
			m[op.kind] = &grp{1, op.sym}
		}
	}

	fmtCount := func(count int, sym symExpr) string {
		s := fmtSym(sym)
		if count == 1 {
			return s
		}
		return fmt.Sprintf("%d%s", count, s)
	}

	var rows []OpRow
	if g, ok := m["cmp"]; ok {
		rows = append(rows, OpRow{"comparisons", fmtCount(g.count, g.sym), ""})
	}
	if g, ok := m["read"]; ok {
		rows = append(rows, OpRow{"array reads", fmtCount(g.count, g.sym), ""})
	}
	if g, ok := m["write"]; ok {
		rows = append(rows, OpRow{"array writes", "<= " + fmtCount(g.count, g.sym), "worst case"})
	}

	domSym := symExpr{}
	for _, op := range all {
		if symDegree(op.sym) > symDegree(domSym) {
			domSym = op.sym
		}
	}
	if symDegree(domSym) > 0 {
		rows = append(rows, OpRow{"loop iters", fmtSym(domSym), ""})
	}
	return rows
}

// StaticAnalyzer implements Analyzer via AST walking.
type StaticAnalyzer struct{}

func (a *StaticAnalyzer) Analyze(algo *AlgoNode) ComplexityReport {
	params := make(map[string]bool, len(algo.Params))
	for _, p := range algo.Params {
		params[p] = true
	}

	var lines []DerivLine
	best := walkSym(algo.Body, params, symExpr{}, nil, 0, &lines)

	footer := ""
	if len(best.bounds) > 0 {
		footer = "= " + strings.Join(best.bounds, " × ") + " × O(1)"
	}

	return ComplexityReport{
		AlgoName:   algo.Name,
		Params:     algo.Params,
		BigO:       "O(" + fmtSym(best.sym) + ")",
		Footer:     footer,
		Derivation: lines,
		OpTable:    buildOpTable(algo.Body, params),
	}
}

type RuntimeAnalyzer struct {
	vm *VM
}

func NewRuntimeAnalyzer(vm *VM) *RuntimeAnalyzer {
	vm.Counters = &VMCounters{}
	return &RuntimeAnalyzer{vm: vm}
}

func (a *RuntimeAnalyzer) Analyze(algo *AlgoNode) ComplexityReport {
	c := a.vm.Counters
	if c == nil {
		return ComplexityReport{AlgoName: algo.Name, BigO: "O(?)"}
	}
	return ComplexityReport{
		AlgoName: algo.Name,
		Params:   algo.Params,
		BigO:     "O(?)",
		OpTable: []OpRow{
			{"comparisons", fmt.Sprintf("%d", c.Comparisons), ""},
			{"array reads", fmt.Sprintf("%d", c.ArrayReads), ""},
			{"array writes", fmt.Sprintf("%d", c.ArrayWrites), ""},
			{"loop iters", fmt.Sprintf("%d", c.LoopIters), ""},
		},
	}
}

func PrintComplexityReport(r ComplexityReport) {
	fmt.Println(sep)
	fmt.Printf("Complexity: %s\n", r.BigO)

	if len(r.Derivation) > 0 {
		fmt.Println()
		fmt.Printf("%s(%s):\n", r.AlgoName, strings.Join(r.Params, ", "))
		const labelW = 38
		for _, line := range r.Derivation {
			indent := strings.Repeat("  ", line.Indent+1)
			label := indent + line.Label
			pad := labelW - len(label)
			if pad < 2 {
				pad = 2
			}
			fmt.Printf("%s%s%s\n", label, strings.Repeat(" ", pad), line.Note)
		}
		if r.Footer != "" {
			fmt.Printf("  %s\n", r.Footer)
		}
	}

	if len(r.OpTable) > 0 {
		fmt.Println()
		const opW = 14
		for _, row := range r.OpTable {
			op := row.Op
			if len(op) < opW {
				op += strings.Repeat(" ", opW-len(op))
			}
			note := ""
			if row.Note != "" {
				note = "  (" + row.Note + ")"
			}
			fmt.Printf("  %s%s%s\n", op, row.Count, note)
		}
	}
}
