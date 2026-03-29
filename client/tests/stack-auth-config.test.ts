import { describe, expect, it } from 'vitest';
import { resolveStackAuthBootstrapConfig } from './stack-auth-config';

describe('resolveStackAuthBootstrapConfig', () => {
  it('uses process PROJECT_ID for Playwright server auth bootstrap', () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {
        PROJECT_ID: 'doppler-project-id',
        SECRET_SERVER_KEY: 'doppler-secret',
      },
      clientEnv: {},
      serverEnv: {},
    });

    expect(config.projectId).toBe('doppler-project-id');
    expect(config.secretServerKey).toBe('doppler-secret');
  });

  it('prefers browser-specific env vars when they are present', () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {
        PROJECT_ID: 'generic-project-id',
        VITE_PROJECT_ID: 'vite-project-id',
      },
      clientEnv: {},
      serverEnv: {},
    });

    expect(config.projectId).toBe('vite-project-id');
  });

  it('falls back to env files when process env is missing', () => {
    const config = resolveStackAuthBootstrapConfig({
      processEnv: {},
      clientEnv: {},
      serverEnv: {
        PROJECT_ID: 'server-env-project-id',
        SECRET_SERVER_KEY: 'server-env-secret',
      },
    });

    expect(config.projectId).toBe('server-env-project-id');
    expect(config.secretServerKey).toBe('server-env-secret');
  });
});
