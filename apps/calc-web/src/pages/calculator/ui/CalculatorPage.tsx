import { Calculator as CalculatorIcon } from "lucide-react";
import { Calculator } from "../../../features/calculate";
import { ThemeToggle } from "../../../shared/theme";

export function CalculatorPage() {
  return (
    <div className="flex min-h-svh flex-col bg-base-200 font-sans text-base-content antialiased lg:h-svh lg:min-h-[46rem]">
      <a
        className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-10 focus:rounded-box focus:bg-base-100 focus:p-4"
        href="#calculator-expression"
      >
        Skip to calculator
      </a>

      <header className="mx-auto flex h-20 w-full max-w-[1600px] shrink-0 items-center justify-between px-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-3 text-xl font-semibold tracking-tight">
          <span className="grid size-10 place-items-center rounded-xl bg-primary text-primary-content shadow-sm">
            <CalculatorIcon size={22} aria-hidden="true" />
          </span>
          <h1>Calculator</h1>
        </div>
        <ThemeToggle />
      </header>

      <main className="mx-auto w-full max-w-[1600px] px-4 pt-2 pb-4 sm:px-6 sm:pb-6 lg:flex lg:min-h-0 lg:flex-1 lg:flex-col lg:px-8 lg:pt-4 lg:pb-8">
        <Calculator />
      </main>
    </div>
  );
}
