import type { CalculatorAction } from "./calculator.types.ts";
import { normalizeExpression } from "./expression.ts";

export function keyboardAction(key: string): CalculatorAction | null {
  if (/^[\d.+\-*/×÷−^%()]$/.test(key)) {
    return { type: "insert", value: normalizeExpression(key) };
  }
  if (key.toLowerCase() === "r" || key === "√")
    return { type: "insert", value: "√(" };
  if (key === "Backspace" || key === "Delete") return { type: "backspace" };
  if (key === "Escape") return { type: "clear" };
  return null;
}
