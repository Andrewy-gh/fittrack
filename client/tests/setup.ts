// Vitest setup file
// This file runs before all tests

import { beforeEach, afterEach, afterAll } from "vitest";
import { installUnitNetworkGuard } from "./unit-network-guard";
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";

// Install before test modules (and therefore application imports) are evaluated.
const network = installUnitNetworkGuard();
afterAll(() => network.assertNoUnexpectedTransport());

// Clean up after each test
afterEach(() => {
  cleanup();
  network.assertNoUnexpectedTransport();
});

// jsdom provides localStorage, but we want to ensure it's clean before each test
beforeEach(() => {
  localStorage.clear();
});
