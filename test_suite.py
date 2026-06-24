import argparse
import re
import subprocess
import sys
from dataclasses import dataclass, field
from pathlib import Path


SEP = "=" * 60

@dataclass
class Case:
    name: str
    file: str
    inputs: list[str]
    expected_return: str
    expected_prints: list[str] = field(default_factory=list)
    description: str = ""


CASES: list[Case] = [
    Case("max_basic",        "demos/01.psu", ["[5,3,8,1,9]", "5"], "9"),
    Case("max_single",       "demos/01.psu", ["[1]", "1"],          "1"),
    Case("max_negatives",    "demos/01.psu", ["[-5,-2,-8]", "3"],   "-2"),
    Case("max_all_same",     "demos/01.psu", ["[4,4,4]", "3"],      "4"),

    Case("min_basic",        "demos/02.psu", ["[5,3,8,1,9]", "5"], "1"),
    Case("min_single",       "demos/02.psu", ["[1]", "1"],          "1"),
    Case("min_negatives",    "demos/02.psu", ["[-5,-2,-8]", "3"],   "-8"),
    Case("min_all_same",     "demos/02.psu", ["[7,7,7]", "3"],      "7"),

    Case("sum_basic",        "demos/03.psu", ["[1,2,3,4,5]", "5"], "15"),
    Case("sum_zeros",        "demos/03.psu", ["[0,0,0]", "3"],      "0"),
    Case("sum_single",       "demos/03.psu", ["[42]", "1"],         "42"),
    Case("sum_negatives",    "demos/03.psu", ["[-1,-2,-3]", "3"],   "-6"),

    Case("search_found_mid",  "demos/04.psu", ["[1,3,5,7,9]", "5", "5"],  "2"),
    Case("search_found_first","demos/04.psu", ["[1,3,5,7,9]", "5", "1"],  "0"),
    Case("search_found_last", "demos/04.psu", ["[1,3,5,7,9]", "5", "9"],  "4"),
    Case("search_not_found",  "demos/04.psu", ["[1,3,5,7,9]", "5", "4"],  "-1"),

    Case("sort_basic",       "demos/05.psu", ["[5,2,8,1,9]", "5"],  "[1, 2, 5, 8, 9]"),
    Case("sort_already",     "demos/05.psu", ["[1,2,3]", "3"],      "[1, 2, 3]"),
    Case("sort_reverse",     "demos/05.psu", ["[3,2,1]", "3"],      "[1, 2, 3]"),
    Case("sort_single",      "demos/05.psu", ["[7]", "1"],          "[7]"),

    Case("mod_evens_mixed",  "demos/06_mod_boolean.psu", ["[2,3,4,0,6,7,8]", "7"], "4",
         description="2,4,6,8 are non-zero evens"),
    Case("mod_evens_none",   "demos/06_mod_boolean.psu", ["[1,3,5]", "3"],          "0"),
    Case("mod_evens_all",    "demos/06_mod_boolean.psu", ["[2,4,6]", "3"],          "3"),
    Case("mod_zero_excluded","demos/06_mod_boolean.psu", ["[0,0,0]", "3"],          "0",
         description="0 is even but excluded by not (A[i] = 0)"),

    Case("repeat_100",  "demos/07_repeat_until.psu", ["100"], "128"),
    Case("repeat_0",    "demos/07_repeat_until.psu", ["0"],   "2",
         description="p doubles once (1->2) then 2>0 triggers exit"),
    Case("repeat_1",    "demos/07_repeat_until.psu", ["1"],   "2"),
    Case("repeat_2",    "demos/07_repeat_until.psu", ["2"],   "4"),
    Case("repeat_127",  "demos/07_repeat_until.psu", ["127"], "128"),
    Case("repeat_128",  "demos/07_repeat_until.psu", ["128"], "256"),

    Case("downto_reverse_5", "demos/08_downto.psu", ["[1,2,3,4,5]", "5"],
         "[5, 4, 3, 2, 1]",
         expected_prints=["1", "2", "3", "4", "5"],
         description="reversed array; prints reversed-array read downto = original order"),
    Case("downto_reverse_1", "demos/08_downto.psu", ["[42]", "1"],
         "[42]",
         expected_prints=["42"]),

    Case("bsearch_found_11", "demos/09_print.psu", ["[1,3,5,7,9,11,13]", "7", "11"],
         "5",
         expected_prints=["3", "5"],
         description="mid probes: 3 then 5"),
    Case("bsearch_found_1",  "demos/09_print.psu", ["[1,3,5,7,9,11,13]", "7", "1"],
         "0",
         expected_prints=["3", "1", "0"]),
    Case("bsearch_not_found","demos/09_print.psu", ["[1,3,5,7,9,11,13]", "7", "4"],
         "-1"),

    Case("gcd_12_8",   "demos/10_function_call.psu", ["12", "8"],  "4"),
    Case("gcd_48_18",  "demos/10_function_call.psu", ["48", "18"], "6"),
    Case("gcd_7_3",    "demos/10_function_call.psu", ["7",  "3"],  "1"),
    Case("gcd_1_1",    "demos/10_function_call.psu", ["1",  "1"],  "1"),
    Case("gcd_prime",  "demos/10_function_call.psu", ["13", "7"],  "1"),

    Case("power_2_10",  "demos/11_recursion.psu", ["2",  "10"], "1024"),
    Case("power_3_4",   "demos/11_recursion.psu", ["3",  "4"],  "81"),
    Case("power_5_0",   "demos/11_recursion.psu", ["5",  "0"],  "1",
         description="base case: exp=0 returns 1"),
    Case("power_1_100", "demos/11_recursion.psu", ["1",  "100"], "1"),
    Case("power_10_3",  "demos/11_recursion.psu", ["10", "3"],   "1000"),

    Case("builtins_mixed", "demos/12_builtins.psu", ["[-3,7,-1,4,2]", "5"],
         "17",
         expected_prints=["-3", "7", "3"],
         description="sum of abs values=17; prints min(-3,0)=-3, max(7,0)=7, floor(17/5)=3"),

    Case("nil_with_positives",   "demos/13_literals.psu", ["[-5,-2,3,1,7]", "5"], "1"),
    Case("nil_no_positives",     "demos/13_literals.psu", ["[-5,-2,-3]", "3"],     "None",
         description="all negative -> returns NIL (prints as None)"),
    Case("nil_single_positive",  "demos/13_literals.psu", ["[1]", "1"],            "1"),
    Case("nil_single_negative",  "demos/13_literals.psu", ["[-1]", "1"],           "None"),

    Case("fib_0",   "demos/14_fibonacci.psu", ["0"],  "0"),
    Case("fib_1",   "demos/14_fibonacci.psu", ["1"],  "1"),
    Case("fib_5",   "demos/14_fibonacci.psu", ["5"],  "5"),
    Case("fib_10",  "demos/14_fibonacci.psu", ["10"], "55"),
]



