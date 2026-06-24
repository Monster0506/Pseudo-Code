package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type CallFrame struct {
	returnPC     int
	savedVars    map[string]Value
	resultTarget string
}

type VM struct {
	instrs    []Instruction
	funcTable map[string]FuncEntry
	pc        int
	vars      map[string]Value
	lastCmp   bool
	retVal    *Value
	callStack []CallFrame
}

func NewVM(instrs []Instruction, funcTable map[string]FuncEntry) *VM {
	return &VM{
		instrs:    instrs,
		funcTable: funcTable,
		vars:      make(map[string]Value),
	}
}

func (vm *VM) getVal(s string) Value {
	if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
		(strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`)) {
		return IntVal(0)
	}
	if n, err := strconv.Atoi(s); err == nil {
		return IntVal(n)
	}
	if v, ok := vm.vars[s]; ok {
		return v
	}
	return IntVal(0)
}

func (vm *VM) setIndexed(target string, val Value) {
	if !strings.Contains(target, "[") {
		vm.vars[target] = val
		return
	}
	bracket := strings.Index(target, "[")
	arrName := target[:bracket]
	idxStr := target[bracket+1 : len(target)-1]
	idx := toInt(vm.getVal(idxStr))

	arr, ok := vm.vars[arrName]
	if !ok || arr.Kind != TypeArray {
		arr = ArrayVal(nil)
	}
	for len(arr.ArrVal) <= idx {
		arr.ArrVal = append(arr.ArrVal, IntVal(0))
	}
	arr.ArrVal[idx] = val
	vm.vars[arrName] = arr
}

func (vm *VM) execOne(instr Instruction) (int, error) {
	ops := instr.Operands

	switch instr.Opcode {
	case ASN:
		if len(ops) > 2 {
			elems := make([]Value, len(ops)-1)
			for i, e := range ops[1:] {
				elems[i] = vm.getVal(e)
			}
			vm.setIndexed(ops[0], ArrayVal(elems))
		} else {
			vm.setIndexed(ops[0], vm.getVal(ops[1]))
		}

	case AOP:
		op := ops[0]
		if len(ops) == 3 {
			result := !toBool(vm.getVal(ops[1]))
			vm.vars[ops[2]] = boolToVal(result)
			vm.lastCmp = result
		} else {
			l, r := toInt(vm.getVal(ops[1])), toInt(vm.getVal(ops[2]))
			var res int
			switch op {
			case "+":
				res = l + r
			case "-":
				res = l - r
			case "*":
				res = l * r
			case "/":
				if r == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				res = l / r
			case "mod":
				if r == 0 {
					return 0, fmt.Errorf("modulo by zero")
				}
				res = l % r
			default:
				return 0, fmt.Errorf("unknown AOP: %s", op)
			}
			vm.vars[ops[3]] = IntVal(res)
		}

	case COM:
		op, leftOp, rightOp, result := ops[0], ops[1], ops[2], ops[3]
		lv := toInt(vm.getVal(leftOp))
		rv := toInt(vm.getVal(rightOp))
		var cmp bool
		switch op {
		case "<":
			cmp = lv < rv
		case ">":
			cmp = lv > rv
		case "<=":
			cmp = lv <= rv
		case ">=":
			cmp = lv >= rv
		case "=":
			cmp = lv == rv
		case "!=":
			cmp = lv != rv
		case "and":
			cmp = toBool(vm.getVal(leftOp)) && toBool(vm.getVal(rightOp))
		case "or":
			cmp = toBool(vm.getVal(leftOp)) || toBool(vm.getVal(rightOp))
		default:
			cmp = false
		}
		vm.vars[result] = boolToVal(cmp)
		vm.lastCmp = cmp

	case IDX:
		arrName, idxOp, result := ops[0], ops[1], ops[2]
		arr := vm.vars[arrName]
		idx := toInt(vm.getVal(idxOp))
		if arr.Kind == TypeArray && idx >= 0 && idx < len(arr.ArrVal) {
			vm.vars[result] = arr.ArrVal[idx]
		} else {
			vm.vars[result] = IntVal(0)
		}

	case CAL:
		funcName := ops[0]
		argOps := ops[1 : len(ops)-1]
		resultTarget := ops[len(ops)-1]
		argVals := make([]Value, len(argOps))
		for i, a := range argOps {
			argVals[i] = vm.getVal(a)
		}
		if builtin, ok := builtins[funcName]; ok {
			res, err := builtin(argVals)
			if err != nil {
				return 0, err
			}
			vm.vars[resultTarget] = res
			break
		}
		entry, ok := vm.funcTable[funcName]
		if !ok {
			return 0, fmt.Errorf("unknown function: %q", funcName)
		}
		vm.callStack = append(vm.callStack, CallFrame{vm.pc + 1, deepCopyVars(vm.vars), resultTarget})
		newVars := make(map[string]Value, len(entry.ParamNames))
		for i, name := range entry.ParamNames {
			if i < len(argVals) {
				newVars[name] = argVals[i]
			}
		}
		vm.vars = newVars
		return entry.StartIdx - vm.pc - 1, nil

	case PRT:
		fmt.Println(vm.getVal(ops[0]).Format())

	case SKP:
		n, _ := strconv.Atoi(ops[0])
		if !vm.lastCmp {
			return n, nil
		}

	case SKPT:
		n, _ := strconv.Atoi(ops[0])
		if vm.lastCmp {
			return n, nil
		}

	case JMP:
		target, _ := strconv.Atoi(ops[0])
		return target - vm.pc - 1, nil

	case RET:
		var rv Value
		if len(ops) > 0 {
			rv = vm.getVal(ops[0])
		} else {
			rv = Nil
		}
		if len(vm.callStack) > 0 {
			frame := vm.callStack[len(vm.callStack)-1]
			vm.callStack = vm.callStack[:len(vm.callStack)-1]
			vm.vars = frame.savedVars
			if frame.resultTarget != "" {
				vm.vars[frame.resultTarget] = rv
			}
			return frame.returnPC - vm.pc - 1, nil
		}
		vm.retVal = &rv
		return len(vm.instrs), nil
	}
	return 0, nil
}

var builtins = map[string]func([]Value) (Value, error){
	"length": func(args []Value) (Value, error) {
		if args[0].Kind == TypeArray {
			return IntVal(len(args[0].ArrVal)), nil
		}
		return IntVal(1), nil
	},
	"floor": func(args []Value) (Value, error) {
		return IntVal(int(math.Floor(float64(toInt(args[0]))))), nil
	},
	"ceil": func(args []Value) (Value, error) {
		return IntVal(int(math.Ceil(float64(toInt(args[0]))))), nil
	},
	"abs": func(args []Value) (Value, error) {
		n := toInt(args[0])
		if n < 0 {
			n = -n
		}
		return IntVal(n), nil
	},
	"sqrt": func(args []Value) (Value, error) {
		return IntVal(int(math.Sqrt(float64(toInt(args[0]))))), nil
	},
	"min": func(args []Value) (Value, error) {
		m := toInt(args[0])
		for _, a := range args[1:] {
			if v := toInt(a); v < m {
				m = v
			}
		}
		return IntVal(m), nil
	},
	"max": func(args []Value) (Value, error) {
		m := toInt(args[0])
		for _, a := range args[1:] {
			if v := toInt(a); v > m {
				m = v
			}
		}
		return IntVal(m), nil
	},
}

func (vm *VM) Run(initial map[string]Value) (Value, error) {
	for k, v := range initial {
		vm.vars[k] = v
	}
	vm.pc = 0
	vm.lastCmp = false
	vm.retVal = nil

	for vm.pc < len(vm.instrs) {
		delta, err := vm.execOne(vm.instrs[vm.pc])
		if err != nil {
			return Nil, err
		}
		vm.pc += 1 + delta
	}

	if vm.retVal != nil {
		return *vm.retVal, nil
	}
	return Nil, nil
}
