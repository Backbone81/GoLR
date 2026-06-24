// Canonical reformatter for GoLR files.
//
// This is a TypeScript port of the IntelliJ plugin's formatter
// (ide/intellij/src/main/kotlin/com/backbone81/golr/GolrFormatter.kt). It rewrites arbitrary
// (possibly messy) GoLR source into the canonical layout demonstrated by
// examples/golang/spec/golang.golr:
//
//   @scanner {
//       horizontal_whitespace: /[ \t]/ @fragment;     // bodies column-aligned within a group
//       vertical_whitespace:   /[\r\n]/ @fragment;
//   }
//
//   @parser {
//       @start: SourceFiles;                           // control directives stay single-line
//
//       Rule                                           // parser rule name on its own line
//           : alternative one                          // one alternative per line, ":"/"|"-led
//           | alternative two
//           ;
//   }
//
// Like the IntelliJ version, the logic is a pure function over text so it can be unit-tested
// directly and stays independent of the editor. The formatting provider
// (src/providers/formattingProvider.ts) wires it into VSCode's "Format Document" action.

import { TokenType, tokenize } from "./tokenizer";

const INDENT = "    ";

/** Reformats `text` into canonical GoLR layout. Returns "" for empty/whitespace-only input. */
export function format(text: string): string {
  const tokens = lex(text);
  if (tokens.length === 0) return "";

  const out = new Output();
  let i = 0;
  let first = true;
  while (i < tokens.length) {
    const token = tokens[i];
    if (!first && token.blankBefore) out.blank();
    first = false;
    if (isComment(token)) {
      out.line(0, token.text);
      i += 1;
    } else if (token.type === TokenType.KeywordSection) {
      i = emitSection(tokens, i, out);
    } else {
      i += 1; // stray top-level token — skip
    }
  }
  return out.build();
}

// ── tokenization ─────────────────────────────────────────────────────────────────────────

// A significant token (whitespace dropped) plus whether the whitespace preceding it contained
// a blank line (>= 2 newlines), which drives blank-line preservation.
interface Tok {
  type: TokenType;
  text: string;
  blankBefore: boolean;
}

function isComment(t: Tok): boolean {
  return t.type === TokenType.CommentLine || t.type === TokenType.CommentBlock;
}

function lex(text: string): Tok[] {
  const result: Tok[] = [];
  let newlines = 0;
  for (const token of tokenize(text)) {
    if (token.type === TokenType.Whitespace) {
      newlines += countNewlines(token.text);
    } else {
      result.push({ type: token.type, text: token.text.replace(/\s+$/, ""), blankBefore: newlines >= 2 });
      newlines = 0;
    }
  }
  return result;
}

function countNewlines(s: string): number {
  let n = 0;
  for (const ch of s) if (ch === "\n") n++;
  return n;
}

// ── item model ───────────────────────────────────────────────────────────────────────────

// A section body is parsed into a flat list of these before emission, so scanner rules can be
// grouped for column alignment and blank lines can be reasoned about per item.
type Item = CommentItem | RuleItem | BlockItem;

interface CommentItem {
  kind: "comment";
  text: string;
  blankBefore: boolean;
}

// A rule or directive ending in ';'. tokens holds everything up to (but excluding) the ';'.
interface RuleItem {
  kind: "rule";
  tokens: Tok[];
  blankBefore: boolean;
}

// A nested "@precedence { ... }" block.
interface BlockItem {
  kind: "block";
  inner: Item[];
  blankBefore: boolean;
  closingBraceBlankBefore: boolean;
}

// Parses items until the RBRACE that closes the enclosing block. Returns the items and the
// index of that RBRACE (not consumed), or the end index if none is found.
function parseItems(tokens: Tok[], start: number): { items: Item[]; end: number } {
  const items: Item[] = [];
  let i = start;
  while (i < tokens.length) {
    const token = tokens[i];
    if (token.type === TokenType.RBrace) {
      return { items, end: i };
    }
    if (isComment(token)) {
      items.push({ kind: "comment", text: token.text, blankBefore: token.blankBefore });
      i++;
      continue;
    }
    if (
      token.type === TokenType.KeywordControl &&
      token.text === "@precedence" &&
      tokens[i + 1]?.type === TokenType.LBrace
    ) {
      const blankBefore = token.blankBefore;
      const inner = parseItems(tokens, i + 2);
      let next = inner.end;
      let braceBlank = false;
      if (next < tokens.length && tokens[next].type === TokenType.RBrace) {
        braceBlank = tokens[next].blankBefore;
        next++;
      }
      items.push({ kind: "block", inner: inner.items, blankBefore, closingBraceBlankBefore: braceBlank });
      i = next;
      continue;
    }
    // Otherwise: a rule/directive running up to the next ';' (or '}').
    const ruleTokens: Tok[] = [];
    const blankBefore = token.blankBefore;
    while (
      i < tokens.length &&
      tokens[i].type !== TokenType.Semicolon &&
      tokens[i].type !== TokenType.RBrace
    ) {
      ruleTokens.push(tokens[i]);
      i++;
    }
    if (i < tokens.length && tokens[i].type === TokenType.Semicolon) i++;
    if (ruleTokens.length > 0) items.push({ kind: "rule", tokens: ruleTokens, blankBefore });
  }
  return { items, end: i };
}

// ── emission ─────────────────────────────────────────────────────────────────────────────

function emitSection(tokens: Tok[], start: number, out: Output): number {
  const keyword = tokens[start].text; // @scanner or @parser
  out.line(0, `${keyword} {`);

  let i = start + 1;
  if (i < tokens.length && tokens[i].type === TokenType.LBrace) i++;

  const { items, end } = parseItems(tokens, i);
  i = end;

  if (keyword === "@scanner") emitScannerItems(items, out, 1);
  else emitParserItems(items, out, 1);

  const brace = tokens[i];
  if (brace && brace.blankBefore) out.blank();
  out.line(0, "}");
  return brace && brace.type === TokenType.RBrace ? i + 1 : i;
}

