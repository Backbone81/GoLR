// Unit tests for the tokenizer. These run in plain Node via Mocha (no VSCode host), mirroring
// the lexer coverage in the IntelliJ plugin's GolrSyntaxHighlighterTest.

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

  test("distinguishes section and control keywords; unknown @word is a bad character", () => {
    assert.deepStrictEqual(pairs("@scanner @parser @left @bogus"), [
      [TokenType.KeywordSection, "@scanner"],
      [TokenType.KeywordSection, "@parser"],
      [TokenType.KeywordControl, "@left"],
      [TokenType.BadCharacter, "@bogus"],
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

  test("unterminated block comment extends to end of file", () => {
    const tokens = significant("/* never closed");
    assert.strictEqual(tokens.length, 1);
    assert.strictEqual(tokens[0].type, TokenType.CommentBlock);
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

  test("a bare slash before a newline is a regex token, not a comment", () => {
    const tokens = significant("/\n");
    assert.strictEqual(tokens[0].type, TokenType.Regex);
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
