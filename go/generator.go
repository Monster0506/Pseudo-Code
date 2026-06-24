package main

import (
	"fmt"
	"strconv"
)

type FuncEntry struct {
	StartIdx   int
	ParamNames []string
}

type Generator struct {
	instrs      []Instruction
	tempN       int
	funcTable   map[string]FuncEntry
	entryParams []string
}

func NewGenerator() *Generator {
	return &Generator{funcTable: make(map[string]FuncEntry)}
}

func (g *Generator) newTemp() string {
	s := fmt.Sprintf("t%d", g.tempN)
	g.tempN++
	return s
}

func (g *Generator) emit(op OpCode, operands ...string) {
	g.instrs = append(g.instrs, Instruction{op, operands})
}

func (g *Generator) Generate(ast *BlockNode) ([]Instruction, map[string]FuncEntry, []string, error) {
	g.instrs = nil
	g.tempN = 0
	g.funcTable = make(map[string]FuncEntry)
	g.entryParams = nil

	var funcs []*AlgoNode
	var bare []Node
	for _, s := range ast.Stmts {
		if a, ok := s.(*AlgoNode); ok {
			funcs = append(funcs, a)
		} else {
			bare = append(bare, s)
		}
	}

	if len(funcs) > 0 {
		entry := funcs[0]
		g.entryParams = entry.Params
		g.funcTable[entry.Name] = FuncEntry{len(g.instrs), entry.Params}
		if err := g.visitBlock(entry.Body); err != nil {
			return nil, nil, nil, err
		}
		g.emit(RET)
	}

	for _, s := range bare {
		if _, err := g.visit(s); err != nil {
			return nil, nil, nil, err
		}
	}

	return g.instrs, g.funcTable, g.entryParams, nil
}

