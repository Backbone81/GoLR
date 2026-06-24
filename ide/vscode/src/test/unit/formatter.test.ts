// Unit tests for the canonical formatter. Mirrors the IntelliJ plugin's GolrFormatterTest,
// including the key end-to-end check that the real golang.golr grammar is already in canonical
// form (i.e. formatting it is a no-op).

import * as assert from "assert";
import * as fs from "fs";
import * as path from "path";
import { format } from "../../language/formatter";

// Path from out/test/unit back to the repository root, then to the example grammar.
const GOLANG_GOLR = path.resolve(
  __dirname,
  "../../../../../examples/golang/spec/golang.golr",
);

suite("formatter", () => {
  test("empty input yields empty output", () => {
    assert.strictEqual(format(""), "");
    assert.strictEqual(format("   \n  \n"), "");
  });

  test("collapses messy scanner section into canonical, column-aligned form", () => {
    const input = `@scanner{
PLUS:"+";
INTEGER:/[0-9]+/;
}`;
    const expected = `@scanner {
    PLUS:    "+";
    INTEGER: /[0-9]+/;
}
`;
    assert.strictEqual(format(input), expected);
  });

  test("one parser alternative per line, led by ':' and '|', with ';' on its own line", () => {
    const input = `@parser{
expression:term "+" term|term;
}`;
    const expected = `@parser {
    expression
        : term "+" term
        | term
        ;
}
`;
    assert.strictEqual(format(input), expected);
  });

  test("keeps control directives single-line and tightens inline @precedence", () => {
    const input = `@parser{
@start:Program;
e:e "+" e @precedence ( PLUS );
}`;
    const expected = `@parser {
    @start: Program;
    e
        : e "+" e @precedence(PLUS)
        ;
}
`;
    assert.strictEqual(format(input), expected);
  });

  test("blank lines between items are preserved but collapsed to at most one", () => {
    const input = `@scanner {
A: "a";



B: "b";
}`;
    const expected = `@scanner {
    A: "a";

    B: "b";
}
`;
    assert.strictEqual(format(input), expected);
  });

  test("formatting is idempotent", () => {
    const input = `@parser{
expression:term "+" term|term;
term:INTEGER;
}`;
    const once = format(input);
    assert.strictEqual(format(once), once);
  });

  test("the real golang.golr grammar is already canonical (formatting is a no-op)", () => {
    const original = fs.readFileSync(GOLANG_GOLR, "utf8");
    assert.strictEqual(format(original), original);
  });
});
