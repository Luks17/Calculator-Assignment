import { expect, test } from "vitest";
import {
  calculatorReducer as reduce,
  initialCalculatorState as initial,
} from "./calculator.reducer.ts";
import type { CalculatorState } from "./calculator.types.ts";

function enter(expression: string): CalculatorState {
  return reduce(initial, { type: "edit", expression });
}

function evaluate(expression: string): CalculatorState {
  const state = enter(expression);
  return reduce(state, {
    type: "evaluate",
    expression: state.expression,
    result: "1",
  });
}

test("starts with an empty expression and no fake history", () => {
  expect(initial.expression).toBe("");
  expect(initial.result).toBeNull();
  expect(initial.history).toEqual([]);
});

test("retains long numbers, fractional digits and trailing zeros as strings", () => {
  const expression =
    "900719925474099312345678901234567890.00000000000000000100";
  let state = initial;
  for (const value of expression)
    state = reduce(state, { type: "insert", value });
  expect(state.expression).toBe(expression);
  state = reduce(state, {
    type: "evaluate",
    expression: state.expression,
    result: "1",
  });
  expect(state.history[0].expression).toBe(expression);
  expect(state.history[0].result).toBeTypeOf("string");
});

test("normalizes typed and pasted arithmetic notation, not numeric values", () => {
  expect(enter("sqrt(81) * 2 - 3 / 4").expression).toBe("√(81) × 2 − 3 ÷ 4");
  expect(enter("<script>alert('no')</script>1+2").expression).toBe("()÷1+2");
});

test("evaluating replaces the expression with the result and saves the original in history", () => {
  const state = evaluate("12×3");
  expect(state.expression).toBe("1");
  expect(state.result).toBe("1");
  expect(state.history).toEqual([{ id: 1, expression: "12×3", result: "1" }]);
});

test("promotes the exact string result without losing precision", () => {
  const result = "900719925474099312345678901234567890.00000000000100";
  const state = reduce(enter("56+58"), {
    type: "evaluate",
    expression: "56+58",
    result,
  });
  expect(state.expression).toBe(result);
  expect(state.history[0]).toEqual({ id: 1, expression: "56+58", result });
});

test("a calculation can continue from the promoted result", () => {
  let state = evaluate("56+58");
  state = reduce(state, { type: "insert", value: "+" });
  state = reduce(state, { type: "insert", value: "2" });
  expect(state.expression).toBe("1+2");
  state = reduce(state, {
    type: "evaluate",
    expression: state.expression,
    result: "1",
  });
  expect(state.expression).toBe("1");
  expect(state.history.map((item) => item.expression)).toEqual([
    "1+2",
    "56+58",
  ]);
});

test("backspace edits the promoted result, not the original expression", () => {
  const state = reduce(evaluate("56+58"), { type: "backspace" });
  expect(state.expression).toBe("");
  expect(state.result).toBeNull();
  expect(state.history[0].expression).toBe("56+58");
});

test("ignores empty submission and repeated equals", () => {
  expect(
    reduce(initial, { type: "evaluate", expression: "", result: "1" }),
  ).toBe(initial);
  expect(
    reduce(enter("   "), { type: "evaluate", expression: "   ", result: "1" })
      .history,
  ).toHaveLength(0);
  const state = evaluate("3+5");
  expect(
    reduce(state, {
      type: "evaluate",
      expression: state.expression,
      result: "1",
    }),
  ).toBe(state);
});

test("ignores a response when the expression was edited, cleared or recalled while waiting", () => {
  const pending = enter("2+3");
  for (const state of [
    reduce(pending, { type: "edit", expression: "4+5" }),
    reduce(pending, { type: "clear" }),
    reduce(pending, { type: "recall", expression: "6+7" }),
  ]) {
    expect(
      reduce(state, { type: "evaluate", expression: "2+3", result: "5" }),
    ).toBe(state);
    expect(state.history).toEqual([]);
  }
});