function emitScannerItems(items: Item[], out: Output, depth: number): void {
  // Assign each rule a group id: a maximal run of directly consecutive rules (no blank line
  // and no comment between them). Bodies are column-aligned within a group.
  const groupOf = new Map<RuleItem, number>();
  let groupId = -1;
  let prevWasRule = false;
  for (const item of items) {
    if (item.kind === "rule") {
      if (!prevWasRule || item.blankBefore) groupId++;
      groupOf.set(item, groupId);
      prevWasRule = true;
    } else {
      prevWasRule = false;
    }
  }
  const groupWidth = new Map<number, number>();
  for (const item of items) {
    if (item.kind === "rule") {
      const width = nameOf(item).length + 1; // "+ 1" for the colon
      const g = groupOf.get(item)!;
      groupWidth.set(g, Math.max(groupWidth.get(g) ?? 0, width));
    }
  }

  let first = true;
  for (const item of items) {
    if (!first && item.blankBefore) out.blank();
    first = false;
    if (item.kind === "comment") {
      out.line(depth, item.text);
    } else if (item.kind === "block") {
      emitPrecedenceBlock(item, out, depth);
    } else {
      const prefix = padEnd(`${nameOf(item)}:`, groupWidth.get(groupOf.get(item)!)!);
      const body = joinSymbols(bodyOf(item));
      out.line(depth, body.length === 0 ? `${prefix};` : `${prefix} ${body};`);
    }
  }
}

function emitParserItems(items: Item[], out: Output, depth: number): void {
  let first = true;
  for (const item of items) {
    if (!first && item.blankBefore) out.blank();
    first = false;
    if (item.kind === "comment") {
      out.line(depth, item.text);
    } else if (item.kind === "block") {
      emitPrecedenceBlock(item, out, depth);
    } else if (item.tokens[0].type === TokenType.KeywordControl) {
      emitDirective(item, out, depth);
    } else {
      emitParserRule(item, out, depth);
    }
  }
}

function emitParserRule(item: RuleItem, out: Output, depth: number): void {
  out.line(depth, nameOf(item));
  const alternatives = splitByPipe(bodyOf(item));
  alternatives.forEach((alternative, index) => {
    const lead = index === 0 ? ":" : "|";
    const body = joinSymbols(alternative);
    out.line(depth + 1, body.length === 0 ? lead : `${lead} ${body}`);
  });
  out.line(depth + 1, ";");
}

// A single-line directive such as "@start: SourceFiles;" or "@left: \"+\" \"-\";".
function emitDirective(item: RuleItem, out: Output, depth: number): void {
  const keyword = item.tokens[0].text;
  const body = joinSymbols(bodyOf(item));
  out.line(depth, body.length === 0 ? `${keyword}:;` : `${keyword}: ${body};`);
}

function emitPrecedenceBlock(block: BlockItem, out: Output, depth: number): void {
  out.line(depth, "@precedence {");
  let first = true;
  for (const item of block.inner) {
    if (!first && item.blankBefore) out.blank();
    first = false;
    if (item.kind === "comment") {
      out.line(depth + 1, item.text);
    } else if (item.kind === "rule") {
      emitDirective(item, out, depth + 1);
    } else {
      emitPrecedenceBlock(item, out, depth + 1);
    }
  }
  if (block.closingBraceBlankBefore) out.blank();
  out.line(depth, "}");
}

// ── helpers ──────────────────────────────────────────────────────────────────────────────

function nameOf(item: RuleItem): string {
  return item.tokens[0].text;
}

// The tokens after the ':' (the rule/directive body).
function bodyOf(item: RuleItem): Tok[] {
  const colon = item.tokens.findIndex((t) => t.type === TokenType.Colon);
  return colon >= 0 ? item.tokens.slice(colon + 1) : item.tokens.slice(1);
}

function splitByPipe(tokens: Tok[]): Tok[][] {
  const result: Tok[][] = [[]];
  for (const token of tokens) {
    if (token.type === TokenType.Pipe) {
      result.push([]);
    } else {
      result[result.length - 1].push(token);
    }
  }
  return result;
}

// Joins body symbols with single spaces, keeping an inline "@precedence(NAME)" annotation
// tight (no spaces around its parentheses).
function joinSymbols(tokens: Tok[]): string {
  let builder = "";
  let i = 0;
  while (i < tokens.length) {
    const token = tokens[i];
    if (
      token.type === TokenType.KeywordControl &&
      token.text === "@precedence" &&
      tokens[i + 1]?.type === TokenType.LParen
    ) {
      const inner = tokens[i + 2]?.text ?? "";
      if (builder.length > 0) builder += " ";
      builder += `@precedence(${inner})`;
      i += 4; // @precedence ( NAME )
    } else {
      if (builder.length > 0) builder += " ";
      builder += token.text;
      i++;
    }
  }
  return builder;
}

function padEnd(s: string, width: number): string {
  return s.length >= width ? s : s + " ".repeat(width - s.length);
}

// Accumulates output lines, normalizing blank lines (no leading or doubled blanks) and
// guaranteeing exactly one trailing newline.
class Output {
  private readonly lines: string[] = [];

  line(depth: number, text: string): void {
    this.lines.push(text.length === 0 ? "" : INDENT.repeat(depth) + text);
  }

  blank(): void {
    if (this.lines.length > 0 && this.lines[this.lines.length - 1].length !== 0) {
      this.lines.push("");
    }
  }

  build(): string {
    return this.lines.join("\n").replace(/\n+$/, "") + "\n";
  }
}
