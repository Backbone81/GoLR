// Build script for the extension bundle.
//
// VSCode loads a single JavaScript file as the extension entry point (see "main" in
// package.json). esbuild bundles src/extension.ts and everything it imports into
// dist/extension.js. We use esbuild rather than plain `tsc` because it is dramatically
// faster and produces one self-contained file, which is the recommended approach for
// shipping VSCode extensions.
//
// Usage:
//   node esbuild.js              one-off development build
//   node esbuild.js --watch      rebuild on every file change
//   node esbuild.js --production  minified build for packaging
const esbuild = require("esbuild");

const watch = process.argv.includes("--watch");
const production = process.argv.includes("--production");

async function main() {
  const context = await esbuild.context({
    entryPoints: ["src/extension.ts"],
    bundle: true,
    format: "cjs",
    // The extension runs in the VSCode extension host, which is a Node.js environment.
    platform: "node",
    outfile: "dist/extension.js",
    // The "vscode" module is provided by the runtime host, not bundled. Marking it external
    // tells esbuild to leave the `require("vscode")` call in place.
    external: ["vscode"],
    minify: production,
    sourcemap: !production,
    // ES2022 matches the tsconfig target and the Node version VSCode 1.120 ships.
    target: "ES2022",
    logLevel: "info",
  });

  if (watch) {
    await context.watch();
  } else {
    await context.rebuild();
    await context.dispose();
  }
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
