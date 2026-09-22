import { QueryClientProvider, useQuery } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { client } from "@/client/client.gen";
import { queryClient } from "@/lib/api/api";
import {
  useDeleteExerciseMutation,
  useRenameExerciseMutation,
  useUpdateExerciseMutation,
} from "./exercises";
let summary = "Old exercise summary";
function Harness({ operation }: { operation: "update" | "rename" | "delete" }) {
  const update = useUpdateExerciseMutation();
  const rename = useRenameExerciseMutation(false);
  const remove = useDeleteExerciseMutation();
  const query = useQuery({
    queryKey: ["muscle-contributions", "owner", "2026-09-14", "UTC"],
    queryFn: async () => summary,
    staleTime: 60_000,
  });
  return (
    <>
      <p>{query.data}</p>
      <button
        onClick={() => {
          if (operation === "delete") remove.mutate({ path: { id: 1 } });
          else
            (operation === "rename" ? rename : update).mutate({
              path: { id: 1 },
              body: { name: "New name" },
            });
        }}
      >
        Change exercise
      </button>
    </>
  );
}
beforeEach(() => {
  queryClient.clear();
  summary = "Old exercise summary";
  vi.spyOn(client, "patch").mockImplementation(async () => {
    summary = "Fresh exercise summary";
    return { data: {} } as never;
  });
  vi.spyOn(client, "delete").mockImplementation(async () => {
    summary = "Fresh exercise summary";
    return { data: {} } as never;
  });
});
afterEach(() => {
  queryClient.clear();
  vi.restoreAllMocks();
});
for (const operation of ["update", "rename", "delete"] as const)
  it(`refreshes muscle contributions after exercise ${operation}`, async () => {
    render(
      <QueryClientProvider client={queryClient}>
        <Harness operation={operation} />
      </QueryClientProvider>,
    );
    expect(await screen.findByText("Old exercise summary")).toBeVisible();
    fireEvent.click(screen.getByRole("button", { name: "Change exercise" }));
    expect(await screen.findByText("Fresh exercise summary")).toBeVisible();
  });
