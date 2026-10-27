/** Normalize notation only. Never parse operands into JavaScript numbers. */
export function normalizeExpression(expression: string): string {
  return expression
    .replace(/sqrt\s*\(/gi, "√(")
    .replaceAll("*", "×")
    .replaceAll("/", "÷")
    .replaceAll("-", "−")
    .replace(/[^\d.+×÷−^%()√\s]/g, "");
}

export function insertIntoExpression(
  expression: string,
  value: string,
): string {
  if (value === ".") {
    const currentNumber = expression.match(/[\d.]+$/)?.[0] ?? "";
    if (currentNumber.includes(".")) return expression;
    return expression + (currentNumber ? "." : "0.");
  }

  if (/^[+×÷^−]$/.test(value)) {
    if (!expression) return value === "−" ? value : expression;
    if (/[+×÷^−]$/.test(expression)) {
      // Minus may start a negative operand, e.g. 2 × −3.
      if (value === "−" && !expression.endsWith("−")) return expression + value;
      return expression.replace(/[+×÷^−]+$/, value);
    }
  }

  if (value === "%" && !/[\d)%]$/.test(expression)) return expression;
  return expression + value;
}

export function toggleSign(expression: string): string {
  if (!expression) return "−";
  if (expression === "−") return "";

  const number = expression.match(/(?:\d+(?:\.\d*)?|\.\d+)$/);
  if (number?.index !== undefined) {
    const prefix = expression.slice(0, number.index);
    const isNegative =
      prefix.endsWith("−") &&
      (prefix.length === 1 || /[+×÷^−(]$/.test(prefix.slice(0, -1)));
    return (isNegative ? prefix.slice(0, -1) : prefix + "−") + number[0];
  }

  if (expression.startsWith("−(") && expression.endsWith(")")) {
    return expression.slice(2, -1);
  }
  return `−(${expression})`;
}
