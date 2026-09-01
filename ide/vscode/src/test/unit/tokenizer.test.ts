// Unit tests for the tokenizer. These run in plain Node via Mocha (no VSCode host).

import * as assert from "assert";
import { TokenType, tokenize, Token } from "../../language/tokenizer";

// Helper: tokenize and drop whitespace so assertions focus on meaningful tokens.
function significant(text: string): Token[] {
  return tokenize(text).filter((t) => t.type !== TokenType.Whitespace);
}

// Helper: the (type, text) pairs of the significant tokens.
function pairs(text: string): Array<[TokenType, string]> {
  return significant(text).map((t) => [t.type, t.text]);
}

suite("tokenizer", () => {
  test("classifies punctuation", () => {
    assert.deepStrictEqual(pairs(":;|{}()"), [
      [TokenType.Colon, ":"],
      [TokenType.Semicolon, ";"],
      [TokenType.Pipe, "|"],
      [TokenType.LBrace, "{"],
      [TokenType.RBrace, "}"],
      [TokenType.LParen, "("],
      [TokenType.RParen, ")"],
    ]);
  });

  test("classifies identifiers", () => {
    assert.deepStrictEqual(pairs("expression term_1 _hidden"), [
      [TokenType.Identifier, "expression"],
      [TokenType.Identifier, "term_1"],
      [TokenType.Identifier, "_hidden"],
    ]);
  });

  test("distinguishes section and control keywords; an unknown @word does not form a token", () => {
    assert.deepStrictEqual(pairs("@scanner @parser @left @bogus"), [
      [TokenType.KeywordSection, "@scanner"],
      [TokenType.KeywordSection, "@parser"],
      [TokenType.KeywordControl, "@left"],
      // "@bogus" matches no keyword: "@" is a stray byte, "bogus" a plain name.
      [TokenType.BadCharacter, "@"],
      [TokenType.Identifier, "bogus"],
    ]);
  });

  test("@error in a rule body is a control keyword", () => {
    assert.deepStrictEqual(pairs('stmt : @error ";"'), [
      [TokenType.Identifier, "stmt"],
      [TokenType.Colon, ":"],
      [TokenType.KeywordControl, "@error"],
      [TokenType.String, '";"'],
    ]);
  });

  test("line comment runs to end of line, not past the newline", () => {
    const tokens = significant("// hello\nx");
    assert.strictEqual(tokens[0].type, TokenType.CommentLine);
    assert.strictEqual(tokens[0].text, "// hello");
    assert.strictEqual(tokens[1].type, TokenType.Identifier);
  });

  test("block comment spans multiple lines", () => {
    const tokens = significant("/* a\n b */ x");
    assert.strictEqual(tokens[0].type, TokenType.CommentBlock);
    assert.strictEqual(tokens[0].text, "/* a\n b */");
    assert.strictEqual(tokens[1].type, TokenType.Identifier);
  });

  test("unterminated block comment does not become a comment token", () => {
    // No closing "*/": matches no comment rule, so the whole run is one BadCharacter token.
    const tokens = significant("/* never closed");
    assert.deepStrictEqual(
      tokens.map((t) => [t.type, t.text]),
      [[TokenType.BadCharacter, "/* never closed"]],
    );
  });

  test("regex literal honours escaped slashes and ends at the closing slash", () => {
    const tokens = significant("/[0-9]\\/+/ x");
    assert.strictEqual(tokens[0].type, TokenType.Regex);
    assert.strictEqual(tokens[0].text, "/[0-9]\\/+/");
    assert.strictEqual(tokens[1].type, TokenType.Identifier);
  });

  test("string literal honours escaped quotes", () => {
    const tokens = significant('"a\\"b" x');
    assert.strictEqual(tokens[0].type, TokenType.String);
    assert.strictEqual(tokens[0].text, '"a\\"b"');
    assert.strictEqual(tokens[1].type, TokenType.Identifier);
  });

  test("an unterminated regex does not run past the end of the line", () => {
    // REGEX/STRING are single-line, so a lone "/" before a newline is a stray byte.
    const tokens = significant("/\nNAME : /x/ ;");
    assert.strictEqual(tokens[0].type, TokenType.BadCharacter);
    assert.strictEqual(tokens[0].text, "/");
    assert.strictEqual(tokens[1].type, TokenType.Identifier);
    assert.strictEqual(tokens[1].text, "NAME");
  });

  test("an unterminated string does not run past the end of the line", () => {
    // '"oops' matches no string rule; the run stops at the newline and the next line is fine.
    const tokens = significant('BAR : "oops\nBAZ : "x" ;');
    assert.strictEqual(tokens[0].text, "BAR");
    assert.strictEqual(tokens[2].type, TokenType.BadCharacter);
    assert.strictEqual(tokens[2].text, '"oops');
    assert.strictEqual(tokens[3].text, "BAZ");
    assert.strictEqual(tokens[4].text, ":");
    assert.strictEqual(tokens[5].type, TokenType.String);
    assert.strictEqual(tokens[5].text, '"x"');
  });

  test("token offsets map back to the source substring", () => {
    const text = "ab : cd ;";
    for (const t of tokenize(text)) {
      assert.strictEqual(text.substring(t.start, t.end), t.text);
    }
  });

  test("empty input yields no tokens", () => {
    assert.deepStrictEqual(tokenize(""), []);
  });
});
