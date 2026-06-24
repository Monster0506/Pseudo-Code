package main

import (
	"fmt"
	"math"
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
	InputSizes  map[string]int64
}

type symExpr map[string]int

type complexity struct {
	poly   symExpr
	logFac string
}

var cplxO1 = complexity{poly: symExpr{}}

func cplxFromParam(p string) complexity   { return complexity{poly: symExpr{p: 1}} }
func cplxLog(p string) complexity         { return complexity{poly: symExpr{}, logFac: p} }

func cplxDegree(c complexity) float64 {
	d := 0
	for _, e := range c.poly {
		d += e
	}
	if c.logFac != "" {
		return float64(d) + 0.5
	}
	return float64(d)
}

func mulCplx(a, b complexity) complexity {
	poly := make(symExpr, len(a.poly)+len(b.poly))
	for k, v := range a.poly {
		poly[k] += v
	}
	for k, v := range b.poly {
		poly[k] += v
	}
	logFac := a.logFac
	if logFac == "" {
		logFac = b.logFac
	}
	return complexity{poly: poly, logFac: logFac}
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

func fmtCplx(c complexity) string {
	polyStr := fmtSym(c.poly)
	if c.logFac == "" {
		return polyStr
	}
	logStr := "log " + c.logFac
	if polyStr == "1" {
		return logStr
	}
	return polyStr + " " + logStr
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

func forBound(n *ForNode, params map[string]bool) (complexity, string) {
	assign, ok := n.Init.(*AssignNode)
	if !ok {
		return cplxO1, "?"
	}
	if n.Direction == "downto" {
		if p := paramInExpr(assign.Value, params); p != "" {
			return cplxFromParam(p), p
		}
		return cplxO1, exprStr(assign.Value)
	}
	if isLitZero(assign.Value) {
		if bin, ok2 := n.End.(*BinaryOpNode); ok2 && bin.Op == "-" {
			if p := paramInExpr(bin.Left, params); p != "" {
				return cplxFromParam(p), p
			}
		}
		if p := paramInExpr(n.End, params); p != "" {
			return cplxFromParam(p), p
		}
	}
	if p := paramInExpr(n.End, params); p != "" {
		return cplxFromParam(p), exprStr(n.End)
	}
	return cplxO1, exprStr(n.End)
}

func whileBound(n *WhileNode, params map[string]bool) (complexity, string) {
	if hasLogarithmicUpdate(n.Cond, n.Body, params) {
		p := findLogParam(n.Cond, params)
		return cplxLog(p), "log " + p
	}
	if p := paramInExpr(n.Cond, params); p != "" {
		return cplxFromParam(p), p
	}
	return fallbackBound(params)
}

func repeatBound(n *RepeatNode, params map[string]bool) (complexity, string) {
	if hasLogarithmicUpdate(n.Cond, n.Body, params) {
		p := findLogParam(n.Cond, params)
		return cplxLog(p), "log " + p
	}
	if p := paramInExpr(n.Cond, params); p != "" {
		return cplxFromParam(p), p
	}
	return fallbackBound(params)
}

func fallbackBound(params map[string]bool) (complexity, string) {
	for _, p := range []string{"n", "m", "k", "N", "M"} {
		if params[p] {
			return cplxFromParam(p), p
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
		return cplxFromParam(lower[0]), lower[0]
	}
	if len(upper) > 0 {
		return cplxFromParam(upper[0]), upper[0]
	}
	return cplxO1, "n"
}

func hasLogarithmicUpdate(cond Node, body *BlockNode, params map[string]bool) bool {
	condVars := extractCondVarNames(cond)
	return hasLogUpdateInBlock(body, condVars)
}

func hasLogUpdateInBlock(b *BlockNode, condVars map[string]bool) bool {
	for _, s := range b.Stmts {
		if hasLogUpdateInStmt(s, condVars) {
			return true
		}
	}
	return false
}

func hasLogUpdateInStmt(n Node, condVars map[string]bool) bool {
	switch node := n.(type) {
	case *AssignNode:
		if _, ok := node.Target.(*ArrayAccessNode); ok {
			return false
		}
		varName := ""
		if id, ok := node.Target.(*IdentNode); ok {
			varName = id.Name
		}
		if condVars[varName] {
			return exprContainsOp(node.Value, "*") ||
				exprContainsOp(node.Value, "/") ||
				(exprContainsOp(node.Value, "mod") && exprUsesVars(node.Value, condVars))
		}
		if exprUsesVars(node.Value, condVars) && exprContainsOp(node.Value, "/") {
			return true
		}
	case *IfNode:
		if hasLogUpdateInBlock(node.Then, condVars) {
			return true
		}
		if node.Else != nil && hasLogUpdateInBlock(node.Else, condVars) {
			return true
		}
	}
	return false
}

func exprContainsOp(n Node, op string) bool {
	switch node := n.(type) {
	case *BinaryOpNode:
		if node.Op == op {
			return true
		}
		return exprContainsOp(node.Left, op) || exprContainsOp(node.Right, op)
	case *FuncCallNode:
		for _, a := range node.Args {
			if exprContainsOp(a, op) {
				return true
			}
		}
	case *UnaryOpNode:
		return exprContainsOp(node.Operand, op)
	}
	return false
}

func exprUsesVars(n Node, vars map[string]bool) bool {
	switch node := n.(type) {
	case *IdentNode:
		return vars[node.Name]
	case *BinaryOpNode:
		return exprUsesVars(node.Left, vars) || exprUsesVars(node.Right, vars)
	case *UnaryOpNode:
		return exprUsesVars(node.Operand, vars)
	case *FuncCallNode:
		for _, a := range node.Args {
			if exprUsesVars(a, vars) {
				return true
			}
		}
	}
	return false
}

func extractCondVarNames(cond Node) map[string]bool {
	vars := map[string]bool{}
	collectCondVarNames(cond, vars)
	return vars
}

func collectCondVarNames(n Node, vars map[string]bool) {
	switch node := n.(type) {
	case *IdentNode:
		vars[node.Name] = true
	case *BinaryOpNode:
		collectCondVarNames(node.Left, vars)
		collectCondVarNames(node.Right, vars)
	case *UnaryOpNode:
		collectCondVarNames(node.Operand, vars)
	case *ArrayAccessNode:
		collectCondVarNames(node.Array, vars)
		collectCondVarNames(node.Index, vars)
	}
}

func findLogParam(cond Node, params map[string]bool) string {
	if p := paramInExpr(cond, params); p != "" {
		return p
	}
	_, str := fallbackBound(params)
	return str
}

func detectRecursion(algo *AlgoNode, params map[string]bool) (complexity, string, bool) {
	calls := collectFuncCalls(algo.Body, algo.Name)
	if len(calls) == 0 {
		return cplxO1, "", false
	}
	for _, call := range calls {
		for i, arg := range call.Args {
			if i >= len(algo.Params) {
				break
			}
			p := algo.Params[i]
			if !params[p] {
				continue
			}
			if isParamMinusConst(arg, p) {
				return cplxFromParam(p), p + " - 1", true
			}
			if isParamDivConst(arg, p) {
				return cplxLog(p), p + " / k", true
			}
		}
	}
	c, str := fallbackBound(params)
	return c, str + " (recursive)", true
}

func collectFuncCalls(n Node, name string) []*FuncCallNode {
	var out []*FuncCallNode
	walkForCalls(n, name, &out)
	return out
}

func walkForCalls(n Node, name string, out *[]*FuncCallNode) {
	switch node := n.(type) {
	case *BlockNode:
		for _, s := range node.Stmts {
			walkForCalls(s, name, out)
		}
	case *FuncCallNode:
		if node.Name == name {
			*out = append(*out, node)
		}
		for _, a := range node.Args {
			walkForCalls(a, name, out)
		}
	case *IfNode:
		walkForCalls(node.Cond, name, out)
		walkForCalls(node.Then, name, out)
		if node.Else != nil {
			walkForCalls(node.Else, name, out)
		}
	case *WhileNode:
		walkForCalls(node.Cond, name, out)
		walkForCalls(node.Body, name, out)
	case *ForNode:
		walkForCalls(node.Body, name, out)
	case *RepeatNode:
		walkForCalls(node.Body, name, out)
	case *AssignNode:
		walkForCalls(node.Value, name, out)
	case *ReturnNode:
		if node.Value != nil {
			walkForCalls(node.Value, name, out)
		}
	case *BinaryOpNode:
		walkForCalls(node.Left, name, out)
		walkForCalls(node.Right, name, out)
	case *UnaryOpNode:
		walkForCalls(node.Operand, name, out)
	}
}

func isParamMinusConst(n Node, param string) bool {
	bin, ok := n.(*BinaryOpNode)
	if !ok || bin.Op != "-" {
		return false
	}
	id, ok := bin.Left.(*IdentNode)
	if !ok || id.Name != param {
		return false
	}
	_, ok = bin.Right.(*LiteralNode)
	return ok
}

func isParamDivConst(n Node, param string) bool {
	bin, ok := n.(*BinaryOpNode)
	if !ok || bin.Op != "/" {
		return false
	}
	id, ok := bin.Left.(*IdentNode)
	if !ok || id.Name != param {
		return false
	}
	_, ok = bin.Right.(*LiteralNode)
	return ok
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
	cplx   complexity
	bounds []string
}

func walkSym(b *BlockNode, params map[string]bool, ctx complexity, ctxBounds []string, depth int, lines *[]DerivLine) walkResult {
	best := walkResult{ctx, ctxBounds}

	for _, stmt := range b.Stmts {
		switch n := stmt.(type) {
		case *ForNode:
			c, bStr := forBound(n, params)
			*lines = append(*lines, DerivLine{depth, forLabel(n), bStr + " iterations"})
			inner := mulCplx(ctx, c)
			innerBounds := append(append([]string{}, ctxBounds...), bStr)
			if !hasLoops(n.Body) {
				*lines = append(*lines, DerivLine{depth + 1, bodyLabel(n.Body), "O(1)"})
			}
			r := walkSym(n.Body, params, inner, innerBounds, depth+1, lines)
			if cplxDegree(r.cplx) > cplxDegree(best.cplx) {
				best = r
			}

		case *WhileNode:
			c, bStr := whileBound(n, params)
			note := bStr + " iterations"
			if !strings.HasPrefix(bStr, "log") {
				note += " (worst case)"
			}
			*lines = append(*lines, DerivLine{depth, "while " + exprStr(n.Cond), note})
			inner := mulCplx(ctx, c)
			innerBounds := append(append([]string{}, ctxBounds...), bStr)
			if !hasLoops(n.Body) {
				*lines = append(*lines, DerivLine{depth + 1, bodyLabel(n.Body), "O(1)"})
			}
			r := walkSym(n.Body, params, inner, innerBounds, depth+1, lines)
			if cplxDegree(r.cplx) > cplxDegree(best.cplx) {
				best = r
			}

		case *RepeatNode:
			c, bStr := repeatBound(n, params)
			note := bStr + " iterations"
			if !strings.HasPrefix(bStr, "log") {
				note += " (worst case)"
			}
			*lines = append(*lines, DerivLine{depth, "repeat/until " + exprStr(n.Cond), note})
			inner := mulCplx(ctx, c)
			innerBounds := append(append([]string{}, ctxBounds...), bStr)
			if !hasLoops(n.Body) {
				*lines = append(*lines, DerivLine{depth + 1, bodyLabel(n.Body), "O(1)"})
			}
			r := walkSym(n.Body, params, inner, innerBounds, depth+1, lines)
			if cplxDegree(r.cplx) > cplxDegree(best.cplx) {
				best = r
			}

		case *IfNode:
			r := walkSym(n.Then, params, ctx, ctxBounds, depth, lines)
			if cplxDegree(r.cplx) > cplxDegree(best.cplx) {
				best = r
			}
			if n.Else != nil {
				r2 := walkSym(n.Else, params, ctx, ctxBounds, depth, lines)
				if cplxDegree(r2.cplx) > cplxDegree(best.cplx) {
					best = r2
				}
			}
		}
	}
	return best
}

type weightedOp struct {
	kind string
	cplx complexity
}

func collectOps(n Node, params map[string]bool, ctx complexity, out *[]weightedOp) {
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
		c, _ := forBound(node, params)
		inner := mulCplx(ctx, c)
		for _, s := range node.Body.Stmts {
			collectOps(s, params, inner, out)
		}
	case *WhileNode:
		c, _ := whileBound(node, params)
		inner := mulCplx(ctx, c)
		for _, s := range node.Body.Stmts {
			collectOps(s, params, inner, out)
		}
	case *RepeatNode:
		c, _ := repeatBound(node, params)
		inner := mulCplx(ctx, c)
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
		collectOps(s, params, cplxO1, &all)
	}

	type grp struct {
		count int
		cplx  complexity
	}
	m := map[string]*grp{}
	for _, op := range all {
		if g, ok := m[op.kind]; ok {
			deg := cplxDegree(op.cplx)
			if deg > cplxDegree(g.cplx) {
				g.count = 1
				g.cplx = op.cplx
			} else if deg == cplxDegree(g.cplx) {
				g.count++
			}
		} else {
			m[op.kind] = &grp{1, op.cplx}
		}
	}

	fmtCount := func(count int, c complexity) string {
		s := fmtCplx(c)
		if count == 1 {
			return s
		}
		return fmt.Sprintf("%d%s", count, s)
	}

	var rows []OpRow
	if g, ok := m["cmp"]; ok {
		rows = append(rows, OpRow{"comparisons", fmtCount(g.count, g.cplx), ""})
	}
	if g, ok := m["read"]; ok {
		rows = append(rows, OpRow{"array reads", fmtCount(g.count, g.cplx), ""})
	}
	if g, ok := m["write"]; ok {
		rows = append(rows, OpRow{"array writes", "<= " + fmtCount(g.count, g.cplx), "worst case"})
	}

	domCplx := cplxO1
	for _, op := range all {
		if cplxDegree(op.cplx) > cplxDegree(domCplx) {
			domCplx = op.cplx
		}
	}
	if cplxDegree(domCplx) > 0 {
		rows = append(rows, OpRow{"loop iters", fmtCplx(domCplx), ""})
	}
	return rows
}

type StaticAnalyzer struct{}

func (a *StaticAnalyzer) Analyze(algo *AlgoNode) ComplexityReport {
	params := make(map[string]bool, len(algo.Params))
	for _, p := range algo.Params {
		params[p] = true
	}

	if cplx, reduceStr, isRecursive := detectRecursion(algo, params); isRecursive {
		return ComplexityReport{
			AlgoName: algo.Name,
			Params:   algo.Params,
			BigO:     "O(" + fmtCplx(cplx) + ")",
			Derivation: []DerivLine{
				{0, "recursive call with " + reduceStr, "T(n) = T(n-1) + O(1)"},
			},
		}
	}

	var lines []DerivLine
	best := walkSym(algo.Body, params, cplxO1, nil, 0, &lines)

	footer := ""
	if len(best.bounds) > 0 {
		footer = "= " + strings.Join(best.bounds, " × ") + " × O(1)"
	}

	return ComplexityReport{
		AlgoName:   algo.Name,
		Params:     algo.Params,
		BigO:       "O(" + fmtCplx(best.cplx) + ")",
		Footer:     footer,
		Derivation: lines,
		OpTable:    buildOpTable(algo.Body, params),
	}
}

type RuntimeAnalyzer struct {
	vm *VM
}

func NewRuntimeAnalyzer(vm *VM) *RuntimeAnalyzer {
	vm.Counters = &VMCounters{InputSizes: make(map[string]int64)}
	return &RuntimeAnalyzer{vm: vm}
}

func (a *RuntimeAnalyzer) Analyze(algo *AlgoNode) ComplexityReport {
	c := a.vm.Counters
	if c == nil {
		return ComplexityReport{AlgoName: algo.Name, BigO: "O(?)"}
	}
	bigO := inferBigOFromCounts(c)
	return ComplexityReport{
		AlgoName: algo.Name,
		Params:   algo.Params,
		BigO:     bigO,
		OpTable: []OpRow{
			{"comparisons", fmt.Sprintf("%d", c.Comparisons), ""},
			{"array reads", fmt.Sprintf("%d", c.ArrayReads), ""},
			{"array writes", fmt.Sprintf("%d", c.ArrayWrites), ""},
			{"loop iters", fmt.Sprintf("%d", c.LoopIters), ""},
		},
	}
}

func inferBigOFromCounts(c *VMCounters) string {
	if c.LoopIters == 0 {
		return "O(1)"
	}
	fCount := float64(c.LoopIters)

	type candidate struct {
		name  string
		score float64
	}
	best := candidate{"?", math.MaxFloat64}

	for param, n := range c.InputSizes {
		if n <= 1 {
			continue
		}
		fN := float64(n)
		logN := math.Log2(fN)

		for _, trial := range []struct {
			suffix string
			val    float64
		}{
			{"1", 1},
			{"log " + param, logN},
			{param, fN},
			{param + " log " + param, fN * logN},
			{param + "²", fN * fN},
			{param + "³", fN * fN * fN},
		} {
			if trial.val <= 0 {
				continue
			}
			ratio := fCount / trial.val
			if ratio < 0.05 || ratio > 200 {
				continue
			}
			rounded := math.Round(ratio)
			if rounded < 1 {
				rounded = 1
			}
			score := math.Abs(ratio-rounded) / rounded
			if score < best.score {
				best.score = score
				best.name = trial.suffix
			}
		}
	}

	if best.name == "?" {
		return "O(?)"
	}
	return "O(" + best.name + ")"
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
