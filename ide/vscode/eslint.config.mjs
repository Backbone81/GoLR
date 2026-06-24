// ESLint flat config (ESLint 9+). Lints the TypeScript sources with the recommended
// TypeScript rules. Running the linter is a VSCode-extension best practice and is wired into
// the `lint` npm script and the Makefile `prepare`-style checks.
import js from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
  {
    ignores: ["dist/**", "out/**", "node_modules/**"],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["src/**/*.ts"],
    rules: {
      // The VSCode provider interfaces dictate a fixed parameter list even when a particular
      // provider does not need every argument (e.g. the cancellation token). We follow the
      // common convention of prefixing those intentionally-unused parameters with "_" and tell
      // the linter to ignore them.
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
    },
  },
);
