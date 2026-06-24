// Auto-completion for GoLR files (Ctrl/Cmd-Space, or as you type).
//
// When the user is typing an identifier, we offer every symbol defined anywhere in the same
// file — terminals from @scanner and nonterminals from @parser — so rule bodies, precedence
// lines, and the @start declaration can be filled in without retyping names. This mirrors the
// IntelliJ plugin's GolrCompletionContributor, and it draws its suggestions from the same
// symbol model that backs Go to Definition / Find Usages / Rename, so completion never drifts
// out of sync with resolution.

import * as vscode from "vscode";
import { ModelCache } from "../language/modelCache";

export class GolrCompletionProvider implements vscode.CompletionItemProvider {
  constructor(private readonly cache: ModelCache) {}

  provideCompletionItems(
    document: vscode.TextDocument,
    _position: vscode.Position,
    _token: vscode.CancellationToken,
    _context: vscode.CompletionContext,
  ): vscode.CompletionItem[] {
    const model = this.cache.get(document);

    // De-duplicate by name in case a file declares the same symbol twice (the resolver
    // tolerates that, so completion should too).
    const seen = new Set<string>();
    const items: vscode.CompletionItem[] = [];
    for (const def of model.definitions) {
      if (seen.has(def.name)) continue;
      seen.add(def.name);

      const item = new vscode.CompletionItem(def.name, vscode.CompletionItemKind.Variable);
      // "terminal" / "nonterminal" detail mirrors the label used across the other features.
      item.detail = def.kind;
      items.push(item);
    }
    return items;
  }
}
