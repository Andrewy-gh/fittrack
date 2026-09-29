import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, test } from "vitest";

const clientRoot = process.cwd();
const canonicalOrigin = "https://fittrack.fly.dev";

function readClientFile(relativePath: string) {
  return readFileSync(resolve(clientRoot, relativePath), "utf8");
}

describe("static agent-discovery assets", () => {
  test("sitemap is valid XML containing only public canonical pages", () => {
    const xml = readClientFile("public/sitemap.xml");
    const document = new DOMParser().parseFromString(xml, "application/xml");

    expect(document.querySelector("parsererror")).toBeNull();
    expect(document.documentElement.localName).toBe("urlset");
    expect(document.documentElement.namespaceURI).toBe(
      "http://www.sitemaps.org/schemas/sitemap/0.9",
    );
    expect(
      [...document.getElementsByTagName("loc")].map(
        (location) => location.textContent,
      ),
    ).toEqual([`${canonicalOrigin}/`, `${canonicalOrigin}/privacy`]);
  });

  test("llms describes public resources without claiming private data is public", () => {
    const llms = readClientFile("public/llms.txt");

    expect(llms).toContain("# FitTrack");
    expect(llms).toContain(`[Privacy Policy](${canonicalOrigin}/privacy)`);
    expect(llms).toContain(`[API documentation](${canonicalOrigin}/swagger/)`);
    expect(llms).toContain(
      `[Read-only agent API specification](${canonicalOrigin}/.well-known/agent-openapi.json)`,
    );
    expect(llms).not.toContain(`${canonicalOrigin}/swagger/doc.json`);
    expect(llms).toContain(`[API health](${canonicalOrigin}/health)`);
    expect(llms).toContain(
      "Interactive application data is private to authenticated users.",
    );
    expect(llms).toContain(
      "does not provide unauthenticated access to workout or account data",
    );
  });
});
