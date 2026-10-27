import {
  insertIntoExpression,
  normalizeExpression,
  toggleSign,
} from "./expression.ts";
import type { CalculatorAction, CalculatorState } from "./calculator.types.ts";

export const initialCalculatorState: CalculatorState = {
  expression: "",
  result: null,
  history: [],
  nextId: 1,
};

export function calculatorReducer(
  state: CalculatorState,
  action: CalculatorAction,
): CalculatorState {
  switch (action.type) {
    case "insert": {
      const continuesResult = /^[+×÷−^%]$/.test(action.value);
      const expression =
        state.result === null
          ? state.expression
          : continuesResult
            ? state.result
            : "";
      return {
        ...state,
        expression: insertIntoExpression(expression, action.value),
        result: null,
      };
    }
    case "edit":
      return {
        ...state,
        expression: normalizeExpression(action.expression),
        result: null,
      };
    case "backspace":
      return {
        ...state,
        expression: state.expression.endsWith("√(")
          ? state.expression.slice(0, -2)
          : state.expression.slice(0, -1),
        result: null,
      };
    case "clear":
      return { ...state, expression: "", result: null };
    case "toggle-sign":
      return {
        ...state,
        expression: toggleSign(state.result ?? state.expression),
        result: null,
      };
    case "evaluate": {
      const expression = state.expression.trim();
      if (!expression || state.result !== null) return state;
      return {
        ...state,
        expression: action.result,
        result: action.result,
        history: [
          { id: state.nextId, expression, result: action.result },
          ...state.history,
        ].slice(0, 20),
        nextId: state.nextId + 1,
      };
    }
    case "recall":
      return { ...state, expression: action.expression, result: null };
    case "clear-history":
      return { ...state, history: [] };
  }
}
