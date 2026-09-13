// Mechanical port of internal/fmt/formatter_test.go, kept as close as possible to that file's
// Describe/Context/It structure (each suite/test here corresponds 1:1 to one Go
// Describe/Context/It) and to its fixtures (each utils.HereDoc(`...`) call becomes a
// heredoc(`...`) call on the same content, closing backtick aligned with the "const ... ="
// line the same way Go's closing backtick aligns with "... := ") so that a new Go test case has
// an obvious, same-shaped place to add here too.

import * as assert from "assert";
import * as fs from "fs";
import * as path from "path";
import { format } from "../../language/formatter";
import { heredoc } from "./heredoc";

suite("GoLR formatting", () => {
  suite("basic layout", () => {
    test("returns empty output for empty input", () => {
      assert.strictEqual(format(""), "");
    });

    test("returns empty output for whitespace-only input", () => {
      assert.strictEqual(format("   \n  \n"), "");
    });

    test("column-aligns a messy scanner section", () => {
      const input = heredoc(`
          @scanner{
          PLUS:"+";
          INTEGER:/[0-9]+/;
          }
      `);
      const expected = heredoc(`
          @scanner {
              PLUS:    "+";
              INTEGER: /[0-9]+/;
          }
      `);
      assert.strictEqual(format(input), expected);
    });

    test("puts one parser alternative per line, led by ':' and '|', with ';' on its own line", () => {
      const input = heredoc(`
          @parser{
          expression:term "+" term|term;
          }
      `);
      const expected = heredoc(`
          @parser {
              expression
                  : term "+" term
                  | term
                  ;
          }
      `);
      assert.strictEqual(format(input), expected);
    });

    test("keeps control directives single-line and tightens inline @precedence", () => {
      const input = heredoc(`
          @parser{
          @start:Program;
          e:e "+" e @precedence ( PLUS );
          }
      `);
      const expected = heredoc(`
          @parser {
              @start: Program;

              e
                  : e "+" e @precedence(PLUS)
                  ;
          }
      `);
      assert.strictEqual(format(input), expected);
    });

    test("collapses multiple blank lines between items to at most one", () => {
      const input = heredoc(`
          @scanner {
          A: "a";



          B: "b";
          }
      `);
      const expected = heredoc(`
          @scanner {
              A: "a";

              B: "b";
          }
      `);
      assert.strictEqual(format(input), expected);
    });

    test("is idempotent", () => {
      const input = heredoc(`
          @parser{
          expression:term "+" term|term;
          term:INTEGER;
          }
      `);
      const once = format(input);
      const twice = format(once);
      assert.strictEqual(twice, once);
    });

    test("tightens an inline @name(...) annotation the same way it tightens @precedence(...)", () => {
      const input = heredoc(`
          @parser{
          file:scanner_section parser_section @name ( file );
          }
      `);
      const expected = heredoc(`
          @parser {
              file
                  : scanner_section parser_section @name(file)
                  ;
          }
      `);
      assert.strictEqual(format(input), expected);
    });

    test("does not invent a blank line inside an empty scanner block", () => {
      // The "}" arrives right after "{" with nothing in between, so it is already at the start
      // of a fresh, indented line; it must not schedule a linebreak of its own on top of that.
      const expected = heredoc(`
          @scanner {
          }
      `);
      assert.strictEqual(format("@scanner{}"), expected);
    });

    test("does not invent a blank line inside an empty parser block", () => {
      const expected = heredoc(`
          @parser {
          }
      `);
      assert.strictEqual(format("@parser{}"), expected);
    });
  });

  suite("comments", () => {
    test("preserves a leading top-level comment and the blank line after it", () => {
      const input = heredoc(`
          // file header

          @scanner {
              PLUS: "+";
          }
      `);
      assert.strictEqual(format(input), input);
    });

    test("preserves a line comment immediately before a scanner rule", () => {
      const input = heredoc(`
          @scanner {
              // marks the arithmetic operators
              PLUS: "+";
          }
      `);
      assert.strictEqual(format(input), input);
    });

    test("preserves a block comment on its own line inside a parser section", () => {
      const input = heredoc(`
          @parser {
              /* entry point */
              file
                  : @empty
                  ;
          }
      `);
      assert.strictEqual(format(input), input);
    });

    test("keeps a comment documenting one alternative on its own line, not run onto the previous one", () => {
      // This is the idiom the project's own bootstrap grammar (golr.golr) uses to document
      // individual error-recovery alternatives. A comment between two "|" alternatives must stay
      // on its own line rather than being swept into the token stream of the alternative before
      // it.
      const input = heredoc(`
          @parser {
              expression
                  : a

                  // explains the next alternative
                  | b
                  | c
                  ;
          }
      `);
      assert.strictEqual(format(input), input);
    });

    test("preserves a trailing comment before a closing brace, along with its blank line", () => {
      const input = heredoc(`
          @scanner {
              PLUS: "+";

              // trailing note
          }
      `);
      assert.strictEqual(format(input), input);
    });

    test("keeps a comment trailing a scanner rule on the same line, without an extra blank line after it", () => {
      // Regression test: a trailing comment used to leave a linebreak pending for the token
      // after it (here ";"), which the "}" then scheduled a second one on top of, producing a
      // spurious blank line. It also used to misuse the pending indent as the separator before
      // "//", widening a single space into a full indentation string.
      const input = heredoc(`
          @scanner {
              PLUS: "+"; // marks the plus operator
          }
      `);
      assert.strictEqual(format(input), input);
    });

    test("keeps a comment trailing the opening brace of a section on the same line", () => {
      const input = heredoc(`
          @scanner { // arithmetic operators
              PLUS: "+";
          }
      `);
      assert.strictEqual(format(input), input);
    });

    test("keeps a comment trailing one alternative on the same line, without an extra blank line before the ';'", () => {
      // Regression test: the same spurious-blank-line bug as above, but reached through
      // onTokenSemi's parser-rule branch instead of onTokenRbrace, and with a fresh indent level
      // in between.
      const input = heredoc(`
          @parser {
              e
                  : a
                  | b // choose a or b
                  ;
          }
      `);
      assert.strictEqual(format(input), input);
    });
  });

  suite("malformed input", () => {
    test("does not fail on an unterminated block comment and keeps its bytes", () => {
      // The input deliberately ends without a trailing newline: it represents a source cut off
      // mid-comment, and closing the template literal on its own line would put a newline into
      // the string that was never there.
      const input = heredoc(`
          @scanner {
              PLUS: "+"; /* unterminated`);
      const output = format(input);
      assert.ok(output.includes("PLUS"));
      assert.ok(output.includes("unterminated"));
    });

    test("keeps a rule name whose body is still being typed, without inventing the ';' or '}' it doesn't have yet", () => {
      // A formatter repositions whitespace around the tokens that exist; it must not invent a
      // closing '}' (or ':'/';') the source never had, even to keep the output superficially
      // well-formed. The input deliberately ends without a trailing newline: it represents a
      // file still being typed, and closing the template literal on its own line would put a
      // newline into the string that was never there.
      const input = heredoc(`
          @parser {
              file
                  : @empty
                  ;

              partial`);
      const expected = heredoc(`
          @parser {
              file
                  : @empty
                  ;

              partial
      `);
      assert.strictEqual(format(input), expected);
    });

    test("closes the block right after a rule missing its terminating ';'", () => {
      // The identifier's own emit leaves indentNext false (no linebreak was scheduled after it,
      // since the ';' that would normally do so is missing); the closing "}" must still start
      // its own line instead of trailing "PLUS" on the same one.
      const expected = heredoc(`
          @scanner {
              PLUS
          }
      `);
      assert.strictEqual(format("@scanner{PLUS}"), expected);
    });
  });

  suite("real grammar files", () => {
    test("keeps every canonical .golr file in the repository unchanged", () => {
      // "ide" carries its own, deliberately non-canonical fixtures for the extension's own
      // formatter tests, so it is excluded rather than treated as a regression.
      const repoRoot = findRepoRoot();
      if (repoRoot === undefined) {
        return;
      }

      const paths: string[] = [];
      walkGolrFiles(repoRoot, paths);
      assert.ok(paths.length > 0, `expected to find at least one .golr file under ${repoRoot}`);

      for (const filePath of paths) {
        const data = fs.readFileSync(filePath, "utf8");
        assert.strictEqual(format(data), data, filePath);
      }
    });
  });
});

// Walks upward from this file's directory looking for go.mod, the repo root's marker file.
// Go's version relies on "go test"'s cwd always being the package directory ("../.." from
// internal/fmt is the repo root, deterministically); a compiled Mocha test has no such
// guarantee, so this looks for the marker file instead - a mechanical adaptation of "repoRoot",
// not a behavioral one.
function findRepoRoot(): string | undefined {
  let dir = __dirname;
  for (;;) {
    if (fs.existsSync(path.join(dir, "go.mod"))) {
      return dir;
    }
    const parent = path.dirname(dir);
    if (parent === dir) {
      return undefined;
    }
    dir = parent;
  }
}

// Collects every ".golr" file under dir, skipping any directory named "ide".
function walkGolrFiles(dir: string, paths: string[]): void {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.name === "ide") {
      continue;
    }
    const entryPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      walkGolrFiles(entryPath, paths);
    } else if (entry.isFile() && entry.name.endsWith(".golr")) {
      paths.push(entryPath);
    }
  }
}
