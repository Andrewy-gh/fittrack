import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { client } from "@/client/client.gen";
import { MuscleContributions } from "./muscle-contributions";
const catalog = [
  {
    id: "Triceps_Pushdown",
    name: "Triceps Pushdown",
    equipment: "cable",
    primaryMuscles: ["triceps"],
    secondaryMuscles: [],
  },
  {
    id: "Pushups",
    name: "Pushups",
    equipment: "body only",
    primaryMuscles: ["chest"],
    secondaryMuscles: ["triceps"],
  },
];
let catalogId = "",
  failSave = false,
  failDetail = false,
  failSummary = false,
  partial = false;
function evidence() {
  const reviewed = catalogId === "Triceps_Pushdown";
  return {
    period: { partial },
    workingSets: 3,
    invalidWorkingSets: 1,
    unclassifiedSets: catalogId ? 0 : 3,
    unreviewedSets: catalogId === "Pushups" ? 3 : 0,
    mappingVersion: "v1",
    catalogVersion: "c1",
    muscles: [
      {
        muscle: "triceps",
        directSets: reviewed ? 3 : 0,
        indirectSets: 0,
        unresolvedSets: reviewed ? 0 : 3,
        invalidMappedSets: reviewed ? 1 : 0,
        invalidUnresolvedSets: reviewed ? 0 : 1,
      },
    ],
    exercises: [
      {
        exerciseId: 42,
        name: "My custom press",
        catalogId,
        workingSets: 3,
        invalidWorkingSets: 1,
        roles: reviewed ? { triceps: "direct" } : {},
      },
    ],
  };
}
beforeEach(() => {
  catalogId = "";
  failSave = false;
  failDetail = false;
  failSummary = false;
  partial = false;
  vi.spyOn(client, "get").mockImplementation(async (options) => {
    if (options.url === "/analytics/muscle-contributions") {
      if (failSummary) throw new Error("offline");
      return { data: evidence() } as never;
    }
    if (options.url === "/exercise-catalog") return { data: catalog } as never;
    if (options.url === "/exercises/{id}") {
      if (failDetail) throw new Error("offline");
      return {
        data: {
          exercise: {
            id: 42,
            name: "My custom press",
            catalog: catalog.find((entry) => entry.id === catalogId),
          },
          sets: [],
        },
      } as never;
    }
    throw new Error(`Unexpected GET ${options.url}`);
  });
  vi.spyOn(client, "put").mockImplementation(async (options) => {
    if (failSave) throw new Error("offline");
    catalogId = (options.body as { catalog_id: string }).catalog_id;
    return { data: {} } as never;
  });
});
afterEach(() => vi.restoreAllMocks());
function mount() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  });
  render(
    <QueryClientProvider client={queryClient}>
      <MuscleContributions
        userId="owner"
        startDate="2026-09-14"
        timezone="America/New_York"
      />
    </QueryClientProvider>,
  );
}
async function openPanel() {
  mount();
  fireEvent.click(await screen.findByText("Exercise matches"));
  fireEvent.click(
    screen.getByRole("button", { name: "Classify My custom press" }),
  );
  await screen.findByRole("button", { name: "Classify exercise" });
}
async function choose(name: string, changing = false) {
  fireEvent.click(
    screen.getByRole("button", {
      name: changing ? "Change classification" : "Classify exercise",
    }),
  );
  fireEvent.click(await screen.findByRole("button", { name }));
  await waitFor(() =>
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
  );
}
it("classifies, changes to an unsupported match and clears while refreshing the selected week", async () => {
  await openPanel();
  await choose("Triceps Pushdown");
  expect(await screen.findByText("3 direct · 0 indirect")).toBeVisible();
  expect(
    screen.getByRole("region", { name: "Classification for My custom press" }),
  ).toBeVisible();
  expect(
    screen.queryByRole("button", { name: "Classify My custom press" }),
  ).not.toBeInTheDocument();
  await waitFor(() =>
    expect(
      screen.getByRole("button", { name: "Change classification" }),
    ).toHaveFocus(),
  );
  await choose("Pushups", true);
  expect(await screen.findByText("No counts available")).toBeVisible();
  expect(
    screen.getByText("Matched · muscle use not reviewed yet."),
  ).toBeVisible();
  expect(
    screen.getByText("3 sets not counted · muscle use not reviewed"),
  ).toBeVisible();
  fireEvent.click(screen.getByRole("button", { name: "Clear" }));
  expect(await screen.findByText("Unclassified")).toBeVisible();
  expect(
    await screen.findByRole("button", { name: "Classify My custom press" }),
  ).toBeVisible();
  expect(client.put).toHaveBeenCalledTimes(3);
  for (const [options] of vi.mocked(client.put).mock.calls)
    expect(options.path).toEqual({ id: 42 });
  const requests = vi
    .mocked(client.get)
    .mock.calls.filter(
      ([options]) => options.url === "/analytics/muscle-contributions",
    );
  expect(requests).toHaveLength(4);
  for (const [options] of requests)
    expect(options.query).toEqual({
      startDate: "2026-09-14",
      timezone: "America/New_York",
    });
  fireEvent.click(screen.getByRole("button", { name: "Done" }));
  expect(screen.getByRole("heading", { name: "Muscle sets" })).toHaveFocus();
});
it("keeps evidence on save error and supports no match, cancel and retry", async () => {
  await openPanel();
  failSave = true;
  fireEvent.click(screen.getByRole("button", { name: "Classify exercise" }));
  fireEvent.change(await screen.findByLabelText("Search catalog"), {
    target: { value: "no match" },
  });
  expect(screen.getByText(/No catalog matches/)).toBeVisible();
  expect(client.put).not.toHaveBeenCalled();
  fireEvent.change(screen.getByLabelText("Search catalog"), {
    target: { value: "" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Triceps Pushdown" }));
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Could not save classification",
  );
  expect(catalogId).toBe("");
  fireEvent.click(
    within(screen.getByRole("dialog")).getByRole("button", { name: "Close" }),
  );
  await waitFor(() =>
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
  );
  expect(
    screen.getByText("3 sets not counted · muscle use not reviewed"),
  ).toBeVisible();
  failSave = false;
  await choose("Triceps Pushdown");
  expect(await screen.findByText("3 direct · 0 indirect")).toBeVisible();
});
it("shows classification loading without hiding evidence", async () => {
  const get = vi.mocked(client.get);
  const original = get.getMockImplementation()!;
  let release!: () => void;
  get.mockImplementation(async (options) => {
    if (options.url === "/exercises/{id}")
      await new Promise<void>((resolve) => {
        release = resolve;
      });
    return original(options);
  });
  mount();
  fireEvent.click(await screen.findByText("Exercise matches"));
  fireEvent.click(
    screen.getByRole("button", { name: "Classify My custom press" }),
  );
  expect(await screen.findByRole("status")).toHaveTextContent(
    "Loading classification",
  );
  expect(
    screen.getByText("3 sets not counted · muscle use not reviewed"),
  ).toBeVisible();
  release();
  expect(
    await screen.findByRole("button", { name: "Classify exercise" }),
  ).toBeVisible();
});
it("retries a failed classification detail request", async () => {
  failDetail = true;
  mount();
  fireEvent.click(await screen.findByText("Exercise matches"));
  fireEvent.click(
    screen.getByRole("button", { name: "Classify My custom press" }),
  );
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Could not load classification",
  );
  failDetail = false;
  fireEvent.click(screen.getByRole("button", { name: "Retry classification" }));
  expect(
    await screen.findByRole("button", { name: "Classify exercise" }),
  ).toBeVisible();
});
it("keeps partial evidence neutral and exposes research sources", async () => {
  partial = true;
  mount();
  expect(await screen.findByText("This week so far")).toBeVisible();
  fireEvent.click(screen.getByText("About these counts"));
  expect(screen.getByRole("link", { name: "ACSM guidance" })).toHaveAttribute(
    "href",
    "https://acsm.org/resistance-training-guidelines-update-2026/",
  );
  expect(
    screen.getByText(
      /different counting methods mean these numbers are not a personal target/,
    ),
  ).toBeVisible();
  expect(
    screen.queryByText(/Below reference|Reference met/),
  ).not.toBeInTheDocument();
});
it("keeps classification available if summary refresh fails after save", async () => {
  await openPanel();
  failSummary = true;
  await choose("Triceps Pushdown");
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Could not load sets",
  );
  expect(
    screen.getByRole("button", { name: "Change classification" }),
  ).toBeVisible();
  failSummary = false;
  fireEvent.click(screen.getByRole("button", { name: "Retry" }));
  expect(await screen.findByText("3 direct · 0 indirect")).toBeVisible();
});

