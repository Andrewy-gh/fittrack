import {
  queryOptions,
  type QueryFunction,
  type QueryFunctionContext,
  type QueryKey,
} from "@tanstack/react-query";

type QuerySource<TData, TQueryKey extends QueryKey> = {
  readonly queryKey: TQueryKey;
  readonly queryFn?: QueryFunction<TData, TQueryKey>;
};

function executeQuerySource<TData, TQueryKey extends QueryKey>(
  source: QuerySource<TData, TQueryKey>,
  context: QueryFunctionContext<QueryKey>,
): TData | Promise<TData> {
  if (!source.queryFn) {
    throw new Error("Query options must provide a query function");
  }

  return source.queryFn({
    client: context.client,
    queryKey: source.queryKey,
    signal: context.signal,
    meta: context.meta,
  });
}

/**
 * Adapt mode-specific React Query options to one application-facing contract.
 *
 * The source query key is preserved at runtime and passed back to its original
 * query function, while the outer option intentionally hides incompatible
 * generated and demo query-key generics from shared callers.
 */
export function normalizeQueryOptions<TData, TQueryKey extends QueryKey>(
  source: QuerySource<TData, TQueryKey>,
) {
  return queryOptions<TData, unknown, TData, QueryKey>({
    queryKey: source.queryKey,
    queryFn: (context) => executeQuerySource(source, context),
  });
}
