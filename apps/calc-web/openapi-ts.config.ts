import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../../contracts/calc/openapi/schema.yml",
  output: {
    path: "src/shared/api/generated",
    postProcess: ["prettier"],
    tsConfigPath: "./tsconfig.app.json",
  },
  plugins: ["@hey-api/client-ofetch"],
});
