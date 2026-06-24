// Semantic (context-aware) highlighting for GoLR files.
//
// The bundled TextMate grammar (syntaxes/golr.tmLanguage.json) already colours comments,
// strings, regexes, and @-keywords quickly and cheaply. What it cannot do is tell a *symbol
// definition* apart from a *symbol reference*, or a terminal apart from a nonterminal —
// because that requires understanding the file's structure, not just matching patterns.
//
// A semantic tokens provider fills that gap. VSCode asks us to classify the meaningful tokens
// in the document; we walk the symbol model and tag each definition/reference with a token
// type (terminal vs nonterminal) and, for definitions, a "declaration" modifier. The theme
// then colours them, giving the same def-vs-use distinction the IntelliJ plugin's PSI-based
// highlighter provides.
//
// The mapping from these token types to actual colours is declared in package.json under
// contributes.semanticTokenScopes.

import * as vscode from "vscode";
import { ModelCache } from "../language/modelCache";

// Token types we emit. "class" reads as a nonterminal (a parser rule), "enum" as a terminal (a
// scanner rule); these standard names let common themes colour them without extra setup, and
// package.json refines the mapping. The "declaration" modifier marks defining occurrences.
const TOKEN_TYPES = ["class", "enum"] as const;
const TOKEN_MODIFIERS = ["declaration"] as const;

export const golrSemanticTokensLegend = new vscode.SemanticTokensLegend(
  [...TOKEN_TYPES],
  [...TOKEN_MODIFIERS],
);

type TokenTypeName = (typeof TOKEN_TYPES)[number];

export class GolrSemanticTokensProvider implements vscode.DocumentSemanticTokensProvider {
  constructor(private readonly cache: ModelCache) {}

  provideDocumentSemanticTokens(
    document: vscode.TextDocument,
    _token: vscode.CancellationToken,
  ): vscode.SemanticTokens {
    const model = this.cache.get(document);
    const builder = new vscode.SemanticTokensBuilder(golrSemanticTokensLegend);

    // Collect every occurrence first so we can emit them in document order — the
    // SemanticTokensBuilder requires tokens to be pushed sorted by position.
    interface Entry {
      start: number;
      end: number;
      type: TokenTypeName;
      modifiers: string[];
    }
    const entries: Entry[] = [];

    for (const def of model.definitions) {
      entries.push({
        start: def.start,
        end: def.end,
        type: def.kind === "terminal" ? "enum" : "class",
        modifiers: ["declaration"],
      });
    }
    for (const ref of model.references) {
      // A reference's colour follows the kind of the symbol it resolves to. If it resolves to
      // nothing (an unknown name), fall back to the nonterminal colour.
      const target = model.definitionsNamed(ref.name)[0];
      entries.push({
        start: ref.start,
        end: ref.end,
        type: target?.kind === "terminal" ? "enum" : "class",
        modifiers: [],
      });
    }

    entries.sort((a, b) => a.start - b.start);

    for (const entry of entries) {
      const range = new vscode.Range(
        document.positionAt(entry.start),
        document.positionAt(entry.end),
      );
      builder.push(range, entry.type, entry.modifiers);
    }

    return builder.build();
  }
}
