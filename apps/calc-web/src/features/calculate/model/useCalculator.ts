import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useCallback, useReducer } from "react";
import { calculationMutationOptions } from "../api/calculate";
import {
  calculatorReducer,
  initialCalculatorState,
} from "./calculator.reducer";

export function useCalculator() {
  const [state, dispatch] = useReducer(
    calculatorReducer,
    initialCalculatorState,
  );
  const queryClient = useQueryClient();
  const { mutate, isPending, error, reset } = useMutation({
    ...calculationMutationOptions,
    onSuccess: (result, expression) => {
      dispatch({ type: "evaluate", expression, result });
    },
  });

  const evaluate = useCallback(() => {
    if (
      !state.expression.trim() ||
      state.result !== null ||
      queryClient.isMutating({
        mutationKey: calculationMutationOptions.mutationKey,
      }) > 0
    )
      return;

    mutate(state.expression);
  }, [state.expression, state.result, queryClient, mutate]);

  return {
    state,
    dispatch,
    evaluate,
    isCalculating: isPending,
    error,
    dismissError: reset,
  };
}
