// Configuration for @vscode/test-cli, which downloads a throwaway VSCode build, launches it
// with this extension loaded, and runs the integration tests inside the real extension host.
//
// These tests exercise the actual VSCode provider APIs (definition, references, rename,
// completion, semantic tokens, formatting) end to end, which is why they need a live host —
// unlike the unit tests, which run in plain Node via Mocha.
//
// The suite runs twice, against two different VSCode builds:
//
//   - the floor, which is the oldest VSCode we claim to support. "engines.vscode" is a
//     minimum, not a range, so this is the version our users may actually be on and the one
//     which proves we use no API newer than we promise. It is read straight out of
//     package.json, so raising the declared floor moves the test with it and the two can
//     never drift apart.
//   - stable, which is whatever VSCode has released by the time the tests run. This one is an
//     early warning that an upcoming VSCode broke us. Note that it is a moving target: this
//     run can start failing without any change on our side.
import { defineConfig } from "@vscode/test-cli";
import { createRequire } from "node:module";

// package.json is JSON, not a module, so it is read through a CommonJS require rather than
// an import assertion.
const require = createRequire(import.meta.url);

// "^1.120.0" -> "1.120.0". \D is the non-digit class, so this strips the leading range
// operator and nothing else. @vscode/test-electron wants a bare version here.
const floorVersion = require("./package.json").engines.vscode.replace(/^\D+/, "");

const common = {
  // The integration tests are compiled from src/test/integration into out/test/integration
  // by `npm run compile-tests`.
  files: "out/test/integration/**/*.test.js",
  mocha: {
    timeout: 20000,
  },
};

export default defineConfig([
  { label: `floor-${floorVersion}`, version: floorVersion, ...common },
  { label: "stable", version: "stable", ...common },
]);
