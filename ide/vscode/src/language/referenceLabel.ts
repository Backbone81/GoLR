// Tiny pure helper used by the CodeLens provider to label a symbol's reference count. It lives
// in its own `vscode`-free module so it can be unit-tested in plain Node, just like the
// tokenizer/model/formatter.

/** Formats a reference count as "1 reference" / "N references" (handles the singular). */
export function referenceCountLabel(count: number): string {
  return count === 1 ? "1 reference" : `${count} references`;
}
