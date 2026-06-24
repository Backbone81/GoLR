// Configuration for @vscode/test-cli, which downloads a throwaway VSCode build, launches it
// with this extension loaded, and runs the integration tests inside the real extension host.
//
// These tests exercise the actual VSCode provider APIs (definition, references, rename,
// completion, semantic tokens, formatting) end to end, which is why they need a live host —
// unlike the unit tests, which run in plain Node via Mocha.
import { defineConfig } from "@vscode/test-cli";

export default defineConfig({
  // The integration tests are compiled from src/test/integration into out/test/integration
  // by `npm run compile-tests`.
  files: "out/test/integration/**/*.test.js",
  mocha: {
    timeout: 20000,
  },
});
