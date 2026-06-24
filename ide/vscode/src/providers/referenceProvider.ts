// "Find All References" / "Find Usages" for GoLR symbols (Shift+F12).
//
// VSCode calls provideReferences with the caret on a symbol and a `context` flag that says
// whether the symbol's own declaration should be included in the results. We resolve the
// symbol under the caret by name and return every reference site, plus (optionally) every
// definition site.
//
// This mirrors the IntelliJ plugin's Find Usages, where a search from a definition lists all
// references that resolve back to it.

import * as vscode from "vscode";
import { ModelCache } from "../language/modelCache";

export class GolrReferenceProvider implements vscode.ReferenceProvider {
  constructor(private readonly cache: ModelCache) {}

  provideReferences(
    document: vscode.TextDocument,
    position: vscode.Position,
    context: vscode.ReferenceContext,
    _token: vscode.CancellationToken,
  ): vscode.Location[] | undefined {
    const model = this.cache.get(document);
    const occurrence = model.symbolAt(document.offsetAt(position));
    if (!occurrence) return undefined;

    const toLocation = (start: number, end: number): vscode.Location =>
      new vscode.Location(
        document.uri,
        new vscode.Range(document.positionAt(start), document.positionAt(end)),
      );

    const locations: vscode.Location[] = model
      .referencesNamed(occurrence.name)
      .map((ref) => toLocation(ref.start, ref.end));

    // includeDeclaration is true for "Find All References" and false when VSCode only wants
    // the non-declaration usages.
    if (context.includeDeclaration) {
      for (const def of model.definitionsNamed(occurrence.name)) {
        locations.push(toLocation(def.start, def.end));
      }
    }

    return locations;
  }
}
