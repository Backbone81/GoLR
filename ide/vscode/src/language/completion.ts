// Decides what completion offers at a caret: the keywords the grammar allows there, and whether
// symbol names may be referenced there, e.g. @scanner and @parser at the top level or @name and
// @error in a production.
//
// The position is found by lexing the text before the caret rather than from the symbol model,
// as the model does not tell a rule body from a precedence line, and a half-typed keyword like
// "@na" is not a token of its own. Like the tokenizer, this module is pure (no `vscode`
// dependency) so it can be unit-tested directly.

import { TokenType, tokenize } from "./tokenizer";

/** What completion offers at a caret. */
export interface CompletionSite {
  /** Offset where the word being completed starts, including a leading "@". */
  start: number;
  keywords: readonly string[];
  /** Whether the names of the symbols defined in the file are offered. */
  symbols: boolean;
}

// Where the caret is in terms of the GoLR grammar, with the keywords valid there and whether
// symbol names may be referenced there.
interface Position {
  keywords: readonly string[];
  symbols: boolean;
}

const TOP_LEVEL: Position = { keywords: ["@scanner", "@parser"], symbols: false };
const SCANNER_RULE_BODY: Position = { keywords: ["@empty", "@skip", "@fragment"], symbols: false };
const PARSER_STATEMENT_START: Position = { keywords: ["@start", "@precedence"], symbols: false };
const START_BODY: Position = { keywords: [], symbols: true };
const PRECEDENCE_LINE_START: Position = {
  keywords: ["@left", "@right", "@none", "@precedence"],
  symbols: false,
};
const PRECEDENCE_LINE_BODY: Position = { keywords: ["@error"], symbols: true };
const RULE_BODY: Position = { keywords: ["@empty", "@error", "@precedence", "@name"], symbols: true };
const PRECEDENCE_ARGUMENT: Position = { keywords: [], symbols: true };
// Positions where a new name is written or nothing can be completed, e.g. the name of a rule,
// between a rule name and its ":", or inside @name(...).
const OTHER: Position = { keywords: [], symbols: false };

// Tokens whose content is free text, where nothing is completed.
const TEXT_TOKENS = new Set([
  TokenType.CommentLine,
  TokenType.CommentBlock,
  TokenType.String,
  TokenType.Regex,
]);

const IDENTIFIER_CHAR_RE = /[\p{L}\p{Nd}_]/u;

/** What completion offers at `offset` in `text`, or undefined inside a comment, string, or regex. */
export function completionAt(text: string, offset: number): CompletionSite | undefined {
  if (inTextToken(text, offset)) return undefined;

  let start = offset;
  while (start > 0 && IDENTIFIER_CHAR_RE.test(text[start - 1])) start--;
  if (start > 0 && text[start - 1] === "@") start--;

  const position = positionAt(text.slice(0, start));
  return {
    start,
    keywords: position.keywords,
    symbols: position.symbols && text[start] !== "@",
  };
}

// A line comment has no closing delimiter, so a caret at its end is still inside it.
function inTextToken(text: string, offset: number): boolean {
  return tokenize(text).some(
    (t) =>
      TEXT_TOKENS.has(t.type) &&
      t.start < offset &&
      (offset < t.end || (offset === t.end && t.type === TokenType.CommentLine)),
  );
}

// Determines the position at the end of `text`.
function positionAt(text: string): Position {
  let section: string | undefined; // "@scanner" or "@parser" once its "{" is seen
  let pendingSection: string | undefined; // a section keyword waiting for its "{"
  let inPrecedenceBlock = false;
  let statement: string | undefined; // first token of the current statement
  let afterColon = false;
  let parenKeyword: string | undefined; // the keyword before an open "("
  let previous: string | undefined;

  for (const t of tokenize(text)) {
    switch (t.type) {
      case TokenType.Whitespace:
      case TokenType.CommentLine:
      case TokenType.CommentBlock:
        continue;
      case TokenType.KeywordSection:
        if (section === undefined) pendingSection = t.text;
        break;
      case TokenType.LBrace:
        if (section === undefined && pendingSection !== undefined) {
          section = pendingSection;
          pendingSection = undefined;
        } else if (section === "@parser" && statement === "@precedence" && !afterColon) {
          inPrecedenceBlock = true;
          statement = undefined;
        }
        break;
      case TokenType.RBrace:
        if (inPrecedenceBlock) inPrecedenceBlock = false;
        else section = undefined;
        statement = undefined;
        afterColon = false;
        break;
      case TokenType.Semicolon:
        statement = undefined;
        afterColon = false;
        break;
      case TokenType.Colon:
        afterColon = true;
        break;
      case TokenType.LParen:
        parenKeyword = previous;
        break;
      case TokenType.RParen:
        parenKeyword = undefined;
        break;
      default:
        if (statement === undefined) statement = t.text;
    }
    previous = t.text;
  }

  if (section === undefined) return pendingSection === undefined ? TOP_LEVEL : OTHER;
  if (parenKeyword === "@precedence") return PRECEDENCE_ARGUMENT;
  if (parenKeyword !== undefined) return OTHER;
  if (section === "@scanner") return afterColon ? SCANNER_RULE_BODY : OTHER;
  if (inPrecedenceBlock) {
    if (statement === undefined) return PRECEDENCE_LINE_START;
    return afterColon ? PRECEDENCE_LINE_BODY : OTHER;
  }
  if (statement === undefined) return PARSER_STATEMENT_START;
  if (!afterColon) return OTHER;
  return statement === "@start" ? START_BODY : RULE_BODY;
}
