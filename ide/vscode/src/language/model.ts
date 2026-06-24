// Symbol model for a GoLR document.
//
// This is the TypeScript counterpart of the IntelliJ plugin's PSI parser
// (ide/intellij/src/main/kotlin/com/backbone81/golr/GolrPsiParser.kt). Where the IntelliJ
// plugin builds a tree of PSI nodes, we build two flat lists — symbol *definitions* and
// symbol *references* — which is all the language features need:
//
//   - Go to Definition: from a reference, find the definitions with the same name.
//   - Find Usages:      from a definition, find the references with the same name.
//   - Rename:           rewrite a definition and every reference with the same name.
//   - Completion:       offer every definition's name.
//   - Semantic tokens:  colour definitions and references differently.
//
// Like the tokenizer, this module is pure (no `vscode` dependency) so it can be unit-tested
// directly. Positions are kept as character offsets into the source string; the providers
// translate offsets to VSCode positions with `document.positionAt`.

import { Token, TokenType, tokenize } from "./tokenizer";

/** A terminal is defined in @scanner; a nonterminal is defined in @parser. */
export type SymbolKind = "terminal" | "nonterminal";

/** The defining occurrence of a symbol — the `NAME` before the `:` in a rule. */
export interface SymbolDefinition {
  name: string;
  kind: SymbolKind;
  /** Offset range of the name itself (not the whole rule). End is exclusive. */
  start: number;
  end: number;
}

/** A use of a symbol — an identifier in a rule body, precedence line, or @start. */
export interface SymbolReference {
  name: string;
  start: number;
  end: number;
}

/** Either kind of occurrence, as returned by {@link GolrModel.symbolAt}. */
export interface SymbolOccurrence {
  name: string;
  start: number;
  end: number;
  isDefinition: boolean;
}

/**
 * The parsed symbol model of a single document. Construct it with {@link buildModel}.
 *
 * Resolution is deliberately file-local and name-based, exactly like the IntelliJ plugin: a
 * reference resolves to every definition in the same file that shares its name (which also
 * tolerates a symbol that is accidentally defined twice).
 */
export class GolrModel {
  constructor(
    readonly definitions: readonly SymbolDefinition[],
    readonly references: readonly SymbolReference[],
  ) {}

  /** All definitions named `name`. */
  definitionsNamed(name: string): SymbolDefinition[] {
    return this.definitions.filter((d) => d.name === name);
  }

  /** All references named `name`. */
  referencesNamed(name: string): SymbolReference[] {
    return this.references.filter((r) => r.name === name);
  }

  /**
   * The symbol occurrence (definition or reference) whose range contains `offset`, or
   * undefined if the offset is not on a symbol. Used to answer "what is under the caret?".
   */
  symbolAt(offset: number): SymbolOccurrence | undefined {
    for (const d of this.definitions) {
      if (offset >= d.start && offset <= d.end) {
        return { name: d.name, start: d.start, end: d.end, isDefinition: true };
      }
    }
    for (const r of this.references) {
      if (offset >= r.start && offset <= r.end) {
        return { name: r.name, start: r.start, end: r.end, isDefinition: false };
      }
    }
    return undefined;
  }
}

/** Tokenizes `text` and parses it into a {@link GolrModel}. */
export function buildModel(text: string): GolrModel {
  // Drop whitespace and comments: the structural parser below only cares about meaningful
  // tokens, mirroring how the IntelliJ PsiBuilder hides those token kinds from the parser.
  const tokens = tokenize(text).filter(
    (t) =>
      t.type !== TokenType.Whitespace &&
      t.type !== TokenType.CommentLine &&
      t.type !== TokenType.CommentBlock,
  );
  return new Parser(tokens).parse();
}

// A small recursive-descent walk over the significant token stream. Each `parseX` method
// advances `this.pos` past the construct it consumes. The structure intentionally follows
// GolrPsiParser.kt method-for-method so the two stay easy to compare.
class Parser {
  private pos = 0;
  private readonly definitions: SymbolDefinition[] = [];
  private readonly references: SymbolReference[] = [];

  constructor(private readonly tokens: Token[]) {}

  parse(): GolrModel {
    while (!this.eof()) {
      const t = this.current();
      if (t.type === TokenType.KeywordSection && t.text === "@scanner") {
        this.parseSection("terminal");
      } else if (t.type === TokenType.KeywordSection && t.text === "@parser") {
        this.parseSection("nonterminal");
      } else {
        // Skip stray top-level tokens (error recovery).
        this.advance();
      }
    }
    return new GolrModel(this.definitions, this.references);
  }

