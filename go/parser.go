package main

import "fmt"

type Node interface{ nodeTag() }

type LiteralNode struct{ Value string }
type IdentNode struct{ Name string }
type ArrayLitNode struct{ Elems []Node }
type ArrayAccessNode struct {
	Array Node
	Index Node
}
type FuncCallNode struct {
	Name string
	Args []Node
}
type BinaryOpNode struct {
	Left, Right Node
	Op          string
}
type UnaryOpNode struct {
	Op      string
	Operand Node
}
type AssignNode struct {
	Target Node
	Value  Node
}
type IfNode struct {
	Cond Node
	Then *BlockNode
	Else *BlockNode
}
type WhileNode struct {
	Cond Node
	Body *BlockNode
}
type ForNode struct {
	Init      Node
	End       Node
	Body      *BlockNode
	Direction string
	Step      Node
}
type PrintNode struct{ Expr Node }
type RepeatNode struct {
	Body *BlockNode
	Cond Node
}
type ReturnNode struct{ Value Node }
type BlockNode struct{ Stmts []Node }
type AlgoNode struct {
	Name   string
	Params []string
	Body   *BlockNode
}

func (n *LiteralNode) nodeTag()     {}
func (n *IdentNode) nodeTag()       {}
func (n *ArrayLitNode) nodeTag()    {}
func (n *ArrayAccessNode) nodeTag() {}
func (n *FuncCallNode) nodeTag()    {}
func (n *BinaryOpNode) nodeTag()    {}
func (n *UnaryOpNode) nodeTag()     {}
func (n *AssignNode) nodeTag()      {}
func (n *IfNode) nodeTag()          {}
func (n *WhileNode) nodeTag()       {}
func (n *ForNode) nodeTag()         {}
func (n *PrintNode) nodeTag()       {}
func (n *RepeatNode) nodeTag()      {}
func (n *ReturnNode) nodeTag()      {}
func (n *BlockNode) nodeTag()       {}
func (n *AlgoNode) nodeTag()        {}

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser { return &Parser{tokens: tokens} }

func (p *Parser) cur() *Token {
	if p.pos < len(p.tokens) {
		return &p.tokens[p.pos]
	}
	return nil
}

func (p *Parser) peek(offset int) *Token {
	i := p.pos + offset
	if i < len(p.tokens) {
		return &p.tokens[i]
	}
	return nil
}

func (p *Parser) advance() { p.pos++ }

func (p *Parser) consume(expected string) (Token, error) {
	t := p.cur()
	if t == nil {
		return Token{}, fmt.Errorf("unexpected end of input, expected %q", expected)
	}
	if expected != "" && t.Value != expected {
		return Token{}, fmt.Errorf("expected %q, got %q", expected, t.Value)
	}
	p.advance()
	return *t, nil
}

func (p *Parser) Parse() (*BlockNode, error) {
	var stmts []Node
	for p.cur() != nil {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmts = append(stmts, s)
		}
	}
	return &BlockNode{stmts}, nil
}

func (p *Parser) parseStatement() (Node, error) {
	t := p.cur()
	if t == nil {
		return nil, nil
	}
	switch t.Value {
	case "if":
		return p.parseIf()
	case "while":
		return p.parseWhile()
	case "for":
		return p.parseFor()
	case "print", "output":
		return p.parsePrint()
	case "repeat":
		return p.parseRepeat()
	case "return":
		return p.parseReturn()
	case "Algorithm":
		return p.parseAlgo()
	}
	if t.Type == IDENTIFIER {
		return p.parseAssignOrExpr()
	}
	return nil, fmt.Errorf("unexpected token: %q", t.Value)
}

func (p *Parser) parseIf() (Node, error) {
	if _, err := p.consume("if"); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	t := p.cur()
	if t == nil || (t.Value != "then" && t.Value != "do") {
		got := "<EOF>"
		if t != nil {
			got = t.Value
		}
		return nil, fmt.Errorf("expected 'then', got %q", got)
	}
	p.advance()

	var thenStmts []Node
	for p.cur() != nil && p.cur().Value != "else" && p.cur().Value != "end" {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			thenStmts = append(thenStmts, s)
		}
	}
	var elseBlock *BlockNode
	if p.cur() != nil && p.cur().Value == "else" {
		p.advance()
		var elseStmts []Node
		for p.cur() != nil && p.cur().Value != "end" {
			s, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if s != nil {
				elseStmts = append(elseStmts, s)
			}
		}
		elseBlock = &BlockNode{elseStmts}
	}
	if _, err := p.consume("end"); err != nil {
		return nil, err
	}
	return &IfNode{cond, &BlockNode{thenStmts}, elseBlock}, nil
}

func (p *Parser) parseWhile() (Node, error) {
	if _, err := p.consume("while"); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.consume("do"); err != nil {
		return nil, err
	}
	var stmts []Node
	for p.cur() != nil && p.cur().Value != "end" {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmts = append(stmts, s)
		}
	}
	if _, err := p.consume("end"); err != nil {
		return nil, err
	}
	return &WhileNode{cond, &BlockNode{stmts}}, nil
}

func (p *Parser) parseFor() (Node, error) {
	if _, err := p.consume("for"); err != nil {
		return nil, err
	}
	init, err := p.parseAssignOrExpr()
	if err != nil {
		return nil, err
	}
	direction := "to"
	if p.cur() != nil && p.cur().Value == "downto" {
		direction = "downto"
		p.advance()
	} else {
		if _, err := p.consume("to"); err != nil {
			return nil, err
		}
	}
	end, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.consume("do"); err != nil {
		return nil, err
	}
	var stmts []Node
	for p.cur() != nil && p.cur().Value != "end" {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmts = append(stmts, s)
		}
	}
	if _, err := p.consume("end"); err != nil {
		return nil, err
	}
	return &ForNode{init, end, &BlockNode{stmts}, direction, nil}, nil
}

