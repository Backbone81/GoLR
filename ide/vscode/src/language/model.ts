// Symbol model for a GoLR document.
//
// Instead of a full syntax tree, we build flat lists of symbol *definitions*, the string
// *aliases* of terminals, and symbol *references* — which is all the language features need:
//
//   - Go to Definition: from a reference, find the definitions (or aliases) with the same name.
//   - Find Usages:      from a definition, find the references by its name or its alias.
//   - Rename:           rewrite a definition (or alias) and every reference with the same name.
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

/**
 * The string of a terminal like `PLUS: "+";`, which productions and precedence lines may use
 * in place of the terminal's name. The name includes the quotes, as that is the text of every
 * reference to it.
 */
export interface AliasDefinition {
  name: string;
  /** Name of the terminal this is the alias of. */
  terminal: string;
  start: number;
  end: number;
}

/**
 * A use of a symbol — an identifier or string in a rule body or precedence line, or the
 * identifier of @start.
 */
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
 * Resolution is deliberately file-local and name-based: a reference resolves to every
 * definition in the same file that shares its name (which also tolerates a symbol that is
 * accidentally defined twice).
 */
export class GolrModel {
  constructor(
    readonly definitions: readonly SymbolDefinition[],
    readonly aliases: readonly AliasDefinition[],
    readonly references: readonly SymbolReference[],
  ) {}

  /** All definitions named `name`. */
  definitionsNamed(name: string): SymbolDefinition[] {
    return this.definitions.filter((d) => d.name === name);
  }

  /** The definitions a reference named `name` resolves to: aliases for a string, else symbols. */
  declarationsNamed(name: string): (SymbolDefinition | AliasDefinition)[] {
    return isAliasName(name)
      ? this.aliases.filter((a) => a.name === name)
      : this.definitionsNamed(name);
  }

  /** All references named `name`. */
  referencesNamed(name: string): SymbolReference[] {
    return this.references.filter((r) => r.name === name);
  }

  /** The usages of the symbol named `name`: references by that name or by its aliases. */
  usagesNamed(name: string): SymbolReference[] {
    const names = new Set([name]);
    for (const a of this.aliases) {
      if (a.terminal === name) names.add(a.name);
    }
    return this.references.filter((r) => names.has(r.name));
  }

  /**
   * The symbol occurrence (definition or reference) whose range contains `offset`, or
   * undefined if the offset is not on a symbol. Used to answer "what is under the caret?".
   */
  symbolAt(offset: number): SymbolOccurrence | undefined {
    for (const d of [...this.definitions, ...this.aliases]) {
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

/** Returns true when `name` is a string alias like `"+"` rather than a symbol name. */
export function isAliasName(name: string): boolean {
  return name.startsWith('"');
}

/** Tokenizes `text` and parses it into a {@link GolrModel}. */
export function buildModel(text: string): GolrModel {
  // Drop whitespace and comments: the structural parser below only cares about meaningful
  // tokens.
  const tokens = tokenize(text).filter(
    (t) =>
      t.type !== TokenType.Whitespace &&
      t.type !== TokenType.CommentLine &&
      t.type !== TokenType.CommentBlock,
  );
  return new Parser(tokens).parse();
}

// A small recursive-descent walk over the significant token stream. Each `parseX` method
// advances `this.pos` past the construct it consumes.
class Parser {
  private pos = 0;
  private readonly definitions: SymbolDefinition[] = [];
  private readonly aliases: AliasDefinition[] = [];
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
    return new GolrModel(this.definitions, this.aliases, this.references);
  }

  // @scanner { rules }   or   @parser { rules }
  //
  // `kind` is the kind that rule definitions in this section produce. Only @parser has the
  // @start / @precedence directives.
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
      } else if (t.type === TokenType.Identifier && kind === "terminal") {
        this.parseScannerRule();
      } else if (t.type === TokenType.Identifier) {
        this.parseParserRule();
      } else {
        this.advance(); // error recovery
      }
    }

