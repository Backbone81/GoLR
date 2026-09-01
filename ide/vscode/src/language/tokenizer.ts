// Tokenizer for the GoLR grammar language.
//
// A thin adapter over the GoLR-generated DFA scanner (generated/scanner.ts, from
// internal/parsergen/frontend/golr/spec/golr.golr via `make generate`). Pure (string in,
// array out) and free of the `vscode` module, so it unit-tests in plain Node.

import { Scanner, Token as ScannerToken } from "./generated/scanner";

/** The kinds of tokens GoLR source can contain. */
export enum TokenType {
  Whitespace = "whitespace",
  CommentLine = "comment-line",
  CommentBlock = "comment-block",
  /** A /.../ regular-expression literal (scanner rule body). */
  Regex = "regex",
  /** A "..." string literal (inline terminal). */
  String = "string",
  /** @scanner or @parser — the two top-level section markers. */
  KeywordSection = "keyword-section",
  /** @skip, @fragment, @empty, @error, @start, @left, @right, @none, @precedence. */
  KeywordControl = "keyword-control",
  /** A bare name: [A-Za-z_][A-Za-z0-9_]*. */
  Identifier = "identifier",
  Colon = "colon",
  Semicolon = "semicolon",
  Pipe = "pipe",
  LBrace = "lbrace",
  RBrace = "rbrace",
  LParen = "lparen",
  RParen = "rparen",
  /** Any byte that begins no token (a stray character, an unterminated literal, a bare `,`). */
  BadCharacter = "bad-character",
}

/**
 * A single token. `start` and `end` are offsets into the original source string (end is
 * exclusive, like a substring range), so callers can map a token straight back to a position
 * in the document.
 */
export interface Token {
  type: TokenType;
  start: number;
  end: number;
  text: string;
}

/**
 * Splits GoLR source into a list of tokens, including whitespace and comments (callers that
 * do not care about those filter them out). The scan never throws: any byte the automaton
 * cannot start a token with becomes a BadCharacter token and the scan moves on.
 */
export function tokenize(text: string): Token[] {
  const bytes = new TextEncoder().encode(text);
  const byteToChar = buildByteToChar(text, bytes.length);

  const scanner = new Scanner(bytes, "");
  const tokens: Token[] = [];
  while (scanner.next()) {
    const startByte = scanner.byteOffset();
    const endByte = startByte + scanner.lexeme().length;
    const start = byteToChar[startByte]!;
    const end = byteToChar[endByte]!;
    tokens.push({
      type: mapToken(scanner.token(), bytes, startByte),
      start,
      end,
      text: text.substring(start, end),
    });
  }
  return tokens;
}

// Maps a generated scanner token to the extension's TokenType. `bytes`/`startByte` are used to
// split a COMMENT into line vs. block by its opening two bytes.
function mapToken(token: ScannerToken, bytes: Uint8Array, startByte: number): TokenType {
  switch (token) {
    case ScannerToken.TokenWhitespace:
      return TokenType.Whitespace;
    case ScannerToken.TokenComment:
      return bytes[startByte] === SLASH && bytes[startByte + 1] === SLASH
        ? TokenType.CommentLine
        : TokenType.CommentBlock;
    case ScannerToken.TokenScanner:
    case ScannerToken.TokenParser:
      return TokenType.KeywordSection;
    case ScannerToken.TokenPrecedence:
    case ScannerToken.TokenName:
    case ScannerToken.TokenStart:
    case ScannerToken.TokenLeft:
    case ScannerToken.TokenRight:
    case ScannerToken.TokenNone:
    case ScannerToken.TokenSkip:
    case ScannerToken.TokenEmpty:
    case ScannerToken.TokenError:
    case ScannerToken.TokenFragment:
      return TokenType.KeywordControl;
    case ScannerToken.TokenLbrace:
      return TokenType.LBrace;
    case ScannerToken.TokenRbrace:
      return TokenType.RBrace;
    case ScannerToken.TokenLparen:
      return TokenType.LParen;
    case ScannerToken.TokenRparen:
      return TokenType.RParen;
    case ScannerToken.TokenColon:
      return TokenType.Colon;
    case ScannerToken.TokenSemi:
      return TokenType.Semicolon;
    case ScannerToken.TokenPipe:
      return TokenType.Pipe;
    case ScannerToken.TokenIdentifier:
      return TokenType.Identifier;
    case ScannerToken.TokenRegex:
      return TokenType.Regex;
    case ScannerToken.TokenString:
      return TokenType.String;
    // COMMA is defined but unused in the grammar; InvalidToken is any byte that begins no
    // token. Both are errors, as is anything else that reaches here.
    default:
      return TokenType.BadCharacter;
  }
}

// Maps each UTF-8 byte index back to the UTF-16 offset of its character, so scanner byte
// offsets become string offsets. One extra slot so the end offset maps too. ASCII input gives
// the identity map; multi-byte content (Unicode in char classes or comments) still maps right.
function buildByteToChar(text: string, byteLength: number): number[] {
  const map = new Array<number>(byteLength + 1);
  let charIdx = 0;
  let byteIdx = 0;
  for (const ch of text) {
    const byteCount = utf8ByteCount(ch.codePointAt(0)!);
    for (let k = 0; k < byteCount; k++) {
      map[byteIdx + k] = charIdx;
    }
    charIdx += ch.length;
    byteIdx += byteCount;
  }
  map[byteLength] = charIdx;
  return map;
}

function utf8ByteCount(codePoint: number): number {
  if (codePoint < 0x80) return 1;
  if (codePoint < 0x800) return 2;
  if (codePoint < 0x10000) return 3;
  return 4;
}

const SLASH = "/".charCodeAt(0);