func (p *Parser) parsePrint() (Node, error) {
	p.advance()
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &PrintNode{expr}, nil
}

func (p *Parser) parseRepeat() (Node, error) {
	if _, err := p.consume("repeat"); err != nil {
		return nil, err
	}
	var stmts []Node
	for p.cur() != nil && p.cur().Value != "until" {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmts = append(stmts, s)
		}
	}
	if _, err := p.consume("until"); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &RepeatNode{&BlockNode{stmts}, cond}, nil
}

func (p *Parser) parseReturn() (Node, error) {
	if _, err := p.consume("return"); err != nil {
		return nil, err
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ReturnNode{val}, nil
}

func (p *Parser) parseAlgo() (Node, error) {
	if _, err := p.consume("Algorithm"); err != nil {
		return nil, err
	}
	nameTok, err := p.consume("")
	if err != nil {
		return nil, err
	}
	if nameTok.Type != IDENTIFIER {
		return nil, fmt.Errorf("expected algorithm name, got %q", nameTok.Value)
	}
	if _, err := p.consume("("); err != nil {
		return nil, err
	}
	var params []string
	if p.cur() != nil && p.cur().Value != ")" {
		t, err := p.consume("")
		if err != nil {
			return nil, err
		}
		params = append(params, t.Value)
		for p.cur() != nil && p.cur().Value == "," {
			p.advance()
			t, err = p.consume("")
			if err != nil {
				return nil, err
			}
			params = append(params, t.Value)
		}
	}
	if _, err := p.consume(")"); err != nil {
		return nil, err
	}
	if _, err := p.consume("do"); err != nil {
		return nil, err
	}
	var stmts []Node
	for p.cur() != nil && p.cur().Value != "end" {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmts = append(stmts, s)
		}
	}
	if _, err := p.consume("end"); err != nil {
		return nil, err
	}
	return &AlgoNode{nameTok.Value, params, &BlockNode{stmts}}, nil
}

func (p *Parser) parseAssignOrExpr() (Node, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.cur() != nil && p.cur().Value == "<-" {
		p.advance()
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &AssignNode{expr, val}, nil
	}
	return expr, nil
}

func (p *Parser) parseExpr() (Node, error) { return p.parseOr() }

func (p *Parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.cur() != nil && p.cur().Value == "or" {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{left, right, "or"}
	}
	return left, nil
}

func (p *Parser) parseAnd() (Node, error) {
	left, err := p.parseCmp()
	if err != nil {
		return nil, err
	}
	for p.cur() != nil && p.cur().Value == "and" {
		p.advance()
		right, err := p.parseCmp()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{left, right, "and"}
	}
	return left, nil
}

var cmpOps = map[string]bool{"<": true, ">": true, "<=": true, ">=": true, "=": true, "!=": true}

func (p *Parser) parseCmp() (Node, error) {
	left, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	for p.cur() != nil && cmpOps[p.cur().Value] {
		op := p.cur().Value
		p.advance()
		right, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{left, right, op}
	}
	return left, nil
}

func (p *Parser) parseAdd() (Node, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for p.cur() != nil && (p.cur().Value == "+" || p.cur().Value == "-") {
		op := p.cur().Value
		p.advance()
		right, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{left, right, op}
	}
	return left, nil
}

func (p *Parser) parseMul() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur() != nil && (p.cur().Value == "*" || p.cur().Value == "/" || p.cur().Value == "mod") {
		op := p.cur().Value
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{left, right, op}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Node, error) {
	if p.cur() != nil && p.cur().Value == "not" {
		p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryOpNode{"not", operand}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.cur() != nil {
		if p.cur().Value == "[" {
			p.advance()
			idx, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.consume("]"); err != nil {
				return nil, err
			}
			expr = &ArrayAccessNode{expr, idx}
		} else if p.cur().Value == "(" {
			ident, ok := expr.(*IdentNode)
			if !ok {
				break
			}
			p.advance()
			var args []Node
			if p.cur() != nil && p.cur().Value != ")" {
				a, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, a)
				for p.cur() != nil && p.cur().Value == "," {
					p.advance()
					a, err = p.parseExpr()
					if err != nil {
						return nil, err
					}
					args = append(args, a)
				}
			}
			if _, err := p.consume(")"); err != nil {
				return nil, err
			}
			expr = &FuncCallNode{ident.Name, args}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	t := p.cur()
	if t == nil {
		return nil, fmt.Errorf("unexpected end of input")
	}
	if t.Type == LITERAL {
		p.advance()
		return &LiteralNode{t.Value}, nil
	}
	if t.Type == IDENTIFIER {
		p.advance()
		return &IdentNode{t.Value}, nil
	}
	if t.Value == "(" {
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(")"); err != nil {
			return nil, err
		}
		return expr, nil
	}
	if t.Value == "[" {
		return p.parseArrayLit()
	}
	return nil, fmt.Errorf("unexpected token in expression: %q", t.Value)
}

func (p *Parser) parseArrayLit() (Node, error) {
	p.advance()
	var elems []Node
	if p.cur() != nil && p.cur().Value != "]" {
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		elems = append(elems, e)
		for p.cur() != nil && p.cur().Value == "," {
			p.advance()
			if p.cur() != nil && p.cur().Value == "]" {
				break
			}
			e, err = p.parseExpr()
			if err != nil {
				return nil, err
			}
			elems = append(elems, e)
		}
	}
	if _, err := p.consume("]"); err != nil {
		return nil, err
	}
	return &ArrayLitNode{elems}, nil
}
