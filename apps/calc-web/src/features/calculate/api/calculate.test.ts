import { MutationObserver, QueryClient } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { calculate, calculationMutationOptions } from "./calculate.ts";

const fetchMock = vi.fn<typeof fetch>();

beforeEach(() => {
  fetchMock.mockReset();
  vi.stubGlobal("fetch", fetchMock);
  vi.stubEnv("VITE_API_BASE_URL", "");
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

test.each([
  [" 2+3 ", "2+3", "5"],
  ["5−7", "5-7", "-2"],
  ["4×6", "4*6", "24"],
  ["8÷2", "8/2", "4"],
  ["2^8", "2^8", "256"],
  ["√(√(81))", "sqrt(sqrt(81))", "3"],
  ["25%", "25%", "0.25"],
  ["(2+3)×4", "(2+3)*4", "20"],
  ["0.1+0.2", "0.1+0.2", "0.3"],
  [
    "999999999999999999999999999999999999999.000000000001+1",
    "999999999999999999999999999999999999999.000000000001+1",
    "1000000000000000000000000000000000000000.00000000000100",
  ],
])(
  "submits %s using API notation and returns the exact result",
  async (expression, expected, result) => {
    fetchMock.mockResolvedValueOnce(Response.json({ result }));

    await expect(calculate(expression)).resolves.toBe(result);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, options] = fetchMock.mock.calls[0];
    expect(url).toBe("http://localhost:8080/api/v1/calculations");
    expect(options?.method).toBe("POST");
    expect(new Headers(options?.headers).get("Content-Type")).toBe(
      "application/json",
    );
    expect(JSON.parse(options?.body as string)).toEqual({
      expression: expected,
    });
  },
);

test("supports configuring the API base URL", async () => {
  vi.stubEnv("VITE_API_BASE_URL", "https://calculator.example");
  fetchMock.mockResolvedValueOnce(Response.json({ result: "5" }));

  await expect(calculate("2+3")).resolves.toBe("5");
  expect(fetchMock.mock.calls[0][0]).toBe(
    "https://calculator.example/api/v1/calculations",
  );
});

test.each([
  [400, "INVALID_REQUEST", "Expression must not be empty"],
  [422, "INVALID_EXPRESSION", "Unexpected token at position 5"],
  [422, "DIVISION_BY_ZERO", "Division by zero is not allowed"],
  [
    422,
    "INVALID_DOMAIN",
    "Cannot calculate the square root of a negative number",
  ],
  [500, "INTERNAL_ERROR", "An unexpected error occurred"],
])(
  "preserves the API error message for HTTP %s / %s",
  async (status, code, message) => {
    fetchMock.mockResolvedValueOnce(
      Response.json({ code, message }, { status }),
    );

    await expect(calculate("1÷0")).rejects.toThrow(message);
  },
);

test("provides a readable fallback for network failures", async () => {
  fetchMock.mockRejectedValueOnce(new TypeError("Failed to fetch"));

  await expect(calculate("2+3")).rejects.toThrow(
    "Unable to calculate. Please try again.",
  );
});

test.each([
  new Response("Bad Gateway", { status: 502 }),
  Response.json({ message: "" }, { status: 500 }),
  Response.json({ message: 42 }, { status: 500 }),
])(
  "provides a readable fallback for unstructured errors (%#)",
  async (response) => {
    fetchMock.mockResolvedValueOnce(response);

    await expect(calculate("2+3")).rejects.toThrow(
      "Unable to calculate. Please try again.",
    );
  },
);

test.each([{}, { result: 5 }, null])(
  "rejects invalid successful responses (%#)",
  async (body) => {
    fetchMock.mockResolvedValueOnce(Response.json(body));

    await expect(calculate("2+3")).rejects.toThrow(
      "The calculator service returned an invalid result.",
    );
  },
);

test("the calculation mutation tracks pending and successful requests", async () => {
  let resolveResponse!: (response: Response) => void;
  const response = new Promise<Response>((resolve) => {
    resolveResponse = resolve;
  });
  fetchMock.mockReturnValueOnce(response);
  const queryClient = new QueryClient();
  const mutation = new MutationObserver(
    queryClient,
    calculationMutationOptions,
  );

  const pending = mutation.mutate("2+3");
  expect(mutation.getCurrentResult().isPending).toBe(true);
  expect(
    queryClient.isMutating({
      mutationKey: calculationMutationOptions.mutationKey,
    }),
  ).toBe(1);

  resolveResponse(Response.json({ result: "5" }));
  await expect(pending).resolves.toBe("5");
  expect(mutation.getCurrentResult()).toMatchObject({
    isPending: false,
    isSuccess: true,
    data: "5",
    error: null,
    variables: "2+3",
  });
  expect(
    queryClient.isMutating({
      mutationKey: calculationMutationOptions.mutationKey,
    }),
  ).toBe(0);
  queryClient.clear();
});

test("the calculation mutation exposes errors without retries and resets for dismissal", async () => {
  fetchMock.mockResolvedValueOnce(
    Response.json(
      {
        code: "DIVISION_BY_ZERO",
        message: "Division by zero is not allowed",
      },
      { status: 422 },
    ),
  );
  const queryClient = new QueryClient();
  const mutation = new MutationObserver(
    queryClient,
    calculationMutationOptions,
  );

  await expect(mutation.mutate("1÷0")).rejects.toThrow(
    "Division by zero is not allowed",
  );
  expect(mutation.getCurrentResult()).toMatchObject({
    isPending: false,
    isError: true,
    error: new Error("Division by zero is not allowed"),
  });
  expect(fetchMock).toHaveBeenCalledTimes(1);

  mutation.reset();
  expect(mutation.getCurrentResult()).toMatchObject({
    isIdle: true,
    error: null,
  });
  queryClient.clear();
});
