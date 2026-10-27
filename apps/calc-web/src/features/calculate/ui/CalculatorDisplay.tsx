import { History } from "lucide-react";
import type { RefObject } from "react";

type CalculatorDisplayProps = {
  expression: string;
  result: string | null;
  inputRef: RefObject<HTMLInputElement | null>;
  onEdit: (expression: string) => void;
  showHistory: boolean;
  onToggleHistory: () => void;
};

export function CalculatorDisplay({
  expression,
  result,
  inputRef,
  onEdit,
  showHistory,
  onToggleHistory,
}: CalculatorDisplayProps) {
  return (
    <div className="min-w-0 shrink-0 p-5 sm:px-6 lg:p-6 xl:p-8">
      <div className="flex items-center justify-between">
        <label
          className="text-xs tracking-widest text-base-content/60"
          htmlFor="calculator-expression"
        >
          EXPRESSION
        </label>
        <button
          className={`btn btn-ghost btn-square btn-sm ${showHistory ? "btn-active text-primary" : "text-base-content/60"}`}
          type="button"
          aria-label={showHistory ? "Hide history" : "Show history"}
          aria-expanded={showHistory}
          aria-controls="calculation-history"
          title={showHistory ? "Hide history" : "Show history"}
          onClick={onToggleHistory}
        >
          <History size={18} aria-hidden="true" />
        </button>
      </div>
      <input
        ref={inputRef}
        id="calculator-expression"
        className="input input-ghost mt-3 h-20 w-full min-w-0 rounded-none border-0 border-b-2 border-transparent bg-transparent px-0 text-right text-4xl font-light tracking-tight tabular-nums shadow-none placeholder:text-base-content/50 focus:border-primary focus:bg-transparent focus:shadow-none focus:outline-none sm:text-5xl lg:h-24 lg:text-6xl"
        type="text"
        inputMode="decimal"
        autoComplete="off"
        spellCheck={false}
        aria-describedby="expression-hint"
        placeholder="0"
        value={expression}
        onChange={(event) => onEdit(event.target.value)}
      />
      <span id="expression-hint" className="sr-only">
        Type an expression or use the keypad. Press Enter to calculate and
        Escape to clear.
      </span>
      <span className="sr-only" role="status" aria-atomic="true">
        {result !== null ? `Result: ${result}` : ""}
      </span>
    </div>
  );
}
