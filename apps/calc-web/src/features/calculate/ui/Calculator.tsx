import { Keyboard } from "lucide-react";
import { useEffect, useReducer, useRef, useState } from "react";
import { calculate } from "../api/calculate";
import {
  calculatorReducer,
  initialCalculatorState,
} from "../model/calculator.reducer";
import { keyboardAction } from "../model/keyboard";
import { CalculationHistory } from "./CalculationHistory";
import { CalculatorDisplay } from "./CalculatorDisplay";
import { CalculatorKeypad } from "./CalculatorKeypad";

export function Calculator() {
  const [state, dispatch] = useReducer(
    calculatorReducer,
    initialCalculatorState,
  );
  const [showHistory, setShowHistory] = useState(true);
  const inputRef = useRef<HTMLInputElement>(null);

  const evaluate = () => {
    if (state.expression.trim() && state.result === null) {
      dispatch({ type: "evaluate", result: calculate(state.expression) });
    }
  };

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.ctrlKey || event.metaKey || event.altKey || event.isComposing)
        return;
      const target = event.target;
      const isExpressionInput = target === inputRef.current;
      const isEditable =
        target instanceof HTMLElement &&
        (target.isContentEditable ||
          ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName));
      if (isEditable && !isExpressionInput) return;

      // Preserve native text editing and native Enter activation on buttons.
      if (event.key === "Enter" || event.key === "=") {
        if (
          event.key === "Enter" &&
          target instanceof HTMLElement &&
          target.closest("button, a")
        )
          return;
        event.preventDefault();
        if (state.expression.trim() && state.result === null) {
          dispatch({ type: "evaluate", result: calculate(state.expression) });
        }
        return;
      }

      if (
        isExpressionInput &&
        event.key !== "Escape" &&
        event.key.toLowerCase() !== "r"
      )
        return;
      const action = keyboardAction(event.key);
      if (action) {
        event.preventDefault();
        dispatch(action);
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [state.expression, state.result]);

  return (
    <div
      className={`grid items-start gap-4 sm:gap-6 lg:min-h-0 lg:flex-1 lg:grid-rows-1 lg:items-stretch ${showHistory ? "md:grid-cols-5 lg:grid-cols-4" : ""}`}
    >
      <section
        className={`card card-border min-w-0 overflow-hidden rounded-2xl border-base-300 bg-base-100 shadow-lg lg:min-h-0 ${showHistory ? "md:col-span-3" : ""}`}
        aria-label="Calculator"
      >
        <CalculatorDisplay
          expression={state.expression}
          result={state.result}
          inputRef={inputRef}
          onEdit={(expression) => dispatch({ type: "edit", expression })}
          showHistory={showHistory}
          onToggleHistory={() => setShowHistory((value) => !value)}
        />
        <CalculatorKeypad onAction={dispatch} onCalculate={evaluate} />
      </section>

      {showHistory && (
        <aside
          className="flex min-w-0 flex-col gap-4 md:col-span-2 md:self-stretch lg:col-span-1 lg:min-h-0"
          aria-label="Calculation tools"
        >
          <CalculationHistory
            history={state.history}
            onRecall={(expression) => {
              dispatch({ type: "recall", expression });
              inputRef.current?.focus();
            }}
            onClear={() => dispatch({ type: "clear-history" })}
          />
          <section
            className="card card-border shrink-0 rounded-2xl border-base-300 bg-base-100/50"
            aria-labelledby="keyboard-title"
          >
            <div className="card-body gap-3 p-5">
              <div className="flex items-center gap-2">
                <Keyboard
                  size={19}
                  className="shrink-0 text-primary"
                  aria-hidden="true"
                />
                <h2 id="keyboard-title" className="text-sm font-medium">
                  Keyboard shortcuts
                </h2>
              </div>
              <div className="mt-1 flex flex-wrap items-center gap-4 text-xs text-base-content/60">
                <span className="flex items-center gap-2">
                  <kbd className="kbd kbd-sm">enter</kbd> calculate
                </span>
                <span className="flex items-center gap-2">
                  <kbd className="kbd kbd-sm">esc</kbd> clear
                </span>
              </div>
            </div>
          </section>
        </aside>
      )}
    </div>
  );
}
