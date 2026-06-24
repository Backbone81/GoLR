// Unit test for the CodeLens label helper. The provider itself depends on the `vscode` module
// and is covered by the integration suite; the pluralization logic is pure and tested here.

import * as assert from "assert";
import { referenceCountLabel } from "../../language/referenceLabel";

suite("codeLens label", () => {
  test("uses the singular form for exactly one reference", () => {
    assert.strictEqual(referenceCountLabel(1), "1 reference");
  });

  test("uses the plural form for zero and many references", () => {
    assert.strictEqual(referenceCountLabel(0), "0 references");
    assert.strictEqual(referenceCountLabel(2), "2 references");
    assert.strictEqual(referenceCountLabel(42), "42 references");
  });
});
