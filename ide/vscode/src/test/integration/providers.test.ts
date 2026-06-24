// Integration tests: these run inside a real VSCode instance (launched by @vscode/test-cli)
// and drive the language features through VSCode's own command API — the same path the editor
// uses when a user clicks "Go to Definition", "Rename", etc. This is the VSCode equivalent of
// the IntelliJ plugin's BasePlatformTestCase tests.

import * as assert from "assert";
import * as path from "path";
import * as vscode from "vscode";

const EXTENSION_ID = "backbone81.golr";

// The fixture lives in the source tree (it is not compiled, so it stays under src/).
const FIXTURE = path.resolve(
  __dirname,
  "../../../src/test/integration/fixtures/sample.golr",
);

// Opens the fixture document and makes sure the extension is active before returning.
async function openFixture(): Promise<vscode.TextDocument> {
  const extension = vscode.extensions.getExtension(EXTENSION_ID);
  assert.ok(extension, `extension ${EXTENSION_ID} should be installed in the test host`);
  await extension!.activate();
  const document = await vscode.workspace.openTextDocument(FIXTURE);
  await vscode.window.showTextDocument(document);
  return document;
}

// Returns the Position at the start of the `occurrence`-th (0-based) appearance of `needle`.
function positionOf(
  document: vscode.TextDocument,
  needle: string,
  occurrence = 0,
): vscode.Position {
  const text = document.getText();
  let index = -1;
  for (let i = 0; i <= occurrence; i++) {
    index = text.indexOf(needle, index + 1);
    assert.notStrictEqual(index, -1, `could not find occurrence ${occurrence} of "${needle}"`);
  }
  // Aim a couple of characters into the word so the caret is unambiguously inside it.
  return document.positionAt(index + 1);
}

