import { Server, Socket } from "node:net";
import { syncBuiltinESMExports } from "node:module";

/** Install the ordinary-unit transport boundary before importing application code.
 * Real sockets may reach only loopback servers listening in this worker. Explicit
 * fetch/client mocks remain usable; they do not reach the guarded transports.
 */
export function installUnitNetworkGuard() {
  // Builtins outlive Vitest's per-file module/jsdom isolation. Keep one socket
  // policy and ledger for the worker, including requests made during teardown.
  const key = Symbol.for("fittrack.unit-network-guard");
  type State = { servers: Set<Server>; violations: Error[] };
  // SAFETY: This private symbol extends the worker's builtin constructor; it
  // is written only below with this exact shape, never read from external data.
  const owner = Socket as typeof Socket & { [key]?: State };
  const existing = owner[key];
  const state = existing ?? { servers: new Set<Server>(), violations: [] };
  const { servers, violations } = state;
  const originalListen = Server.prototype.listen;
  const originalConnect = Socket.prototype.connect;
  const originalFetch = globalThis.fetch;

  function deny(): never {
    // Never include URLs, headers or credentials in diagnostics.
    const error = new Error(
      "Unit network isolation: unexpected transport; use an explicit mock or a worker-owned loopback fixture.",
    );
    violations.push(error);
    // Vitest treats this independent unhandled rejection as a run failure,
    // including background attempts after the last hook. A provider catching
    // the returned/thrown error cannot hide the isolation violation.
    void Promise.reject(error);
    throw error;
  }

  function allowed(host: unknown, port: unknown) {
    if (!["127.0.0.1", "::1", "[::1]"].includes(String(host))) return false;
    return [...servers].some((server) => {
      const address = server.address();
      return (
        address !== null &&
        typeof address !== "string" &&
        address.address === String(host).replace(/^\[|\]$/g, "") &&
        address.port === Number(port)
      );
    });
  }

  // SAFETY: Reflect forwards all Node overloads unchanged after admission. The
  // wrappers preserve the original receiver and return type of each overload.
  if (!existing) {
    owner[key] = state;
    Server.prototype.listen = function (this: Server, ...args: unknown[]) {
      servers.add(this);
      this.once("close", () => servers.delete(this));
      // oxlint-disable-next-line anti-slop/no-reflect-apply -- Node interop: preserve every listen overload without normalizing its arguments.
      return Reflect.apply(originalListen, this, args);
    } as typeof originalListen;
    // SAFETY: Node supports options, port/host, and IPC path overloads. Unknown
    // shapes and IPC are denied; permitted TCP calls are forwarded unchanged.
    Socket.prototype.connect = function (this: Socket, ...args: unknown[]) {
      const first = Array.isArray(args[0]) ? args[0][0] : args[0];
      const options =
        typeof first === "object" && first !== null ? first : undefined;
      const host = options && "host" in options ? options.host : args[1];
      const port = options && "port" in options ? options.port : first;
      if (
        (options && "path" in options && options.path != null) ||
        !allowed(host, port)
      )
        deny();
      // oxlint-disable-next-line anti-slop/no-reflect-apply -- Node interop: forward the admitted connect overload unchanged.
      return Reflect.apply(originalConnect, this, args);
    } as typeof originalConnect;
    syncBuiltinESMExports();
  }

  globalThis.fetch = async (input, init) => {
    const url = new URL(input instanceof Request ? input.url : String(input));
    if (
      !allowed(url.hostname, url.port || (url.protocol === "https:" ? 443 : 80))
    )
      deny();
    return originalFetch(input, init);
  };

  return {
    /** Fail even when a client catches and logs the transport rejection. */
    assertNoUnexpectedTransport() {
      if (violations.length)
        throw new Error(
          `Unit network isolation was violated (${violations.length} blocked attempts)`,
          { cause: violations[0] },
        );
    },
  };
}
