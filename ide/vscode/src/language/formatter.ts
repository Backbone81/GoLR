// Structural port of internal/fmt/formatter.go. Kept as close as possible to that file's shape
// (same fields, same onTokenX methods, same two-pass structure) so a fix made there has a
// same-named counterpart here to apply it to. The only intentional deviations are things
// TypeScript/JavaScript itself dictates:
//
//   - currentContext() returns Token | undefined instead of using an "invalid token" sentinel,
//     since that's the idiomatic way to express "no value" here.
//   - There is no built-in growable byte buffer (unlike Go's bytes.Buffer), so a small local
//     ByteBuffer class fills that role.
//
// Drives the generated Scanner (generated/scanner.ts) directly, the same way formatter.go drives
// the Go scanner directly, rather than going through tokenizer.ts's TokenType mapping - that
// mapping collapses distinctions (e.g. @scanner/@parser both become KeywordSection) this
// formatter needs kept apart (onTokenScanner vs onTokenParser vs onTokenStart vs
// onTokenPrecedence are separate handlers).
//
// format(), at the bottom of this file, plays the role internal/fmt/golr.go's GoLRString plays
// for the CLI: encode to bytes, run the formatter, decode back to a string. It rewrites arbitrary
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
// The formatting provider (src/providers/formattingProvider.ts) wires format() into VSCode's
// "Format Document" action.

import { Config, DefaultConfig } from "./config";
import { Scanner, Token } from "./generated/scanner";

const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder();

const NEWLINE_BYTE = 0x0a; // '\n'
const SLASH_BYTE = 0x2f; // '/'

// Notes down where a scanner rule's right hand side starts, so a second pass can align it with
// the other rules in the same group.
interface ScannerRuleRHS {
  /** The byte position in the output right after the rule's ":", where padding is inserted. */
  offset: number;
  /** The length of the rule's identifier. */
  ruleNameLen: number;
}

// Grows a byte buffer the way Go's bytes.Buffer does for Format's own use: write() appends,
// length reports the byte count so far (used for scanner-rule-column marks), bytes() returns the
// accumulated content. TypeScript has no built-in equivalent (Uint8Array/ArrayBuffer are fixed
// size): this collects the written chunks and concatenates them once, in bytes(), rather than
// copying on every write.
class ByteBuffer {
  #chunks: Uint8Array[] = [];
  #len = 0;

  write(data: Uint8Array): void {
    this.#chunks.push(data);
    this.#len += data.length;
  }

  get length(): number {
    return this.#len;
  }

  bytes(): Uint8Array {
    const result = new Uint8Array(this.#len);
    let offset = 0;
    for (const chunk of this.#chunks) {
      result.set(chunk, offset);
      offset += chunk.length;
    }
    return result;
  }
}

export class Formatter {
  #config: Config;

  #scanner!: Scanner;
  #indentLevel = 0;

  // The length of the most recently emitted identifier inside a @scanner section, i.e. the left
  // hand side of the scanner rule whose ":" comes next.
  #lastScannerIdentifierLen = 0;

  // Collects scanner rule marks seen during the first pass, grouped by contiguous runs of rules
  // not separated by an explicit blank line. It always has at least one (possibly empty) group;
  // a new one is started as soon as an explicit blank line is seen inside a @scanner section, and
  // it stays empty if no rule follows before the next one. Consumed by alignScannerRules
  // afterwards.
  #scannerRuleGroups: ScannerRuleRHS[][] = [];

  // Reports if the next output needs to be indented or not.
  #indentNext = false;

  // Reports if a linebreak was requested but not yet written to the output. Requesting a
  // linebreak is lazy so that a same-line trailing comment can still be emitted before it.
  #pendingLinebreak = false;

  // Reports if a blank line before the next emitted content has already been scheduled, so a
  // second, independent reason to want one (e.g. automatic spacing between rules and a user's own
  // blank line around a comment coinciding) does not stack into two. Cleared once real content is
  // emitted.
  #pendingBlankLine = false;

  // Reports if the next emit should not write a whitespace to separate the previous token from
  // the next one.
  #emitTight = false;

  // A stack which describes the current nesting.
  #context: Token[] = [];

  // Reports if the whitespace just consumed contained a blank line, i.e. the user separated the
  // surrounding tokens by an empty line of their own.
  #explicitBlankLine = false;

  // Reports if the whitespace just consumed contained a linebreak, i.e. the next token starts on
  // a new source line instead of trailing the previous token.
  #explicitNewline = false;

  // Reports if the token just emitted was a comment, so a following explicit blank line can be
  // attributed to it instead of to whatever token comes after the whitespace.
  #explicitComment = false;

