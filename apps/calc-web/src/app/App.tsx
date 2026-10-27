import { CalculatorPage } from "../pages/calculator";
import { QueryProvider } from "./providers/QueryProvider";

export function App() {
  return (
    <QueryProvider>
      <CalculatorPage />
    </QueryProvider>
  );
}
