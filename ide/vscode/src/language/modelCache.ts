// Per-document cache of the parsed GoLR symbol model.
//
// Every language feature (definition, references, rename, completion, semantic tokens) needs
// the symbol model of the document it is acting on. Rebuilding it from scratch on every
// keystroke and every provider call would be wasteful, so this cache stores the most recently
// built model for each document and reuses it until the document changes.
//
// VSCode gives every open document a monotonically increasing `version` number that bumps on
// each edit, so we key the cache on (document URI, version): a cache hit means the document
// has not changed since we last parsed it.

import * as vscode from "vscode";
import { GolrModel, buildModel } from "./model";

interface CacheEntry {
  version: number;
  model: GolrModel;
}

export class ModelCache {
  private readonly entries = new Map<string, CacheEntry>();

  /** Returns the symbol model for `document`, parsing it only if the cached copy is stale. */
  get(document: vscode.TextDocument): GolrModel {
    const key = document.uri.toString();
    const cached = this.entries.get(key);
    if (cached && cached.version === document.version) {
      return cached.model;
    }
    const model = buildModel(document.getText());
    this.entries.set(key, { version: document.version, model });
    return model;
  }

  /** Drops the cached model for a document (called when a document is closed). */
  invalidate(uri: vscode.Uri): void {
    this.entries.delete(uri.toString());
  }
}
