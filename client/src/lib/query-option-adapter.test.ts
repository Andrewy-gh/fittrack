import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it } from "vitest";

import { normalizeQueryOptions } from "./query-option-adapter";

describe("normalizeQueryOptions", () => {
  it("preserves the source key and delegates with its cancellation context", async () => {
    const sourceKey = ["source", 42] as const;
    const controller = new AbortController();
    const source = {
      queryKey: sourceKey,
      queryFn: ({
        queryKey,
        signal,
      }: {
        queryKey: typeof sourceKey;
        signal: AbortSignal;
      }) => ({
        queryKey,
        signal,
      }),
    };
    const options = normalizeQueryOptions(source);

    if (!options.queryFn) {
      throw new Error("normalized query options must provide a query function");
    }

    const result = await options.queryFn({
      client: new QueryClient(),
      queryKey: options.queryKey,
      signal: controller.signal,
      meta: undefined,
    });

    expect(options.queryKey).toEqual(sourceKey);
    expect(result).toEqual({
      queryKey: sourceKey,
      signal: controller.signal,
    });
  });
});
