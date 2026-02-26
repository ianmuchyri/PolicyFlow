/**
 * Generates MDX API docs from openapi.yaml into content/docs/api/.
 * Run: node scripts/generate-docs.mjs
 * Or via: pnpm generate
 */
import { createOpenAPI } from "fumadocs-openapi/server";
import { generateFiles } from "fumadocs-openapi";

const openapi = createOpenAPI({
  input: ["./openapi.yaml"],
});

await generateFiles({
  input: openapi,
  output: "./content/docs/api",
  per: "tag",
  includeDescription: true,
});

console.log("✅ API docs generated → content/docs/api/");