  #output!: ByteBuffer;

  constructor(config: Config = DefaultConfig) {
    this.#config = config;
  }

  format(source: Uint8Array, filePath: string): Uint8Array {
    this.#indentLevel = 0;
    this.#indentNext = true;
    this.#pendingLinebreak = false;
    this.#pendingBlankLine = false;
    this.#emitTight = false;
    this.#context = [];
    this.#explicitBlankLine = false;
    this.#explicitNewline = false;
    this.#explicitComment = false;
    this.#lastScannerIdentifierLen = 0;
    this.#scannerRuleGroups = [[]];

    this.#output = new ByteBuffer();
    this.#scanner = new Scanner(source, filePath);
    while (this.#scanner.next()) {
      // explicitComment must be reset before every token, not just at the bottom of the loop
      // like explicitBlankLine/explicitNewline, so it reflects only the token from the
      // immediately preceding iteration; TokenWhitespace below still needs that old value, so
      // grab a copy first.
      const explicitCommentBefore = this.#explicitComment;
      this.#explicitComment = false;

      switch (this.#scanner.token()) {
        case Token.TokenWhitespace:
          this.#onTokenWhitespace(explicitCommentBefore);
          continue;
        case Token.TokenComment:
          this.#onTokenComment();
          break;
        case Token.TokenColon:
          this.#onTokenColon();
          break;
        case Token.TokenPipe:
          this.#onTokenPipe();
          break;
        case Token.TokenSemi:
          this.#onTokenSemi();
          break;
        case Token.TokenLbrace:
          this.#onTokenLbrace();
          break;
        case Token.TokenRbrace:
          this.#onTokenRbrace();
          break;
        case Token.TokenLparen:
          this.#onTokenLparen();
          break;
        case Token.TokenRparen:
          this.#onTokenRparen();
          break;
        case Token.TokenIdentifier:
          this.#onTokenIdentifier();
          break;
        case Token.TokenScanner:
          this.#onTokenScanner();
          break;
        case Token.TokenParser:
          this.#onTokenParser();
          break;
        case Token.TokenStart:
          this.#onTokenStart();
          break;
        case Token.TokenPrecedence:
          this.#onTokenPrecedence();
          break;
        default:
          this.#emit(this.#scanner.lexeme());
          break;
      }

      this.#explicitBlankLine = false;
      this.#explicitNewline = false;
      // explicitComment is intentionally not reset here; see the reset at the top of the loop.
    }
    this.#ensureTrailingNewline();
    return this.#alignScannerRules(this.#output.bytes());
  }

  // Guarantees that non-empty output ends with exactly one "\n", regardless of which token ended
  // the input (a well-formed "}" leaves a blank line pending that would otherwise contribute a
  // second, unwanted trailing "\n", while a file that just stops mid-rule leaves nothing pending
  // at all).
  #ensureTrailingNewline(): void {
    const bytes = this.#output.bytes();
    if (bytes.length === 0) {
      return;
    }
    if (bytes[bytes.length - 1] !== NEWLINE_BYTE) {
      this.#output.write(textEncoder.encode("\n"));
    }
  }

  #onTokenWhitespace(explicitCommentBefore: boolean): void {
    // We are looking for linebreaks and explicit blank lines by the user.
    let newlines = 0;
    for (const b of this.#scanner.lexeme()) if (b === NEWLINE_BYTE) newlines++;
    if (newlines >= 1) {
      this.#explicitNewline = true;
    }
    if (newlines >= 2) {
      this.#explicitBlankLine = true;
      if (explicitCommentBefore) {
        // The user separated the comment we just emitted from what follows with a blank line;
        // keep it instead of collapsing the comment onto the next token.
        this.#blankLine();
      }
      if (this.#currentContext() === Token.TokenScanner) {
        // Start a new alignment group. It stays empty if no rule follows before the next blank
        // line, which is harmless.
        this.#scannerRuleGroups.push([]);
      }
    }
  }

