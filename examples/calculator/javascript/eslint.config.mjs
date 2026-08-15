import js from "@eslint/js";
import globals from "globals";

// The generated scanner and parser are linted along with the hand written main.js. They are the reason this config
// exists: "node --check" only parses a file, so nothing else in the build would notice an unused binding or a typo in
// a name which is never called.
export default [
    js.configs.recommended,
    {
        languageOptions: {
            ecmaVersion: "latest",
            sourceType: "module",
            // The calculator is a command line program, so it uses process, and the scanner works on bytes, so it uses
            // TextEncoder and TextDecoder.
            globals: globals.node,
        },
    },
];