def run_case(case: Case, runner: list[str], repo_root: Path) -> tuple[str, list[str], str]:
    cmd = runner + [str(repo_root / case.file)]
    stdin_data = "\n".join(case.inputs) + "\n"

    result = subprocess.run(
        cmd,
        input=stdin_data,
        capture_output=True,
        text=True,
        encoding="utf-8",
        cwd=str(repo_root),
    )

    raw = result.stdout
    if result.returncode != 0:
        raise RuntimeError(
            f"Interpreter exited with code {result.returncode}.\n"
            f"stderr: {result.stderr}\nstdout: {raw}"
        )

    return_value, print_lines = parse_output(raw)
    return return_value, print_lines, raw


def parse_output(output: str) -> tuple[str, list[str]]:
    parts = output.split(SEP)

    return_value = ""
    instr_section = None
    for part in parts:
        for line in part.splitlines():
            if line.strip().startswith("Return value:"):
                return_value = line.strip()[len("Return value:"):].strip()
        if re.search(r"^\s*\d+:", part, re.MULTILINE):
            instr_section = part

    print_lines = []
    if instr_section:
        for line in instr_section.splitlines():
            stripped = line.strip()
            if stripped and not re.match(r"^\d+:", stripped):
                print_lines.append(stripped)

    return return_value, print_lines



