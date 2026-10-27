import { expect, test } from "vitest";
import { calculate } from "./calculate.ts";

test("the preview returns the string 1 for every supported operation", () => {
  for (const expression of [
    "2+3",
    "5−2",
    "4×6",
    "8÷2",
    "2^8",
    "√(9)",
    "25%",
    "1÷0",
    "999999999999999999999999999999999999999.000000000001+1",
  ]) {
    expect(calculate(expression)).toBe("1");
  }
});
