import { mutationOptions } from "@tanstack/react-query";
import { calculate as requestCalculation } from "../../../shared/api";

/** Translate display notation only; parsing and arithmetic belong to the API. */
export async function calculate(expression: string): Promise<string> {
  const { data, error, response } = await requestCalculation({
    ...(import.meta.env.VITE_API_BASE_URL
      ? { baseUrl: import.meta.env.VITE_API_BASE_URL }
      : {}),
    body: {
      expression: expression
        .trim()
        .replaceAll("−", "-")
        .replaceAll("×", "*")
        .replaceAll("÷", "/")
        .replaceAll("√", "sqrt"),
    },
  });

  if (error) {
    const message =
      response && typeof error.message === "string" && error.message.trim()
        ? error.message
        : "Unable to calculate. Please try again.";
    throw new Error(message);
  }

  if (typeof data?.result !== "string") {
    throw new Error("The calculator service returned an invalid result.");
  }

  return data.result;
}

export const calculationMutationOptions = mutationOptions({
  mutationKey: ["calculate"],
  mutationFn: calculate,
  retry: false,
});