class TestResult:
    def __init__(self, case: Case):
        self.case = case
        self.passed = False
        self.failures: list[str] = []
        self.error: str | None = None

    def __repr__(self):
        status = "PASS" if self.passed else "FAIL"
        return f"{status}  {self.case.name}"


def run_all(runner: list[str], repo_root: Path) -> list[TestResult]:
    results = []
    for case in CASES:
        tr = TestResult(case)
        try:
            return_value, print_lines, _ = run_case(case, runner, repo_root)

            if return_value != case.expected_return:
                tr.failures.append(
                    f"  return: got {return_value!r}, expected {case.expected_return!r}"
                )

            if case.expected_prints and print_lines != case.expected_prints:
                tr.failures.append(
                    f"  prints: got {print_lines}, expected {case.expected_prints}"
                )

            tr.passed = not tr.failures

        except Exception as exc:
            tr.error = str(exc)

        results.append(tr)

    return results


def print_summary(results: list[TestResult]) -> int:
    passed = sum(1 for r in results if r.passed)
    failed = [r for r in results if not r.passed]

    col_w = max(len(r.case.name) for r in results) + 2

    for r in results:
        status = "\033[32mPASS\033[0m" if r.passed else "\033[31mFAIL\033[0m"
        desc = f"  # {r.case.description}" if r.case.description else ""
        print(f"  {status}  {r.case.name:<{col_w}}{desc}")

    print()
    print(f"Results: {passed}/{len(results)} passed")

    if failed:
        print()
        for r in failed:
            print(f"\033[31m{'-' * 60}\033[0m")
            print(f"FAILED: {r.case.name}")
            if r.case.description:
                print(f"  ({r.case.description})")
            print(f"  file:   {r.case.file}")
            print(f"  inputs: {r.case.inputs}")
            if r.error:
                print(f"  error:  {r.error}")
            for f in r.failures:
                print(f)
        return 1

    return 0


def _parse_args(argv=None):
    p = argparse.ArgumentParser(description="Pseudocode interpreter test suite")
    p.add_argument(
        "--runner",
        default="./pseudo",
        help='Command to invoke the interpreter, e.g. "./pseudo" or "python main.py"',
    )
    p.add_argument(
        "--root",
        default=".",
        help="Repository root directory (default: current directory)",
    )
    return p.parse_args(argv)


def main(argv=None):
    args = _parse_args(argv)
    runner = args.runner.split()
    repo_root = Path(args.root).resolve()

    print(f"Runner : {' '.join(runner)}")
    print(f"Root   : {repo_root}")
    print(f"Cases  : {len(CASES)}")
    print()

    results = run_all(runner, repo_root)
    sys.exit(print_summary(results))


def pytest_addoption(parser):
    parser.addoption(
        "--runner",
        action="store",
        default="./pseudo",
        help="Interpreter command string",
    )


def pytest_configure(config):
    pass


try:
    import pytest

    @pytest.fixture(scope="session")
    def interpreter_runner(request):
        return request.config.getoption("--runner", default="python main.py").split()

    @pytest.fixture(scope="session")
    def repo_root():
        return Path(__file__).parent.resolve()

    @pytest.mark.parametrize("case", CASES, ids=lambda c: c.name)
    def test_case(case: Case, interpreter_runner, repo_root):
        return_value, print_lines, raw = run_case(case, interpreter_runner, repo_root)

        assert return_value == case.expected_return, (
            f"Return value mismatch.\n"
            f"  got:      {return_value!r}\n"
            f"  expected: {case.expected_return!r}\n"
            f"  inputs:   {case.inputs}\n"
            f"  raw output:\n{raw}"
        )

        if case.expected_prints:
            assert print_lines == case.expected_prints, (
                f"Print output mismatch.\n"
                f"  got:      {print_lines}\n"
                f"  expected: {case.expected_prints}\n"
                f"  raw output:\n{raw}"
            )

except ImportError:
    pass


if __name__ == "__main__":
    main()