test("negative API results use display notation and can be continued or sign-toggled", () => {
  const state = reduce(enter("2−7"), {
    type: "evaluate",
    expression: "2−7",
    result: "-5",
  });
  expect(state.expression).toBe("−5");
  expect(state.result).toBe("-5");
  expect(state.history[0].result).toBe("-5");
  expect(reduce(state, { type: "toggle-sign" }).expression).toBe("5");
  expect(reduce(state, { type: "insert", value: "×" }).expression).toBe("−5×");
});

test("a digit starts a fresh expression after a result", () => {
  expect(
    reduce(evaluate("2+3"), { type: "insert", value: "7" }).expression,
  ).toBe("7");
});

test("an operator continues from the string result", () => {
  const state = reduce(evaluate("2+3"), { type: "insert", value: "×" });
  expect(state.expression).toBe("1×");
  expect(state.result).toBeNull();
});

test("prevents duplicate decimal points and supplies leading zero", () => {
  expect(reduce(initial, { type: "insert", value: "." }).expression).toBe("0.");
  expect(reduce(enter("1.25"), { type: "insert", value: "." }).expression).toBe(
    "1.25",
  );
  expect(
    reduce(enter("1.25+"), { type: "insert", value: "." }).expression,
  ).toBe("1.25+0.");
});

test("replaces a trailing operator and allows negative operands", () => {
  expect(reduce(enter("2+"), { type: "insert", value: "×" }).expression).toBe(
    "2×",
  );
  expect(reduce(enter("2×"), { type: "insert", value: "−" }).expression).toBe(
    "2×−",
  );
  expect(reduce(initial, { type: "insert", value: "÷" }).expression).toBe("");
});

test("backspace removes one character or an empty square-root entry", () => {
  expect(reduce(enter("123"), { type: "backspace" }).expression).toBe("12");
  expect(reduce(enter("2+√("), { type: "backspace" }).expression).toBe("2+");
  expect(reduce(initial, { type: "backspace" }).expression).toBe("");
});

test("clear resets the display but preserves session history", () => {
  const state = reduce(evaluate("5%"), { type: "clear" });
  expect(state.expression).toBe("");
  expect(state.result).toBeNull();
  expect(state.history).toHaveLength(1);
});

test("sign toggle keeps precision and is reversible", () => {
  let state = enter("12+999999999999999999.00100");
  state = reduce(state, { type: "toggle-sign" });
  expect(state.expression).toBe("12+−999999999999999999.00100");
  state = reduce(state, { type: "toggle-sign" });
  expect(state.expression).toBe("12+999999999999999999.00100");
  expect(reduce(evaluate("3+4"), { type: "toggle-sign" }).expression).toBe(
    "−1",
  );
});

test("can recall an expression without changing history", () => {
  const state = reduce(evaluate("2^8"), {
    type: "recall",
    expression: "√(81)",
  });
  expect(state.expression).toBe("√(81)");
  expect(state.result).toBeNull();
  expect(state.history).toHaveLength(1);
});

test("history is newest first, bounded to 20, and separately clearable", () => {
  let state = initial;
  for (let index = 0; index < 25; index++) {
    state = reduce(state, { type: "edit", expression: `${index}+1` });
    state = reduce(state, {
      type: "evaluate",
      expression: state.expression,
      result: "1",
    });
  }
  expect(state.history).toHaveLength(20);
  expect(state.history[0].expression).toBe("24+1");
  expect(state.history.at(-1)?.expression).toBe("5+1");
  expect(new Set(state.history.map((item) => item.id)).size).toBe(20);
  const cleared = reduce(state, { type: "clear-history" });
  expect(cleared.history).toEqual([]);
  expect(cleared.expression).toBe(state.expression);
  expect(cleared.result).toBe(state.result);
});

test("does not mutate previous state", () => {
  const before = enter("7+8");
  const after = reduce(before, {
    type: "evaluate",
    expression: before.expression,
    result: "1",
  });
  expect(after).not.toBe(before);
  expect(before.result).toBeNull();
  expect(before.history).toEqual([]);
});
