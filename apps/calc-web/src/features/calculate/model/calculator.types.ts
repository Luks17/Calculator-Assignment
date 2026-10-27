export type Calculation = {
  id: number;
  expression: string;
  result: string;
};

export type CalculatorState = {
  expression: string;
  result: string | null;
  history: Calculation[];
  nextId: number;
};

export type CalculatorAction =
  | { type: "insert"; value: string }
  | { type: "edit"; expression: string }
  | { type: "backspace" }
  | { type: "clear" }
  | { type: "toggle-sign" }
  | { type: "evaluate"; result: string }
  | { type: "recall"; expression: string }
  | { type: "clear-history" };
