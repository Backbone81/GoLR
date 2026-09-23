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
//      (an identifier, or a string literal when renaming a terminal's string alias) and return
//      a WorkspaceEdit that rewrites the definition together with every reference, so the
//      whole file stays consistent. Renaming a terminal keeps its string, and renaming a
//      string keeps the terminal name.

import * as vscode from "vscode";
import { isAliasName } from "../language/model";
import { ModelCache } from "../language/modelCache";
import { TokenType, tokenize } from "../language/tokenizer";

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
    const model = this.cache.get(document);
    const occurrence = model.symbolAt(document.offsetAt(position));
    if (!occurrence) return undefined;

    if (isAliasName(occurrence.name)) {
      if (!isStringLiteral(newName)) {
        throw new Error(`'${newName}' is not a string literal.`);
      }
    } else if (!IDENTIFIER_RE.test(newName)) {
      throw new Error(`'${newName}' is not a valid GoLR identifier.`);
    }

    const edit = new vscode.WorkspaceEdit();
    const rename = (start: number, end: number): void => {
      edit.replace(
        document.uri,
        new vscode.Range(document.positionAt(start), document.positionAt(end)),
        newName,
      );
    };

    // Rewrite every definition and every reference that shares the old name. Renaming from a
    // reference therefore updates the definition too.
    for (const def of model.declarationsNamed(occurrence.name)) rename(def.start, def.end);
    for (const ref of model.referencesNamed(occurrence.name)) rename(ref.start, ref.end);

    return edit;
  }
}

// The new name of a string alias must lex as exactly one string token.
function isStringLiteral(text: string): boolean {
  const tokens = tokenize(text);
  return tokens.length === 1 && tokens[0].type === TokenType.String && tokens[0].end === text.length;
}
