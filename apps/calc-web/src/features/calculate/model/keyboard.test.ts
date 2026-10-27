import { expect, test } from "vitest";
import { keyboardAction } from "./keyboard.ts";

test("maps number keys and all supported operators without numeric conversion", () => {
  for (const [key, value] of [
    ["9", "9"],
    [".", "."],
    ["+", "+"],
    ["-", "−"],
    ["*", "×"],
    ["/", "÷"],
    ["^", "^"],
    ["%", "%"],
    ["r", "√("],
    ["(", "("],
    [")", ")"],
  ]) {
    expect(keyboardAction(key)).toEqual({ type: "insert", value });
  }
});

test("maps editing shortcuts and ignores unrelated keys", () => {
  expect(keyboardAction("Escape")).toEqual({ type: "clear" });
  expect(keyboardAction("Backspace")).toEqual({ type: "backspace" });
  expect(keyboardAction("Delete")).toEqual({ type: "backspace" });
  for (const key of ["Tab", "a", "ArrowLeft", "F5"])
    expect(keyboardAction(key)).toBeNull();
});
