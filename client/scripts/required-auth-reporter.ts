import type {
  FullConfig,
  FullResult,
  Reporter,
  Suite,
} from "@playwright/test/reporter";

/** Fail closed on final results, including runtime skips and unexecuted cases. */
export default class RequiredAuthReporter implements Reporter {
  private suite: Suite | undefined;

  onBegin(_config: FullConfig, suite: Suite) {
    this.suite = suite;
  }

  async onEnd(result: FullResult): Promise<{ status: "failed" } | undefined> {
    const tests = this.suite?.allTests() ?? [];
    const passed = tests.filter(
      (test) =>
        test.expectedStatus === "passed" &&
        test.results.at(-1)?.status === "passed" &&
        test.outcome() === "expected",
    ).length;
    const skipped = tests.filter(
      (test) =>
        !test.results.length || test.results.at(-1)?.status === "skipped",
    ).length;
    console.log(
      `Required authenticated suite: ${tests.length} selected, ${passed} passed, ${skipped} skipped/unexecuted.`,
    );
    if (
      !tests.length ||
      passed !== tests.length ||
      result.status !== "passed"
    ) {
      console.error("Required authenticated coverage incomplete or failed.");
      return { status: "failed" };
    }
  }
}
