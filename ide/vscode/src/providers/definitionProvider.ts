// "Go to Definition" for GoLR symbols (Ctrl/Cmd-click, or F12).
//
// VSCode calls provideDefinition when the user invokes "Go to Definition" with the caret on a
// word. We look at the symbol under the caret — whether it is a reference (e.g. `term` in a
// rule body) or a definition itself — and return the location(s) of every matching definition
// in the same file. Returning several locations makes VSCode show a peek list, which mirrors
// the IntelliJ plugin's poly-resolve behaviour for accidentally-duplicated symbols.

import * as vscode from "vscode";
import { ModelCache } from "../language/modelCache";

export class GolrDefinitionProvider implements vscode.DefinitionProvider {
  constructor(private readonly cache: ModelCache) {}

  provideDefinition(
    document: vscode.TextDocument,
    position: vscode.Position,
    _token: vscode.CancellationToken,
  ): vscode.Definition | undefined {
    const model = this.cache.get(document);
    // Translate the caret position into a character offset so we can query the model.
    const occurrence = model.symbolAt(document.offsetAt(position));
    if (!occurrence) return undefined;

    const definitions = model.definitionsNamed(occurrence.name);
    if (definitions.length === 0) return undefined;

    return definitions.map(
      (def) =>
        new vscode.Location(
          document.uri,
          new vscode.Range(document.positionAt(def.start), document.positionAt(def.end)),
        ),
    );
  }
}
