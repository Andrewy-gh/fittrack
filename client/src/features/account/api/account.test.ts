import { beforeEach, describe, expect, it, vi } from "vitest";

const { getUser } = vi.hoisted(() => ({
  getUser: vi.fn(),
}));

const { applyLocalDevAuthHeader } = vi.hoisted(() => ({
  applyLocalDevAuthHeader: vi.fn((headers: Headers) => {
    headers.set("x-fittrack-dev-e2e-user", "local-e2e-user");
    return headers;
  }),
}));

vi.mock("@/stack", () => ({
  stackClientApp: {
    getUser,
  },
}));

vi.mock("@/lib/local-dev-auth", () => ({
  applyLocalDevAuthHeader,
}));

import { deleteAccount } from "@/features/account/api/account";

function latestRequest(): Request {
  return vi.mocked(fetch).mock.calls.at(-1)?.[0] as Request;
}

describe("account api wrapper", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    getUser.mockReset();
    applyLocalDevAuthHeader.mockClear();
    getUser.mockResolvedValue({
      getAuthJson: vi.fn().mockResolvedValue({ accessToken: "token-123" }),
    });
  });

  it("deletes the current FitTrack account", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 204 }),
    );

    await deleteAccount();

    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    expect(latestRequest().url).toContain("/api/account");
    expect(latestRequest().method).toBe("DELETE");
  });
});
