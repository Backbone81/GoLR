// Unit tests for deciding what completion offers at a caret.

import * as assert from "assert";
import { CompletionSite, completionAt } from "../../language/completion";
import { heredoc } from "./heredoc";

// Returns what completion offers at the <caret> marker in `text`.
function complete(text: string): CompletionSite | undefined {
  const source = heredoc(text);
  const offset = source.indexOf("<caret>");
  return completionAt(source.replace("<caret>", ""), offset);
}

suite("completion", () => {
  test("offers the section keywords at the top level", () => {
    const site = complete(`
      @scanner {
      A : "a" ;
      }
      <caret>
    `);
    assert.deepStrictEqual(site?.keywords, ["@scanner", "@parser"]);
    assert.strictEqual(site?.symbols, false);
  });

  test("offers the scanner annotations in a scanner rule body", () => {
    const site = complete(`
      @scanner {
      A : "a" <caret>
      }
    `);
    assert.deepStrictEqual(site?.keywords, ["@empty", "@skip", "@fragment"]);
    assert.strictEqual(site?.symbols, false);
  });

  // A new rule starts here, so no symbols are offered, only the statements of a @parser section.
  test("offers the parser statement keywords at the start of a statement", () => {
    const site = complete(`
      @parser {
      a : a ;
      <caret>
      }
    `);
    assert.deepStrictEqual(site?.keywords, ["@start", "@precedence"]);
    assert.strictEqual(site?.symbols, false);
  });

  test("offers the associativity keywords in a precedence block", () => {
    const site = complete(`
      @parser {
      @precedence {
      @left : a ;
      <caret>
      }
      a : a ;
      }
    `);
    assert.deepStrictEqual(site?.keywords, ["@left", "@right", "@none", "@precedence"]);
    assert.strictEqual(site?.symbols, false);
  });

  test("offers @error and symbols in a precedence line", () => {
    const site = complete(`
      @parser {
      @precedence {
      @left : a <caret> ;
      }
      }
    `);
    assert.deepStrictEqual(site?.keywords, ["@error"]);
    assert.strictEqual(site?.symbols, true);
  });

  test("offers only symbols after @start", () => {
    const site = complete(`
      @parser {
      @start : <caret> ;
      }
    `);
    assert.deepStrictEqual(site?.keywords, []);
    assert.strictEqual(site?.symbols, true);
  });

  test("offers only symbols inside @precedence(...)", () => {
    const site = complete(`
      @parser {
      a : a @precedence(<caret>) ;
      }
    `);
    assert.deepStrictEqual(site?.keywords, []);
    assert.strictEqual(site?.symbols, true);
  });

  test("offers the rule body keywords and symbols in a rule body", () => {
    const site = complete(`
      @parser {
      a : a <caret> ;
      }
    `);
    assert.deepStrictEqual(site?.keywords, ["@empty", "@error", "@precedence", "@name"]);
    assert.strictEqual(site?.symbols, true);
  });

  // After an "@" only a keyword can follow, so symbols are left out.
  test("offers only the keywords after an @, replacing from the @", () => {
    const text = heredoc(`
      @parser {
      a : a @na<caret> ;
      }
    `);
    const site = complete(text);
    assert.deepStrictEqual(site?.keywords, ["@empty", "@error", "@precedence", "@name"]);
    assert.strictEqual(site?.symbols, false);
    assert.strictEqual(site?.start, text.indexOf("@na"));
  });

  test("replaces a partially typed identifier from its start", () => {
    const text = heredoc(`
      @parser {
      expression : term ;
      term : expr<caret> ;
      }
    `);
    const site = complete(text);
    assert.strictEqual(site?.symbols, true);
    assert.strictEqual(site?.start, text.lastIndexOf("expr"));
  });

  // Inside @name(...) a new production name is written, so nothing is offered.
  test("offers nothing inside @name(...)", () => {
    const site = complete(`
      @parser {
      a : a @name(<caret>) ;
      }
    `);
    assert.deepStrictEqual(site?.keywords, []);
    assert.strictEqual(site?.symbols, false);
  });

  test("offers nothing between a rule name and its colon", () => {
    const site = complete(`
      @parser {
      a <caret> : a ;
      }
    `);
    assert.deepStrictEqual(site?.keywords, []);
    assert.strictEqual(site?.symbols, false);
  });

  test("offers nothing in a comment", () => {
    assert.strictEqual(
      complete(`
        @parser {
        // <caret>
        a : a ;
        }
      `),
      undefined,
    );
  });

  test("offers nothing in a string or a regex", () => {
    assert.strictEqual(
      complete(`
        @scanner {
        A : "a<caret>" ;
        }
      `),
      undefined,
    );
    assert.strictEqual(
      complete(`
        @scanner {
        A : /a<caret>/ ;
        }
      `),
      undefined,
    );
  });

  test("offers the scanner annotations right after a string", () => {
    const site = complete(`
      @scanner {
      A : "a"<caret>
      }
    `);
    assert.deepStrictEqual(site?.keywords, ["@empty", "@skip", "@fragment"]);
  });
});
