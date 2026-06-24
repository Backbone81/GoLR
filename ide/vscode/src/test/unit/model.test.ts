// Unit tests for the symbol model (definitions vs references, terminal vs nonterminal).
// Mirrors the resolution coverage of the IntelliJ plugin's GoToDefinition/FindUsages tests,
// but at the level of the pure model.

import * as assert from "assert";
import { buildModel } from "../../language/model";

suite("model", () => {
  test("scanner rules define terminals; their bodies have no references", () => {
    const model = buildModel(`
      @scanner {
        PLUS: "+";
        INT:  /[0-9]+/;
      }
    `);
    assert.deepStrictEqual(
      model.definitions.map((d) => [d.name, d.kind]),
      [
        ["PLUS", "terminal"],
        ["INT", "terminal"],
      ],
    );
    assert.strictEqual(model.references.length, 0);
  });

  test("parser rules define nonterminals; body identifiers are references", () => {
    const model = buildModel(`
      @parser {
        expression : term "+" term ;
        term : INT ;
      }
    `);
    assert.deepStrictEqual(
      model.definitions.map((d) => [d.name, d.kind]),
      [
        ["expression", "nonterminal"],
        ["term", "nonterminal"],
      ],
    );
    // term (x2) and INT are references; "+" is a string literal and not a reference.
    assert.deepStrictEqual(
      model.references.map((r) => r.name),
      ["term", "term", "INT"],
    );
  });

  test("@start declares a reference to the start symbol", () => {
    const model = buildModel(`
      @parser {
        @start : Program ;
        Program : INT ;
      }
    `);
    assert.ok(model.references.some((r) => r.name === "Program"));
    assert.ok(model.definitions.some((d) => d.name === "Program"));
  });

  test("precedence-line symbols are references", () => {
    const model = buildModel(`
      @parser {
        @precedence {
          @left : PLUS MINUS ;
        }
        e : e PLUS e ;
      }
    `);
    const refNames = model.references.map((r) => r.name);
    assert.ok(refNames.includes("PLUS"));
    assert.ok(refNames.includes("MINUS"));
  });

  test("inline @precedence(NAME) records NAME as a reference", () => {
    const model = buildModel(`
      @parser {
        e : e PLUS e @precedence(PLUS) ;
      }
    `);
    // Two PLUS references: one in the body, one inside @precedence(...).
    assert.strictEqual(model.references.filter((r) => r.name === "PLUS").length, 2);
  });

  test("symbolAt reports the occurrence under an offset", () => {
    const text = `@parser {
expression : term ;
}`;
    const model = buildModel(text);
    const defOffset = text.indexOf("expression") + 2;
    const refOffset = text.indexOf("term") + 1;

    const def = model.symbolAt(defOffset);
    assert.ok(def);
    assert.strictEqual(def!.name, "expression");
    assert.strictEqual(def!.isDefinition, true);

    const ref = model.symbolAt(refOffset);
    assert.ok(ref);
    assert.strictEqual(ref!.name, "term");
    assert.strictEqual(ref!.isDefinition, false);

    assert.strictEqual(model.symbolAt(text.indexOf("{")), undefined);
  });

  test("definitionsNamed / referencesNamed query by name", () => {
    const model = buildModel(`
      @parser {
        a : b b ;
        b : a ;
      }
    `);
    assert.strictEqual(model.definitionsNamed("a").length, 1);
    assert.strictEqual(model.referencesNamed("b").length, 2);
  });
});
