import { beforeEach, describe, expect, it, vi } from "vitest";

type RequestInterceptor = (request: Request) => Promise<Request>;

const mocks = vi.hoisted(() => ({
  applyLocalDevAuthHeader: vi.fn((headers: Headers) => headers),
  getUser: vi.fn(),
  requestInterceptors: [] as RequestInterceptor[],
  setConfig: vi.fn(),
}));

vi.mock("@/client/client.gen", () => ({
  client: {
    setConfig: mocks.setConfig,
    interceptors: {
      request: {
        use: vi.fn((fn: RequestInterceptor) => {
          mocks.requestInterceptors.push(fn);
        }),
      },
      response: {
        use: vi.fn(),
      },
    },
  },
}));

vi.mock("@/lib/local-dev-auth", () => ({
  applyLocalDevAuthHeader: mocks.applyLocalDevAuthHeader,
}));

vi.mock("@/stack", () => ({
  stackClientApp: {
    getUser: mocks.getUser,
  },
}));

async function loadRequestInterceptor(): Promise<RequestInterceptor> {
  mocks.requestInterceptors.length = 0;
  await import("./client-config");

  const interceptor = mocks.requestInterceptors.at(0);
  if (!interceptor) {
    throw new Error("request interceptor was not registered");
  }
  return interceptor;
}

describe("client auth request interceptor", () => {
  beforeEach(() => {
    vi.resetModules();
    mocks.applyLocalDevAuthHeader.mockClear();
    mocks.getUser.mockReset();
    mocks.setConfig.mockClear();
  });

  it("configures the generated client with the local test API base URL", async () => {
    await loadRequestInterceptor();

    expect(mocks.setConfig).toHaveBeenCalledWith({
      baseUrl: "http://localhost/api",
    });
  });

  it("adds the Stack access token when the session is available", async () => {
    mocks.getUser.mockResolvedValue({
      getAuthJson: vi.fn().mockResolvedValue({ accessToken: "token-123" }),
    });
    const interceptor = await loadRequestInterceptor();

    const request = await interceptor(new Request("http://localhost/api"));

    expect(request.headers.get("x-stack-access-token")).toBe("token-123");
    expect(mocks.applyLocalDevAuthHeader).not.toHaveBeenCalled();
  });

  it("throws a friendly retry message when Stack cannot verify the session", async () => {
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    mocks.getUser.mockRejectedValue(new TypeError("Load failed"));
    const { AUTH_SESSION_UNAVAILABLE_MESSAGE } =
      await import("./client-config");
    const interceptor = mocks.requestInterceptors.at(0);
    if (!interceptor) {
      throw new Error("request interceptor was not registered");
    }

    await expect(
      interceptor(new Request("http://localhost/api/workouts")),
    ).rejects.toEqual({
      message: AUTH_SESSION_UNAVAILABLE_MESSAGE,
    });
    expect(mocks.applyLocalDevAuthHeader).not.toHaveBeenCalled();

    consoleError.mockRestore();
  });
});
