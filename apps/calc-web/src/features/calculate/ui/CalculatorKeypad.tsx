import { Delete } from "lucide-react";
import type { ReactNode } from "react";
import type { CalculatorAction } from "../model/calculator.types";

type Key = {
  label: ReactNode;
  name: string;
  value?: string;
  action?: "clear" | "backspace" | "toggle-sign" | "evaluate";
  kind?: "function" | "operator" | "clear" | "equals";
  shortcut?: string;
};

const keys: Key[] = [
  {
    label: "AC",
    name: "Clear expression",
    action: "clear",
    kind: "clear",
    shortcut: "Escape",
  },
  {
    label: "(",
    name: "Open parenthesis",
    value: "(",
    kind: "function",
    shortcut: "(",
  },
  {
    label: ")",
    name: "Close parenthesis",
    value: ")",
    kind: "function",
    shortcut: ")",
  },
  {
    label: <Delete size={24} className="lg:size-8" aria-hidden="true" />,
    name: "Backspace",
    action: "backspace",
    kind: "function",
    shortcut: "Backspace",
  },
  {
    label: "√",
    name: "Square root",
    value: "√(",
    kind: "function",
    shortcut: "R",
  },
  {
    label: (
      <span>
        x<sup className="relative -top-1 text-xs">y</sup>
      </span>
    ),
    name: "Exponentiation",
    value: "^",
    kind: "function",
    shortcut: "^",
  },
  {
    label: "%",
    name: "Percentage",
    value: "%",
    kind: "function",
    shortcut: "%",
  },
  { label: "÷", name: "Divide", value: "÷", kind: "operator", shortcut: "/" },
  { label: "7", name: "7", value: "7" },
  { label: "8", name: "8", value: "8" },
  { label: "9", name: "9", value: "9" },
  { label: "×", name: "Multiply", value: "×", kind: "operator", shortcut: "*" },
  { label: "4", name: "4", value: "4" },
  { label: "5", name: "5", value: "5" },
  { label: "6", name: "6", value: "6" },
  { label: "−", name: "Subtract", value: "−", kind: "operator", shortcut: "-" },
  { label: "1", name: "1", value: "1" },
  { label: "2", name: "2", value: "2" },
  { label: "3", name: "3", value: "3" },
  { label: "+", name: "Add", value: "+", kind: "operator", shortcut: "+" },
  { label: "±", name: "Change sign", action: "toggle-sign", kind: "function" },
  { label: "0", name: "0", value: "0" },
  { label: ".", name: "Decimal point", value: "." },
  {
    label: "=",
    name: "Calculate",
    action: "evaluate",
    kind: "equals",
    shortcut: "Enter",
  },
];

const keyAppearance = {
  number: "bg-base-100 hover:bg-base-300",
  function: "btn-ghost bg-base-300/50 text-base-content/70",
  operator: "btn-soft btn-primary",
  clear: "btn-soft btn-warning",
  equals: "btn-primary",
};

type CalculatorKeypadProps = {
  onAction: (action: CalculatorAction) => void;
  onCalculate: () => void;
  isCalculating: boolean;
};

export function CalculatorKeypad({
  onAction,
  onCalculate,
  isCalculating,
}: CalculatorKeypadProps) {
  return (
    <div
      className="grid grid-cols-4 gap-2 border-t border-base-300 bg-base-200/50 p-3 lg:min-h-0 lg:flex-1 lg:grid-rows-6 lg:gap-3 lg:p-4"
      role="group"
      aria-label="Calculator keypad"
    >
      {keys.map((key) => (
        <button
          key={key.name}
          type="button"
          className={`btn btn-lg w-full rounded-xl px-0 text-2xl font-normal motion-reduce:transition-none lg:h-full lg:min-h-12 lg:text-3xl ${keyAppearance[key.kind ?? "number"]}`}
          aria-label={key.name}
          aria-busy={key.action === "evaluate" && isCalculating}
          disabled={key.action === "evaluate" && isCalculating}
          title={key.shortcut ? `${key.name} (${key.shortcut})` : key.name}
          onPointerDown={(event) => event.preventDefault()}
          onClick={() => {
            if (key.action === "evaluate") onCalculate();
            else if (key.action) onAction({ type: key.action });
            else if (key.value !== undefined)
              onAction({ type: "insert", value: key.value });
          }}
        >
          {key.action === "evaluate" && isCalculating ? (
            <span
              className="loading loading-spinner loading-sm"
              aria-hidden="true"
            />
          ) : (
            key.label
          )}
        </button>
      ))}
    </div>
  );
}
