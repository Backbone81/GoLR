// Structural port of internal/utils/heredoc.go. Lets formatter.test.ts's fixtures be written as
// indented template literals aligned with the surrounding code, the same way the Go tests write a
// raw string literal wrapped in utils.HereDoc - template literals themselves are already Go raw
// strings' equivalent (no escaping, no interpretation), but unlike Kotlin's trimIndent() they also
// don't auto-dedent, so this fills that one remaining gap.
export function heredoc(s: string): string {
  const trimmed = s.startsWith("\n") ? s.slice(1) : s;
  const lines = trimmed.split("\n");

  const indent = commonIndent(lines);
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (line.trim() === "") {
      // Force blank lines to be fully empty instead of relying on the prefix strip below: a
      // blank line may be shorter than indent (e.g. an editor stripped its trailing whitespace),
      // in which case stripping only a matching prefix would leave its leftover whitespace in
      // place rather than removing it.
      lines[i] = "";
    } else if (line.startsWith(indent)) {
      lines[i] = line.slice(indent.length);
    }
  }
  return lines.join("\n");
}

// Returns the longest leading whitespace prefix shared by every non-blank line.
function commonIndent(lines: string[]): string {
  let indent = "";
  let found = false;
  for (const line of lines) {
    if (line.trim() === "") {
      // Blank lines carry no indentation information and must not narrow the common prefix: a
      // short or empty blank line would otherwise collapse indent to less than what the actual
      // content agrees on.
      continue;
    }
    const match = /^[ \t]*/.exec(line);
    const lineIndent = match ? match[0] : "";
    if (!found) {
      // Seed indent with the first non-blank line's own indentation. Without this, indent would
      // start as "" and commonPrefix("", lineIndent) is always "", so no line could ever
      // establish an initial value.
      indent = lineIndent;
      found = true;
      continue;
    }
    indent = commonPrefix(indent, lineIndent);
  }
  return indent;
}

// Returns the longest prefix shared by a and b.
function commonPrefix(a: string, b: string): string {
  let i = 0;
  while (i < a.length && i < b.length && a[i] === b[i]) {
    i++;
  }
  return a.slice(0, i);
}
