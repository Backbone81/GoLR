// Rename refactoring for GoLR symbols (F2).
//
// Renaming happens in two steps that VSCode drives:
//
//   1. prepareRename — called when the user presses F2. We confirm the caret is actually on a
//      GoLR symbol and tell VSCode which text range is being renamed (so it can show the inline
//      rename box pre-filled with the current name). Throwing here makes VSCode show
//      "You cannot rename this element".
//
//   2. provideRenameEdits — called when the user confirms a new name. We validate the new name
//      (using the same identifier rule as the lexer, mirroring the IntelliJ plugin's
//      GolrNamesValidator) and return a WorkspaceEdit that rewrites the definition together
//      with every reference, so the whole file stays consistent.

import * as vscode from "vscode";
import { ModelCache } from "../language/modelCache";

// A GoLR identifier starts with a letter or underscore and continues with letters, digits, or
// underscores. The \p{...} classes make this Unicode-aware, matching the tokenizer.
const IDENTIFIER_RE = /^[\p{L}_][\p{L}\p{Nd}_]*$/u;

export class GolrRenameProvider implements vscode.RenameProvider {
  constructor(private readonly cache: ModelCache) {}

  prepareRename(
    document: vscode.TextDocument,
    position: vscode.Position,
    _token: vscode.CancellationToken,
  ): vscode.Range {
    const model = this.cache.get(document);
    const occurrence = model.symbolAt(document.offsetAt(position));
    if (!occurrence) {
      // Rejecting here is how a RenameProvider says "this element is not renameable".
      throw new Error("You cannot rename this element.");
    }
    return new vscode.Range(
      document.positionAt(occurrence.start),
      document.positionAt(occurrence.end),
    );
  }

  provideRenameEdits(
    document: vscode.TextDocument,
    position: vscode.Position,
    newName: string,
    _token: vscode.CancellationToken,
  ): vscode.WorkspaceEdit | undefined {
    if (!IDENTIFIER_RE.test(newName)) {
      throw new Error(`'${newName}' is not a valid GoLR identifier.`);
    }

    const model = this.cache.get(document);
    const occurrence = model.symbolAt(document.offsetAt(position));
    if (!occurrence) return undefined;

    const edit = new vscode.WorkspaceEdit();
    const rename = (start: number, end: number): void => {
      edit.replace(
        document.uri,
        new vscode.Range(document.positionAt(start), document.positionAt(end)),
        newName,
      );
    };

    // Rewrite every definition and every reference that shares the old name. Renaming from a
    // reference therefore updates the definition too, exactly like the IntelliJ plugin.
    for (const def of model.definitionsNamed(occurrence.name)) rename(def.start, def.end);
    for (const ref of model.referencesNamed(occurrence.name)) rename(ref.start, ref.end);

    return edit;
  }
}