  #onTokenComment(): void {
    const lexeme = this.#scanner.lexeme();
    const isLineComment = lexeme.length >= 2 && lexeme[0] === SLASH_BYTE && lexeme[1] === SLASH_BYTE;
    if (this.#explicitNewline) {
      // The comment starts on its own line rather than trailing the previous token.
      if (this.#explicitBlankLine) {
        // The user separated the comment from the previous content with a blank line; keep it.
        this.#blankLine();
      } else {
        this.#linebreak();
      }
      this.#emit(lexeme);
    } else if (this.#pendingLinebreak) {
      // The comment trails the previous token, but a linebreak (and possibly a blank line) is
      // already scheduled to run after that token. Emit the comment before that instead of
      // flushing it early. indentNext is forced false for the emit: it is only true here because
      // the scheduled linebreak hasn't fired yet, and leaving it true would make emit treat the
      // comment as opening a fresh indented line (writing the indent string as a separator)
      // instead of trailing the current one.
      const hadBlankLine = this.#pendingBlankLine;
      this.#pendingLinebreak = false;
      this.#pendingBlankLine = false;
      this.#indentNext = false;
      this.#emit(lexeme);
      this.#pendingLinebreak = true;
      this.#pendingBlankLine = hadBlankLine;
      this.#indentNext = true;
    } else {
      this.#emit(lexeme);
    }
    if (isLineComment || this.#explicitNewline) {
      // Anything after "//" on the same line would otherwise be swallowed into the comment. A
      // comment that started on its own line must not have the following content glued onto it
      // either, even for a block comment which doesn't force this lexically.
      this.#linebreak();
    }
    // Read by the top of the loop on the next token, to detect a blank line right after this
    // comment.
    this.#explicitComment = true;
  }

  #onTokenColon(): void {
    switch (this.#currentContext()) {
      case Token.TokenScanner: {
        // Note down where the rule's right hand side starts, so the second pass can align it.
        const mark: ScannerRuleRHS = {
          offset: this.#output.length + 1,
          ruleNameLen: this.#lastScannerIdentifierLen,
        };
        this.#scannerRuleGroups[this.#scannerRuleGroups.length - 1].push(mark);

        this.#emitTight = true;
        this.#emit(this.#scanner.lexeme());
        break;
      }
      case Token.TokenParser:
        // Indent the alternatives under the rule name; matching indentDec runs at the rule's ";".
        this.#indentInc();
        this.#linebreak();
        this.#emit(this.#scanner.lexeme());
        break;
      default:
        this.#emitTight = true;
        this.#emit(this.#scanner.lexeme());
        break;
    }
  }

  #onTokenPipe(): void {
    this.#linebreak();
    this.#emit(this.#scanner.lexeme());
  }

  #onTokenSemi(): void {
    switch (this.#currentContext()) {
      case Token.TokenParser:
        this.#linebreak();
        this.#emit(this.#scanner.lexeme());
        this.#indentDec();
        // Separate top-level rules with a blank line; idempotent, so it composes with a blank
        // line a trailing comment on the next rule also wants (see onTokenComment).
        this.#blankLine();
        break;
      case Token.TokenStart:
        // "@start: X;" stays on one line rather than being spread out like a regular rule.
        this.#emitTight = true;
        this.#emit(this.#scanner.lexeme());
        this.#popContext();
        this.#blankLine();
        break;
      default:
        this.#emitTight = true;
        this.#emit(this.#scanner.lexeme());
        this.#linebreak();
        break;
    }
  }

  #onTokenLbrace(): void {
    this.#emit(this.#scanner.lexeme());
    this.#linebreak();
    this.#indentInc();
  }

  #onTokenRbrace(): void {
    this.#indentLevel = Math.max(this.#indentLevel - 1, 0);
    // A blank line may still be pending to separate the previous rule from a following one (see
    // onTokenSemi); it must not leak into a blank line before the closing brace itself.
    this.#pendingBlankLine = false;
    if (!this.#indentNext) {
      // Skip if we're already at the start of a fresh line (e.g. an empty "{}" block), otherwise
      // this would add a spurious blank line before "}".
      this.#linebreak();
    }
    this.#emit(this.#scanner.lexeme());
    this.#blankLine();

    // We remove any context we did push onto the context stack.
    this.#popContext();
  }

  #onTokenLparen(): void {
    if (this.#currentContext() === Token.TokenPrecedence) {
      // "@precedence" is ambiguous: "@precedence { ... }" opens a block, but "@precedence(...)"
      // is an inline alternative annotation with no matching "}". onTokenPrecedence below pushes
      // eagerly as if it were the block form; seeing a "(" instead of a "{" proves it wasn't, so
      // undo it.
      this.#popContext();
    }

    // Left parenthesis always sit close to the previous token.
    this.#emitTight = true;
    this.#emit(this.#scanner.lexeme());

    // The following token also needs to sit close to the left parenthesis.
    this.#emitTight = true;
  }

  #onTokenRparen(): void {
    // Right parenthesis always sit close to the previous token.
    this.#emitTight = true;
    this.#emit(this.#scanner.lexeme());
  }

  #onTokenIdentifier(): void {
    if (this.#currentContext() === Token.TokenScanner) {
      // This identifier is the left hand side of a scanner rule; remember its length for the ":"
      // that follows right after.
      this.#lastScannerIdentifierLen = this.#scanner.byteLength();
      if (this.#explicitBlankLine) {
        // Preserve the user's blank line between scanner rules.
        this.#blankLine();
      }
    }
    this.#emit(this.#scanner.lexeme());
  }

  #onTokenScanner(): void {
    // Track that we're inside @scanner; popped again on the matching "}".
    this.#emit(this.#scanner.lexeme());
    this.#pushContext(Token.TokenScanner);
  }

  #onTokenParser(): void {
    // Track that we're inside @parser; popped again on the matching "}".
    this.#emit(this.#scanner.lexeme());
    this.#pushContext(Token.TokenParser);
  }

  #onTokenStart(): void {
    // Track that we're inside "@start ... ;"; popped again on the matching ";" so the colon and
    // semicolon are formatted inline instead of like a regular rule's.
    this.#emit(this.#scanner.lexeme());
    this.#pushContext(Token.TokenStart);
  }

  #onTokenPrecedence(): void {
    // Pushed eagerly as if opening a "@precedence { ... }" block; onTokenLparen above repairs
    // this if it turns out to be the inline "@precedence(...)" annotation instead.
    this.#emit(this.#scanner.lexeme());
    this.#pushContext(Token.TokenPrecedence);
  }

  #indentInc(): void {
    this.#indentLevel++;
  }

  #indentDec(): void {
    this.#indentLevel = Math.max(this.#indentLevel - 1, 0);
  }

  // Requests a linebreak before the next emitted content. Like blankLine, it is lazy (the actual
  // "\n" is written by emit, so that a same-line trailing comment can still be inserted before
  // it) and idempotent, so a second, independent reason to want one (e.g. a trailing comment
  // already scheduling one, followed by a token that unconditionally wants one too) does not
  // stack into an unwanted blank line.
  #linebreak(): void {
    this.#pendingLinebreak = true;
    this.#indentNext = true;
  }

  // Requests a blank line before the next emitted content. Like linebreak, it is lazy: the actual
  // "\n\n" is written by emit. It is idempotent, so a second, independent reason to want a blank
  // line (e.g. automatic spacing between rules and a user's own blank line around a comment
  // coinciding) does not stack into two.
  #blankLine(): void {
    this.#pendingLinebreak = true;
    this.#pendingBlankLine = true;
    this.#indentNext = true;
  }

  // Flushes any pending linebreak or blank line, indents if needed, and writes data.
  #emit(data: Uint8Array): void {
    if (this.#pendingBlankLine) {
      // A pending blank line implies a pending linebreak too (see blankLine), so check it first.
      this.#output.write(textEncoder.encode("\n\n"));
    } else if (this.#pendingLinebreak) {
      this.#output.write(textEncoder.encode("\n"));
    }
    this.#pendingLinebreak = false;
    this.#pendingBlankLine = false;
    if (this.#indentNext) {
      for (let i = 0; i < this.#indentLevel; i++) {
        this.#output.write(textEncoder.encode(this.#config.indentation));
      }
      this.#emitTight = true;
    }
    if (!this.#emitTight) {
      this.#output.write(textEncoder.encode(" "));
    }
    this.#indentNext = false;
    this.#emitTight = false;
    this.#output.write(data);
  }

  #pushContext(token: Token): void {
    this.#context.push(token);
  }

  #popContext(): void {
    this.#context.pop();
  }

  #currentContext(): Token | undefined {
    return this.#context[this.#context.length - 1];
  }

  // The second pass over the pretty printed output. It pads the ":" of scanner rules within each
  // group (a run of rules not separated by an explicit blank line) so their right hand sides
  // start in the same column.
  #alignScannerRules(data: Uint8Array): Uint8Array {
    const result = new ByteBuffer();
    let lastOffset = 0;
    for (const group of this.#scannerRuleGroups) {
      let maxPrefixLen = 0;
      for (const mark of group) {
        maxPrefixLen = Math.max(maxPrefixLen, mark.ruleNameLen);
      }

      for (const mark of group) {
        result.write(data.subarray(lastOffset, mark.offset));
        for (let i = 0; i < maxPrefixLen - mark.ruleNameLen; i++) {
          result.write(textEncoder.encode(" "));
        }
        lastOffset = mark.offset;
      }
    }
    result.write(data.subarray(lastOffset));
    return result.bytes();
  }
}

/** Reformats `text` into canonical GoLR layout. Returns "" for empty/whitespace-only input. */
export function format(text: string): string {
  const formatted = new Formatter().format(textEncoder.encode(text), "in-memory");
  return textDecoder.decode(formatted);
}
