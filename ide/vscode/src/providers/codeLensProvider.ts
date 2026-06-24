// "N references" CodeLens above every GoLR symbol definition.
//
// This is the VSCode counterpart of the IntelliJ plugin's GolrReferencesCodeVisionProvider,
// which draws a clickable "N usages" indicator above each rule. VSCode's equivalent mechanism
// is CodeLens: a small actionable annotation rendered on the line above a range.
//
// For each symbol definition we render a lens labelled with how many references the symbol has
// in the file. Clicking it invokes the built-in `editor.action.showReferences` command, which
// opens VSCode's reference peek — reusing exactly the reference set the Find All References
// feature already computes from the same symbol model.

import * as vscode from "vscode";
import { ModelCache } from "../language/modelCache";
import { referenceCountLabel } from "../language/referenceLabel";

export class GolrCodeLensProvider implements vscode.CodeLensProvider {
  constructor(private readonly cache: ModelCache) {}

  provideCodeLenses(
    document: vscode.TextDocument,
    _token: vscode.CancellationToken,
  ): vscode.CodeLens[] {
    const model = this.cache.get(document);
    const lenses: vscode.CodeLens[] = [];

    for (const def of model.definitions) {
      const definitionRange = new vscode.Range(
        document.positionAt(def.start),
        document.positionAt(def.end),
      );

      // The reference sites this lens will reveal when clicked.
      const referenceLocations = model
        .referencesNamed(def.name)
        .map(
          (ref) =>
            new vscode.Location(
              document.uri,
              new vscode.Range(document.positionAt(ref.start), document.positionAt(ref.end)),
            ),
        );

      // Build the lens with its command already filled in. (CodeLens supports a lazy
      // resolveCodeLens step, but the count is cheap to compute, so we do it eagerly.)
      lenses.push(
        new vscode.CodeLens(definitionRange, {
          title: referenceCountLabel(referenceLocations.length),
          command: "editor.action.showReferences",
          // showReferences expects: the document URI, the position to anchor the peek on, and
          // the list of locations to show.
          arguments: [document.uri, definitionRange.start, referenceLocations],
        }),
      );
    }

    return lenses;
  }
}
