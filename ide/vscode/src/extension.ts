// Entry point of the GoLR VSCode extension.
//
// VSCode loads this module and calls `activate` the first time a GoLR file is opened (see
// "activationEvents" in package.json). Activation wires up every language feature by
// registering a provider for each one against the `golr` language id. When the extension is
// unloaded, VSCode calls `deactivate`.
//
// All providers share a single ModelCache so that opening, say, the rename box and then
// asking for references does not reparse the document twice.

import * as vscode from "vscode";
import { ModelCache } from "./language/modelCache";
import { GolrDefinitionProvider } from "./providers/definitionProvider";
import { GolrReferenceProvider } from "./providers/referenceProvider";
import { GolrRenameProvider } from "./providers/renameProvider";
import { GolrCompletionProvider } from "./providers/completionProvider";
import {
  GolrSemanticTokensProvider,
  golrSemanticTokensLegend,
} from "./providers/semanticTokensProvider";
import { GolrFormattingProvider } from "./providers/formattingProvider";
import { GolrCodeLensProvider } from "./providers/codeLensProvider";

// The document selector that scopes every provider to GoLR files only. The language id `golr`
// is declared in package.json's contributes.languages.
const GOLR_SELECTOR: vscode.DocumentSelector = { language: "golr" };

export function activate(context: vscode.ExtensionContext): void {
  const cache = new ModelCache();

  // Each registration returns a Disposable. Pushing them onto context.subscriptions lets
  // VSCode tear them down automatically when the extension deactivates.
  context.subscriptions.push(
    vscode.languages.registerDefinitionProvider(GOLR_SELECTOR, new GolrDefinitionProvider(cache)),

    vscode.languages.registerReferenceProvider(GOLR_SELECTOR, new GolrReferenceProvider(cache)),

    vscode.languages.registerRenameProvider(GOLR_SELECTOR, new GolrRenameProvider(cache)),

    vscode.languages.registerCompletionItemProvider(GOLR_SELECTOR, new GolrCompletionProvider(cache)),

    vscode.languages.registerDocumentSemanticTokensProvider(
      GOLR_SELECTOR,
      new GolrSemanticTokensProvider(cache),
      golrSemanticTokensLegend,
    ),

    vscode.languages.registerDocumentFormattingEditProvider(
      GOLR_SELECTOR,
      new GolrFormattingProvider(),
    ),

    vscode.languages.registerCodeLensProvider(GOLR_SELECTOR, new GolrCodeLensProvider(cache)),

    // Drop a document's cached model once it is closed so the cache does not grow unbounded.
    vscode.workspace.onDidCloseTextDocument((doc) => cache.invalidate(doc.uri)),
  );
}

export function deactivate(): void {
  // Nothing to clean up manually: everything we registered was added to
  // context.subscriptions, which VSCode disposes for us.
}