    this.expectIf(TokenType.RBrace);
  }

  // NAME : /regex/ @fragment? ;   or   NAME : "string" @skip? ;   or   NAME : @empty ;
  // The leading NAME defines a terminal, and its string (if any) is the terminal's alias. A
  // @fragment has no alias, and a scanner body contains no references.
  private parseScannerRule(): void {
    const name = this.parseDefinition("terminal");

    const bodyStart = this.pos;
    while (!this.atRuleEnd()) this.advance();
    const body = this.tokens.slice(bodyStart, this.pos);

    const isFragment = body.some(
      (t) => t.type === TokenType.KeywordControl && t.text === "@fragment",
    );
    const alias = body.find((t) => t.type === TokenType.String);
    if (alias && !isFragment) {
      this.aliases.push({ name: alias.text, terminal: name, start: alias.start, end: alias.end });
    }

    this.expectIf(TokenType.Semicolon);
  }

  // NAME : body ;
  // The leading NAME defines a nonterminal; every identifier and string in the body is a
  // reference, a string referring to a terminal by its alias.
  private parseParserRule(): void {
    this.parseDefinition("nonterminal");

    while (!this.atRuleEnd()) {
      const t = this.current();
      if (isSymbolToken(t)) {
        this.addReference(t);
        this.advance();
      } else if (t.type === TokenType.KeywordControl && t.text === "@precedence") {
        // Inline @precedence(SYMBOL): the SYMBOL inside the parentheses is a reference.
        this.advance(); // consume @precedence
        this.expectIf(TokenType.LParen);
        if (!this.eof() && isSymbolToken(this.current())) {
          this.addReference(this.current());
          this.advance();
        }
        this.expectIf(TokenType.RParen);
      } else if (t.type === TokenType.KeywordControl && t.text === "@name") {
        // Inline @name(NAME): the NAME is the production's own name, not a symbol
        // reference, so there is nothing to resolve it to.
        this.advance(); // consume @name
        this.expectIf(TokenType.LParen);
        if (!this.eof() && this.current().type === TokenType.Identifier) {
          this.advance();
        }
        this.expectIf(TokenType.RParen);
      } else {
        // "|", @empty, @error, etc. @error is a symbol in the body, but it is built in
        // rather than declared anywhere, so there is nothing to resolve it to.
        this.advance();
      }
    }

    this.expectIf(TokenType.Semicolon);
  }

  // Records the rule NAME at the cursor as a definition of `kind` and consumes `NAME :`.
  private parseDefinition(kind: SymbolKind): string {
    const nameTok = this.current();
    this.definitions.push({ name: nameTok.text, kind, start: nameTok.start, end: nameTok.end });
    this.advance(); // consume NAME
    this.expectIf(TokenType.Colon);
    return nameTok.text;
  }

  private atRuleEnd(): boolean {
    return (
      this.eof() ||
      this.current().type === TokenType.Semicolon ||
      this.current().type === TokenType.RBrace
    );
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

  // @left : SYM "+" ;   (also @right / @none) — every identifier and string is a reference.
  private parsePrecedenceLine(): void {
    this.advance(); // consume @left / @right / @none
    this.expectIf(TokenType.Colon);
    while (!this.eof() && this.current().type !== TokenType.Semicolon) {
      const t = this.current();
      if (isSymbolToken(t)) {
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
  // (lenient: no error markers are emitted).
  private expectIf(type: TokenType): void {
    if (!this.eof() && this.current().type === type) this.advance();
  }

  // Same as expectIf — named `expect` at the section entry point purely for readability.
  private expect(type: TokenType): void {
    this.expectIf(type);
  }
}

// An identifier references a symbol by name, a string references a terminal by its alias.
function isSymbolToken(t: Token): boolean {
  return t.type === TokenType.Identifier || t.type === TokenType.String;
}
