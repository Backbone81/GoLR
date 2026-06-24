// Code formatting for GoLR files ("Format Document", Shift+Alt+F).
//
// VSCode calls provideDocumentFormattingEdits and expects a list of text edits that transform
// the document into its formatted form. Because our formatter rewrites the entire file into a
// canonical layout (see src/language/formatter.ts, a port of the IntelliJ GolrFormatter), the
// simplest and most robust approach is to return a single edit that replaces the whole
// document with the formatted text. VSCode diffs the old and new text internally, so the
// editor's undo history and cursor position stay sensible.

import * as vscode from "vscode";
import { format } from "../language/formatter";

export class GolrFormattingProvider implements vscode.DocumentFormattingEditProvider {
  provideDocumentFormattingEdits(
    document: vscode.TextDocument,
    _options: vscode.FormattingOptions,
    _token: vscode.CancellationToken,
  ): vscode.TextEdit[] {
    const original = document.getText();
    const formatted = format(original);

    // Nothing to do if the document is already canonical — returning no edits avoids marking
    // the file dirty unnecessarily.
    if (formatted === original) return [];

    // Replace the full range [start of document, end of document) with the formatted text.
    const fullRange = new vscode.Range(
      document.positionAt(0),
      document.positionAt(original.length),
    );
    return [vscode.TextEdit.replace(fullRange, formatted)];
  }
}