it("focuses retry after a saved match fails to refresh, without showing stale actions", async () => {
  await openPanel();
  failDetail = true;
  await choose("Triceps Pushdown");
  expect(catalogId).toBe("Triceps_Pushdown");
  expect(await screen.findByText("3 direct · 0 indirect")).toBeVisible();
  const retry = screen.getByRole("button", { name: "Retry classification" });
  await waitFor(() => expect(retry).toHaveFocus());
  expect(screen.queryByText("Unclassified")).not.toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Classify exercise" }),
  ).not.toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Change classification" }),
  ).not.toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Clear" }),
  ).not.toBeInTheDocument();
  failDetail = false;
  fireEvent.click(retry);
  expect(
    await screen.findByRole("button", { name: "Change classification" }),
  ).toBeVisible();
  expect(screen.getByText("Triceps Pushdown")).toBeVisible();
  await waitFor(() =>
    expect(
      screen.getByRole("heading", { name: "My custom press" }),
    ).toHaveFocus(),
  );
  expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  const requests = vi
    .mocked(client.get)
    .mock.calls.filter(
      ([options]) => options.url === "/analytics/muscle-contributions",
    );
  for (const [options] of requests)
    expect(options.query).toEqual({
      startDate: "2026-09-14",
      timezone: "America/New_York",
    });
});