func (g *Generator) visitBlock(b *BlockNode) error {
	for _, s := range b.Stmts {
		if _, err := g.visit(s); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) visit(n Node) (string, error) {
	switch node := n.(type) {
	case *LiteralNode:
		return node.Value, nil
	case *IdentNode:
		return node.Name, nil
	case *ArrayLitNode:
		return g.visitArrayLit(node)
	case *ArrayAccessNode:
		return g.visitArrayAccess(node)
	case *BinaryOpNode:
		return g.visitBinaryOp(node)
	case *UnaryOpNode:
		return g.visitUnaryOp(node)
	case *AssignNode:
		return "", g.visitAssign(node)
	case *IfNode:
		return "", g.visitIf(node)
	case *WhileNode:
		return "", g.visitWhile(node)
	case *ForNode:
		return "", g.visitFor(node)
	case *RepeatNode:
		return "", g.visitRepeat(node)
	case *ReturnNode:
		return "", g.visitReturn(node)
	case *BlockNode:
		return "", g.visitBlock(node)
	case *AlgoNode:
		if err := g.visitBlock(node.Body); err != nil {
			return "", err
		}
		g.emit(RET)
		return "", nil
	}
	return "", fmt.Errorf("unknown node type %T", n)
}

func (g *Generator) visitArrayLit(node *ArrayLitNode) (string, error) {
	t := g.newTemp()
	elems := make([]string, len(node.Elems))
	for i, e := range node.Elems {
		v, err := g.visit(e)
		if err != nil {
			return "", err
		}
		elems[i] = v
	}
	ops := append([]string{t}, elems...)
	g.emit(ASN, ops...)
	return t, nil
}

func (g *Generator) visitArrayAccess(node *ArrayAccessNode) (string, error) {
	var arrName string
	if id, ok := node.Array.(*IdentNode); ok {
		arrName = id.Name
	} else {
		v, err := g.visit(node.Array)
		if err != nil {
			return "", err
		}
		arrName = v
	}
	idx, err := g.visit(node.Index)
	if err != nil {
		return "", err
	}
	t := g.newTemp()
	g.emit(IDX, arrName, idx, t)
	return t, nil
}

func (g *Generator) visitBinaryOp(node *BinaryOpNode) (string, error) {
	left, err := g.visit(node.Left)
	if err != nil {
		return "", err
	}
	right, err := g.visit(node.Right)
	if err != nil {
		return "", err
	}
	t := g.newTemp()
	switch node.Op {
	case "+", "-", "*", "/", "mod":
		g.emit(AOP, node.Op, left, right, t)
	default:
		g.emit(COM, node.Op, left, right, t)
	}
	return t, nil
}

func (g *Generator) visitUnaryOp(node *UnaryOpNode) (string, error) {
	operand, err := g.visit(node.Operand)
	if err != nil {
		return "", err
	}
	t := g.newTemp()
	g.emit(AOP, node.Op, operand, t)
	return t, nil
}

func (g *Generator) visitAssign(node *AssignNode) error {
	val, err := g.visit(node.Value)
	if err != nil {
		return err
	}
	switch target := node.Target.(type) {
	case *IdentNode:
		g.emit(ASN, target.Name, val)
	case *ArrayAccessNode:
		var arrName string
		if id, ok := target.Array.(*IdentNode); ok {
			arrName = id.Name
		} else {
			v, err := g.visit(target.Array)
			if err != nil {
				return err
			}
			arrName = v
		}
		idx, err := g.visit(target.Index)
		if err != nil {
			return err
		}
		g.emit(ASN, arrName+"["+idx+"]", val)
	default:
		return fmt.Errorf("invalid assignment target %T", node.Target)
	}
	return nil
}

func (g *Generator) visitIf(node *IfNode) error {
	if _, err := g.visit(node.Cond); err != nil {
		return err
	}
	skpIdx := len(g.instrs)
	g.emit(SKP, "0")
	thenStart := len(g.instrs)

	if err := g.visitBlock(node.Then); err != nil {
		return err
	}

	if node.Else != nil {
		jmpIdx := len(g.instrs)
		g.emit(JMP, "0")
		thenSize := jmpIdx - thenStart + 1
		g.instrs[skpIdx] = Instruction{SKP, []string{strconv.Itoa(thenSize)}}

		if err := g.visitBlock(node.Else); err != nil {
			return err
		}
		endIdx := len(g.instrs)
		g.instrs[jmpIdx] = Instruction{JMP, []string{strconv.Itoa(endIdx)}}
	} else {
		thenSize := len(g.instrs) - thenStart
		g.instrs[skpIdx] = Instruction{SKP, []string{strconv.Itoa(thenSize)}}
	}
	return nil
}

func (g *Generator) visitWhile(node *WhileNode) error {
	loopStart := len(g.instrs)
	if _, err := g.visit(node.Cond); err != nil {
		return err
	}
	skpIdx := len(g.instrs)
	g.emit(SKP, "0")
	bodyStart := len(g.instrs)

	if err := g.visitBlock(node.Body); err != nil {
		return err
	}
	g.emit(JMP, strconv.Itoa(loopStart))

	bodySize := len(g.instrs) - bodyStart
	g.instrs[skpIdx] = Instruction{SKP, []string{strconv.Itoa(bodySize)}}
	return nil
}

func (g *Generator) visitFor(node *ForNode) error {
	if err := g.visitAssign(node.Init.(*AssignNode)); err != nil {
		return err
	}
	loopVar := node.Init.(*AssignNode).Target.(*IdentNode).Name

	loopStart := len(g.instrs)
	endVal, err := g.visit(node.End)
	if err != nil {
		return err
	}
	tmpCmp := g.newTemp()
	g.emit(COM, "<=", loopVar, endVal, tmpCmp)

	skpIdx := len(g.instrs)
	g.emit(SKP, "0")
	bodyStart := len(g.instrs)

	if err := g.visitBlock(node.Body); err != nil {
		return err
	}

	tmpInc := g.newTemp()
	g.emit(AOP, "+", loopVar, "1", tmpInc)
	g.emit(ASN, loopVar, tmpInc)
	g.emit(JMP, strconv.Itoa(loopStart))

	bodySize := len(g.instrs) - bodyStart
	g.instrs[skpIdx] = Instruction{SKP, []string{strconv.Itoa(bodySize)}}
	return nil
}

func (g *Generator) visitRepeat(node *RepeatNode) error {
	bodyStart := len(g.instrs)
	if err := g.visitBlock(node.Body); err != nil {
		return err
	}
	if _, err := g.visit(node.Cond); err != nil {
		return err
	}
	g.emit(SKPT, "1")
	g.emit(JMP, strconv.Itoa(bodyStart))
	return nil
}

func (g *Generator) visitReturn(node *ReturnNode) error {
	if node.Value != nil {
		val, err := g.visit(node.Value)
		if err != nil {
			return err
		}
		g.emit(RET, val)
	} else {
		g.emit(RET)
	}
	return nil
}
