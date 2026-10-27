/**
 * Preview-only calculation boundary. Expressions and results stay strings;
 * the future backend will own parsing and arithmetic, not this frontend.
 */
export function calculate(expression: string): string {
  void expression;
  return "1";
}
