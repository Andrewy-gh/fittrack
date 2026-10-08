import { createServer, preview } from "vite";
import { readPort } from "./local-server-config";

const command = process.argv[2];
if (command === "start") {
  const server = await createServer({
    server: { port: readPort("FRONTEND_PORT", 3000), strictPort: true },
  });
  await server.listen();
  server.printUrls();
} else if (command === "test") {
  const server = await preview({
    preview: {
      host: "127.0.0.1",
      port: readPort("E2E_PORT", 5173),
      strictPort: true,
    },
  });
  server.printUrls();
} else {
  throw new Error("Expected start or test as the local server command");
}
