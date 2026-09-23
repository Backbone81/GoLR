// Unit tests for the symbol model: definitions vs references, terminal vs nonterminal.

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
    assert.deepStrictEqual(
      model.references.map((r) => r.name),
      ["term", '"+"', "term", "INT"],
    );
  });

  test("a terminal's string is its alias; a fragment has none", () => {
    const text = `
      @scanner {
        PLUS:  "+";
        WS:    " " @skip;
        DIGIT: /[0-9]/ @fragment;
        INT:   /{DIGIT}+/;
      }
    `;
    const model = buildModel(text);
    assert.deepStrictEqual(
      model.aliases.map((a) => [a.name, a.terminal]),
      [
        ['"+"', "PLUS"],
        ['" "', "WS"],
      ],
    );
    const plus = model.aliases[0];
    assert.strictEqual(text.slice(plus.start, plus.end), '"+"');
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

  test("precedence-line strings are references", () => {
    const model = buildModel(`
      @parser {
        @precedence {
          @left : "+" "-" ;
        }
      }
    `);
    assert.deepStrictEqual(
      model.references.map((r) => r.name),
      ['"+"', '"-"'],
    );
  });

  test("inline @precedence(SYMBOL) records a string SYMBOL as a reference", () => {
    const model = buildModel(`
      @parser {
        e : e "-" e @precedence("-") ;
      }
    `);
    assert.strictEqual(model.referencesNamed('"-"').length, 2);
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

  test("usagesNamed of a terminal covers references by name and by alias", () => {
    const model = buildModel(`
      @scanner {
        PLUS: "+";
      }
      @parser {
        @precedence {
          @left : "+" ;
        }
        e : e PLUS e | e "+" e ;
      }
    `);
    assert.strictEqual(model.usagesNamed("PLUS").length, 3);
    // A string is its own symbol: its usages are only the references by that string.
    assert.strictEqual(model.usagesNamed('"+"').length, 2);
  });

  test("declarationsNamed resolves a string to the alias and a name to the definition", () => {
    const model = buildModel(`
      @scanner {
        PLUS: "+";
      }
    `);
    assert.deepStrictEqual(
      model.declarationsNamed('"+"').map((d) => d.name),
      ['"+"'],
    );
    assert.deepStrictEqual(
      model.declarationsNamed("PLUS").map((d) => d.name),
      ["PLUS"],
    );
    assert.strictEqual(model.declarationsNamed('"-"').length, 0);
  });

  test("symbolAt reports a terminal's alias as a definition", () => {
    const text = `@scanner {
PLUS: "+";
}`;
    const model = buildModel(text);
    const alias = model.symbolAt(text.indexOf('"+"') + 1);
    assert.ok(alias);
    assert.strictEqual(alias!.name, '"+"');
    assert.strictEqual(alias!.isDefinition, true);
  });
});