suite("GoLR language features (integration)", () => {
  test("Go to Definition jumps from a reference to its definition", async () => {
    const document = await openFixture();
    // "term" appears three times: as references in the `expression` body (occurrences 0 and 1)
    // and as the defining rule name (occurrence 2). Jumping from a reference must land on the
    // definition.
    const refPos = positionOf(document, "term", 0);

    const locations = await vscode.commands.executeCommand<vscode.Location[]>(
      "vscode.executeDefinitionProvider",
      document.uri,
      refPos,
    );

    assert.ok(locations && locations.length >= 1, "expected at least one definition location");
    const defPos = positionOf(document, "term", 2);
    assert.ok(
      locations.some((loc) => loc.range.start.line === defPos.line),
      "definition should be on the line of the 'term' rule",
    );
  });

  test("Find All References lists every usage of a symbol", async () => {
    const document = await openFixture();
    const defPos = positionOf(document, "term", 2); // the 'term' definition (occurrence 2)

    const locations = await vscode.commands.executeCommand<vscode.Location[]>(
      "vscode.executeReferenceProvider",
      document.uri,
      defPos,
    );

    assert.ok(locations, "expected reference results");
    // "term" is referenced twice in the body of `expression`.
    const refCount = locations.filter((loc) => {
      const word = document.getText(loc.range);
      return word === "term";
    }).length;
    assert.ok(refCount >= 2, `expected >= 2 references, got ${refCount}`);
  });

  test("Rename produces edits for the definition and every reference", async () => {
    const document = await openFixture();
    const refPos = positionOf(document, "term", 0);

    const edit = await vscode.commands.executeCommand<vscode.WorkspaceEdit>(
      "vscode.executeDocumentRenameProvider",
      document.uri,
      refPos,
      "factor",
    );

    assert.ok(edit, "expected a workspace edit");
    const edits = edit.get(document.uri);
    // 1 definition + 2 references = 3 edits.
    assert.strictEqual(edits.length, 3, "rename should touch the definition and both references");
    for (const e of edits) {
      assert.strictEqual(e.newText, "factor");
    }
  });

  test("Completion offers the file's defined symbols with terminal/nonterminal detail", async () => {
    const document = await openFixture();
    // Inside the body of the `term` rule, after INTEGER.
    const pos = positionOf(document, "INTEGER", 1);

    const list = await vscode.commands.executeCommand<vscode.CompletionList>(
      "vscode.executeCompletionItemProvider",
      document.uri,
      pos,
    );

    assert.ok(list, "expected a completion list");
    const labels = list.items.map((i) => (typeof i.label === "string" ? i.label : i.label.label));
    for (const name of ["PLUS", "INTEGER", "expression", "term"]) {
      assert.ok(labels.includes(name), `completion should offer "${name}"`);
    }
    const terminal = list.items.find(
      (i) => (typeof i.label === "string" ? i.label : i.label.label) === "PLUS",
    );
    assert.strictEqual(terminal?.detail, "terminal");
  });

  test("CodeLens shows a reference count above each definition", async () => {
    const document = await openFixture();
    const lenses = await vscode.commands.executeCommand<vscode.CodeLens[]>(
      "vscode.executeCodeLensProvider",
      document.uri,
    );

    assert.ok(lenses, "expected CodeLenses");
    // One lens per definition: PLUS, INTEGER, expression, term.
    assert.strictEqual(lenses.length, 4);

    // The 'term' nonterminal is referenced twice in the body of `expression`.
    const termDefLine = positionOf(document, "term", 2).line;
    const termLens = lenses.find((l) => l.range.start.line === termDefLine);
    assert.ok(termLens, "expected a lens on the 'term' definition line");
    assert.strictEqual(termLens!.command?.title, "2 references");
    assert.strictEqual(termLens!.command?.command, "editor.action.showReferences");
  });

  test("Semantic tokens are produced for the document", async () => {
    const document = await openFixture();
    const tokens = await vscode.commands.executeCommand<vscode.SemanticTokens>(
      "vscode.provideDocumentSemanticTokens",
      document.uri,
    );
    assert.ok(tokens, "expected semantic tokens");
    // Each token is encoded as 5 integers; a non-empty grammar must yield several.
    assert.ok(tokens.data.length >= 5, "expected at least one semantic token");
    assert.strictEqual(tokens.data.length % 5, 0, "token data must be a multiple of 5");
  });

  test("Format Document rewrites messy input into canonical layout", async () => {
    // Open a fresh in-memory GoLR document with deliberately messy spacing.
    const messy = `@parser{\nexpression:term "+" term|term;\nterm:INTEGER;\n}`;
    const document = await vscode.workspace.openTextDocument({ language: "golr", content: messy });
    await vscode.window.showTextDocument(document);

    const edits = await vscode.commands.executeCommand<vscode.TextEdit[]>(
      "vscode.executeFormatDocumentProvider",
      document.uri,
      { tabSize: 4, insertSpaces: true },
    );

    assert.ok(edits && edits.length > 0, "messy input should produce formatting edits");

    // Apply the edits and confirm the document now holds the canonical layout.
    const workspaceEdit = new vscode.WorkspaceEdit();
    for (const e of edits) workspaceEdit.replace(document.uri, e.range, e.newText);
    await vscode.workspace.applyEdit(workspaceEdit);

    const expected = `@parser {
    expression
        : term "+" term
        | term
        ;
    term
        : INTEGER
        ;
}
`;
    assert.strictEqual(document.getText(), expected);
  });

  test("Format Document is a no-op on already-canonical input", async () => {
    const document = await openFixture();
    const edits = await vscode.commands.executeCommand<vscode.TextEdit[] | undefined>(
      "vscode.executeFormatDocumentProvider",
      document.uri,
      { tabSize: 4, insertSpaces: true },
    );
    // When the formatter requests no changes, VSCode reports either an empty list or undefined.
    assert.ok(!edits || edits.length === 0, "canonical input should need no formatting edits");
  });
});
