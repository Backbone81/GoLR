// Auto-completion for GoLR files (Ctrl/Cmd-Space, or as you type).
//
// We offer the keywords valid at the caret (see src/language/completion.ts), and where a symbol
// may be referenced — rule bodies, precedence lines, @start, @precedence(...) — every symbol
// defined anywhere in the same file: terminals from @scanner and nonterminals from @parser. The
// symbols come from the same symbol model that backs Go to Definition / Find Usages / Rename,
// so completion never drifts out of sync with resolution.

import * as vscode from "vscode";
import { completionAt } from "../language/completion";
import { ModelCache } from "../language/modelCache";

export class GolrCompletionProvider implements vscode.CompletionItemProvider {
  constructor(private readonly cache: ModelCache) {}

  provideCompletionItems(
    document: vscode.TextDocument,
    position: vscode.Position,
    _token: vscode.CancellationToken,
    _context: vscode.CompletionContext,
  ): vscode.CompletionItem[] {
    const site = completionAt(document.getText(), document.offsetAt(position));
    if (!site) return [];

    // The word range VSCode derives leaves out the "@" of a keyword, so we replace our own.
    const range = new vscode.Range(document.positionAt(site.start), position);
    const items: vscode.CompletionItem[] = [];

    for (const keyword of site.keywords) {
      const item = new vscode.CompletionItem(keyword, vscode.CompletionItemKind.Keyword);
      item.range = range;
      items.push(item);
    }

    if (site.symbols) {
      // De-duplicate by name in case a file declares the same symbol twice (the resolver
      // tolerates that, so completion should too).
      const seen = new Set<string>();
      for (const def of this.cache.get(document).definitions) {
        if (seen.has(def.name)) continue;
        seen.add(def.name);

        const item = new vscode.CompletionItem(def.name, vscode.CompletionItemKind.Variable);
        // "terminal" / "nonterminal" detail mirrors the label used across the other features.
        item.detail = def.kind;
        item.range = range;
        items.push(item);
      }
    }

    return items;
  }
}
