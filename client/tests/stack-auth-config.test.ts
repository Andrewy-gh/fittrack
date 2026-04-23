import { describe, expect, it } from "vitest";
import { resolveStackAuthBootstrapConfig } from "./stack-auth-config";

describe("resolveStackAuthBootstrapConfig", () => {
  it("uses process PROJECT_ID for Playwright server auth bootstrap", () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {
        PROJECT_ID: "doppler-project-id",
        SECRET_SERVER_KEY: "doppler-secret",
      },
      clientEnv: {},
      serverEnv: {},
    });

    expect(config.projectId).toBe("doppler-project-id");
    expect(config.secretServerKey).toBe("doppler-secret");
  });

  it("prefers browser-specific env vars when they are present", () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {
        PROJECT_ID: "generic-project-id",
        VITE_PROJECT_ID: "vite-project-id",
      },
      clientEnv: {},
      serverEnv: {},
    });

    expect(config.projectId).toBe("vite-project-id");
  });

  it("keeps the client env project id ahead of generic process PROJECT_ID", () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {
        PROJECT_ID: "doppler-project-id",
        SECRET_SERVER_KEY: "doppler-secret",
      },
      clientEnv: {
        VITE_PROJECT_ID: "client-env-project-id",
      },
      serverEnv: {},
    });

    expect(config.projectId).toBe("client-env-project-id");
    expect(config.secretServerKey).toBe("doppler-secret");
  });

  it("falls back to env files when process env is missing", () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {},
      clientEnv: {},
      serverEnv: {
        PROJECT_ID: "server-env-project-id",
        SECRET_SERVER_KEY: "server-env-secret",
      },
    });

    expect(config.projectId).toBe("server-env-project-id");
    expect(config.secretServerKey).toBe("server-env-secret");
  });

  it("reads the local E2E auth contract from env", () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {
        E2E_LOCAL_AUTH_ENABLED: "true",
        E2E_LOCAL_AUTH_USER_ID: "fittrack-e2e",
        E2E_LOCAL_AUTH_EMAIL: "fittrack-e2e@example.test",
        E2E_LOCAL_AUTH_DISPLAY_NAME: "FitTrack E2E",
        E2E_LOCAL_AUTH_API_BASE_URL: "http://localhost:9090",
      },
      clientEnv: {},
      serverEnv: {},
    });

    expect(config.localAuthEnabled).toBe(true);
    expect(config.localAuthUserId).toBe("fittrack-e2e");
    expect(config.localAuthEmail).toBe("fittrack-e2e@example.test");
    expect(config.localAuthDisplayName).toBe("FitTrack E2E");
    expect(config.localAuthApiBaseUrl).toBe("http://localhost:9090");
  });

  it("can enable local E2E auth from client env flags", () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {},
      clientEnv: {
        VITE_E2E_LOCAL_AUTH_ENABLED: "1",
      },
      serverEnv: {},
    });

    expect(config.localAuthEnabled).toBe(true);
  });
});
