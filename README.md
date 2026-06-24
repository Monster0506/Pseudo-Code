# Pseudo

An interpreter for academic-style pseudocode, written in Go.

See `demos/` for example programs.

## Running

```bash
go build -o pseudo .
./pseudo demos/01.psu
```

You will be prompted for the input parameters of the entry Algorithm.

## Testing

```bash
python test_suite.py
```

Requires Python 3.11+ and [uv](https://github.com/astral-sh/uv). Run `uv sync` first if needed.

## Language

```
Algorithm name(param1, param2) do
    x <- 0
    for i <- 0 to n - 1 do
        if A[i] > x then
            x <- A[i]
        end
    end
    return x
end
```

- Assignment: `<-`
- Arithmetic: `+ - * / mod`
- Comparison: `= != < > <= >=`
- Logic: `and or not`
- Control: `if/then/else/end`, `while/do/end`, `for/to/downto/do/end`, `repeat/until`
- Print: `print <expr>`
- Literals: `true`, `false`, `NIL`, `infinity`
- Builtins: `length`, `abs`, `floor`, `ceil`, `sqrt`, `min`, `max`