  // @scanner { rules }   or   @parser { rules }
  //
  // `kind` is the kind that rule definitions in this section produce. Terminals never have
  // identifier references in their bodies, so the only structural difference between the two
  // sections is the @start / @precedence directives that may appear in @parser.
  private parseSection(kind: SymbolKind): void {
    this.advance(); // consume @scanner / @parser
    this.expect(TokenType.LBrace);

    while (!this.eof() && this.current().type !== TokenType.RBrace) {
      const t = this.current();
      if (kind === "nonterminal" && t.type === TokenType.KeywordControl && t.text === "@start") {
        this.parseStartDeclaration();
      } else if (
        kind === "nonterminal" &&
        t.type === TokenType.KeywordControl &&
        t.text === "@precedence"
      ) {
        this.parsePrecedenceBlock();
      } else if (t.type === TokenType.Identifier) {
        this.parseRule(kind);
      } else {
        this.advance(); // error recovery
      }
    }

    this.expectIf(TokenType.RBrace);
  }

  // NAME : body ;
  // The leading NAME is a definition. In @parser bodies, every identifier is a reference; in
  // @scanner bodies there are none (so the body loop simply finds nothing to record).
  private parseRule(kind: SymbolKind): void {
    const nameTok = this.current();
    this.definitions.push({
      name: nameTok.text,
      kind,
      start: nameTok.start,
      end: nameTok.end,
    });
    this.advance(); // consume NAME
    this.expectIf(TokenType.Colon);

    while (
      !this.eof() &&
      this.current().type !== TokenType.Semicolon &&
      this.current().type !== TokenType.RBrace
    ) {
      const t = this.current();
      if (t.type === TokenType.Identifier) {
        this.addReference(t);
        this.advance();
      } else if (t.type === TokenType.KeywordControl && t.text === "@precedence") {
        // Inline @precedence(NAME): the NAME inside the parentheses is a reference.
        this.advance(); // consume @precedence
        this.expectIf(TokenType.LParen);
        if (!this.eof() && this.current().type === TokenType.Identifier) {
          this.addReference(this.current());
          this.advance();
        }
        this.expectIf(TokenType.RParen);
      } else {
        this.advance(); // strings, @empty, "|", etc.
      }
    }

    this.expectIf(TokenType.Semicolon);
  }

  // @start : NAME ;   — NAME is a reference to the grammar's start nonterminal.
  private parseStartDeclaration(): void {
    this.advance(); // consume @start
    this.expectIf(TokenType.Colon);
    if (!this.eof() && this.current().type === TokenType.Identifier) {
      this.addReference(this.current());
      this.advance();
    }
    this.expectIf(TokenType.Semicolon);
  }

  // @precedence { lines }   — a container; each line is a precedence directive.
  private parsePrecedenceBlock(): void {
    this.advance(); // consume @precedence
    this.expectIf(TokenType.LBrace);
    while (!this.eof() && this.current().type !== TokenType.RBrace) {
      if (this.current().type === TokenType.KeywordControl) {
        this.parsePrecedenceLine();
      } else {
        this.advance(); // error recovery
      }
    }
    this.expectIf(TokenType.RBrace);
  }

  // @left : SYM SYM ;   (also @right / @none) — every identifier is a reference.
  private parsePrecedenceLine(): void {
    this.advance(); // consume @left / @right / @none
    this.expectIf(TokenType.Colon);
    while (!this.eof() && this.current().type !== TokenType.Semicolon) {
      const t = this.current();
      if (t.type === TokenType.Identifier) {
        this.addReference(t);
      }
      this.advance();
    }
    this.expectIf(TokenType.Semicolon);
  }

  // ── cursor helpers ─────────────────────────────────────────────────────────────────────

  private addReference(t: Token): void {
    this.references.push({ name: t.text, start: t.start, end: t.end });
  }

  private current(): Token {
    return this.tokens[this.pos];
  }

  private eof(): boolean {
    return this.pos >= this.tokens.length;
  }

  private advance(): void {
    this.pos++;
  }

  // Consume the current token if it matches `type`; otherwise leave the cursor in place
  // (lenient, matching the IntelliJ parser which does not emit error markers).
  private expectIf(type: TokenType): void {
    if (!this.eof() && this.current().type === type) this.advance();
  }

  // Same as expectIf — named `expect` at the section entry point purely for readability.
  private expect(type: TokenType): void {
    this.expectIf(type);
  }
}
