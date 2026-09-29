import { describe, it, expect, vi, beforeEach } from "vitest";
import { isApiError, getErrorMessage, showErrorToast } from "./errors";
import { toast } from "sonner";

vi.mock("sonner", () => ({
  toast: {
    error: vi.fn(),
  },
}));

describe("isApiError", () => {
  it("returns false for Error instances without message property structure", () => {
    const error = new Error("Test error");
    expect(isApiError(error)).toBe(true);
  });
});

describe("getErrorMessage", () => {
  it("extracts message from Error instance", () => {
    const error = new Error("Standard error");
    expect(getErrorMessage(error)).toBe("Standard error");
  });

  it("returns string as-is", () => {
    expect(getErrorMessage("Simple error string")).toBe("Simple error string");
  });

  it("returns fallback for empty string", () => {
    expect(getErrorMessage("")).toBe("An unexpected error occurred");
    expect(getErrorMessage("   ")).toBe("An unexpected error occurred");
  });

  it("returns fallback for ApiError with empty message", () => {
    expect(getErrorMessage({ message: "" })).toBe(
      "An unexpected error occurred",
    );
  });
});

describe("showErrorToast", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls toast.error with extracted message", () => {
    showErrorToast({ message: "Test error" });
    expect(toast.error).toHaveBeenCalledWith("Test error");
  });

  it("calls toast.error with fallback for unknown error", () => {
    showErrorToast(null);
    expect(toast.error).toHaveBeenCalledWith("An unexpected error occurred");
  });

  it("uses custom fallback when provided", () => {
    showErrorToast(null, "Custom error message");
    expect(toast.error).toHaveBeenCalledWith("Custom error message");
  });

  it("logs request_id when available", () => {
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    const error = { message: "API error", request_id: "req-123" };

    showErrorToast(error);

    expect(consoleErrorSpy).toHaveBeenCalledWith(
      "[Error] request_id: req-123",
      error,
    );
    consoleErrorSpy.mockRestore();
  });

  it("does not log when request_id is not available", () => {
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    showErrorToast({ message: "API error" });

    expect(consoleErrorSpy).not.toHaveBeenCalled();
    consoleErrorSpy.mockRestore();
  });
});
