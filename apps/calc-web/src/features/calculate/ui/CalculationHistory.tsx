import { ArrowRight, History, Trash2 } from "lucide-react";
import type { Calculation } from "../model/calculator.types";

type CalculationHistoryProps = {
  history: Calculation[];
  onRecall: (expression: string) => void;
  onClear: () => void;
};

export function CalculationHistory({
  history,
  onRecall,
  onClear,
}: CalculationHistoryProps) {
  return (
    <section
      id="calculation-history"
      className="card card-border min-h-64 flex-1 overflow-hidden rounded-2xl border-base-300 bg-base-100 shadow-sm lg:min-h-0"
      aria-labelledby="history-title"
    >
      <header className="flex h-14 shrink-0 items-center justify-between px-4">
        <h2
          id="history-title"
          className="flex items-center gap-2 text-sm font-medium"
        >
          <History
            size={17}
            className="text-base-content/60"
            aria-hidden="true"
          />
          History{" "}
          <span className="badge badge-ghost badge-xs">{history.length}</span>
        </h2>
        <button
          className="btn btn-ghost btn-square btn-sm text-base-content/60"
          type="button"
          aria-label="Clear history"
          title="Clear history"
          disabled={history.length === 0}
          onClick={onClear}
        >
          <Trash2 size={16} aria-hidden="true" />
        </button>
      </header>

      {history.length === 0 ? (
        <div className="flex flex-1 flex-col items-center justify-center px-4 py-8 text-center">
          <div
            className="mb-4 grid size-16 place-items-center rounded-2xl border border-base-300 bg-base-200 text-primary"
            aria-hidden="true"
          >
            <History size={28} aria-hidden="true" />
          </div>
          <p className="text-sm text-base-content/60">No calculations yet.</p>
        </div>
      ) : (
        <ol
          className="menu max-h-80 w-full flex-1 flex-nowrap gap-1 overflow-y-auto px-3 py-0 lg:min-h-0 lg:max-h-none"
          aria-label="Recent calculations"
        >
          {history.map((item) => (
            <li key={item.id}>
              <button
                type="button"
                className="group relative flex min-w-0 flex-col items-end gap-1 rounded-xl p-3"
                onClick={() => onRecall(item.expression)}
                title={`Reuse ${item.expression}`}
                aria-label={`Reuse expression ${item.expression}, result ${item.result}`}
              >
                <span className="w-full truncate text-right text-sm text-base-content/60">
                  {item.expression}
                </span>
                <span className="text-2xl font-medium text-primary">
                  <span className="mr-1 text-sm text-base-content/50">=</span>{" "}
                  {item.result}
                </span>
                <ArrowRight
                  size={15}
                  className="absolute bottom-4 left-3 text-base-content/60 opacity-0 group-hover:opacity-100 group-focus-visible:opacity-100"
                  aria-hidden="true"
                />
              </button>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}
