package com.backbone81.golr

import com.intellij.lang.ASTNode
import com.intellij.lang.PsiBuilder
import com.intellij.lang.PsiParser
import com.intellij.psi.tree.IElementType

// Turns the flat token stream from GolrLexer into a structured PSI tree.
//
// The lexer only knows about individual tokens (IDENTIFIER, COLON, REGEX, …).
// Without a parser, the PSI tree is a flat list of tokens with no semantic structure.
//
//   - "Go to Definition" requires GolrSymbolReference nodes that carry a PsiReference.
//   - Rename requires GolrSymbolDefinition nodes that implement PsiNameIdentifierOwner.
//   - Find Usages requires both kinds of nodes so IntelliJ can match references to definitions.
//
// This parser does NOT validate the grammar (it produces no error markers). Its only
// job is to wrap identifier tokens into the correct composite node type based on their
// structural position: definition (left of ":") vs. reference (right of ":", or in
// precedence lines).
//
// IntelliJ passes us a PsiBuilder, which is a cursor over the token stream. Whitespace
// and comment tokens are automatically skipped by the builder (configured in
// GolrParserDefinition.getWhitespaceTokens / getCommentTokens), so the parser only sees
// meaningful tokens.
//
// To create a composite node we:
//   1. Call builder.mark() — returns a Marker that records the start position.
//   2. Consume one or more tokens with builder.advanceLexer().
//   3. Call marker.done(elementType) — the token range between start and now becomes
//      a node of that type in the finished tree.
//
// The finished tree maps to PSI objects via GolrParserDefinition.createElement().
class GolrPsiParser : PsiParser {

    override fun parse(root: IElementType, builder: PsiBuilder): ASTNode {
        // The root marker wraps the entire file. Its type is the FILE element type
        // supplied by GolrParserDefinition.getFileNodeType().
        val fileMarker = builder.mark()

        while (!builder.eof()) {
            when {
                builder.tokenType == GolrTokenTypes.KEYWORD_SECTION && builder.tokenText == "@scanner" ->
                    parseScannerSection(builder)

                builder.tokenType == GolrTokenTypes.KEYWORD_SECTION && builder.tokenText == "@parser" ->
                    parseParserSection(builder)

                // Skip unexpected top-level tokens (e.g. a comment that the builder
                // somehow surfaced, or a malformed file).
                else -> builder.advanceLexer()
            }
        }

        fileMarker.done(root)
        return builder.treeBuilt
    }

    // ── @scanner { rules } ───────────────────────────────────────────────────────────────

    // Consumes the entire @scanner block. Every direct child rule produces a
    // SYMBOL_DEFINITION node with a NAME_ELEMENT for the terminal name.
    // Scanner rule bodies (regex, string, @empty, @skip, @fragment) contain no
    // symbol references — fragment names are embedded inside the REGEX token as
    // "{FragmentName}" and are not standalone IDENTIFIER tokens.
    private fun parseScannerSection(builder: PsiBuilder) {
        advance(builder)                                // consume @scanner
        advanceIf(builder, GolrTokenTypes.LBRACE)      // consume {

        while (!builder.eof() && builder.tokenType != GolrTokenTypes.RBRACE) {
            when (builder.tokenType) {
                GolrTokenTypes.IDENTIFIER -> parseScannerRule(builder)
                else -> builder.advanceLexer()          // error recovery
            }
        }

        advanceIf(builder, GolrTokenTypes.RBRACE)      // consume }
    }

    // Single scanner rule:   NAME : /regex/ @fragment? ;
    //                     or NAME : "string" @skip? ;
    //                     or NAME : @empty ;
    private fun parseScannerRule(builder: PsiBuilder) {
        val ruleMarker = builder.mark()

        // Wrap the defining IDENTIFIER as a NAME_ELEMENT composite node.
        val nameMarker = builder.mark()
        builder.advanceLexer()
        nameMarker.done(GolrElementTypes.NAME_ELEMENT)

        advanceIf(builder, GolrTokenTypes.COLON)

        // Consume the rule body (regex / string / @empty / annotations) until semicolon.
        while (!builder.eof() && builder.tokenType != GolrTokenTypes.SEMICOLON) {
            builder.advanceLexer()
        }
        advanceIf(builder, GolrTokenTypes.SEMICOLON)

        ruleMarker.done(GolrElementTypes.SYMBOL_DEFINITION)
    }

    // ── @parser { [@start] [@precedence {...}] rules } ───────────────────────────────────

    private fun parseParserSection(builder: PsiBuilder) {
        advance(builder)                                // consume @parser
        advanceIf(builder, GolrTokenTypes.LBRACE)      // consume {

        while (!builder.eof() && builder.tokenType != GolrTokenTypes.RBRACE) {
            when {
                // @start : NAME ;  — declares the grammar's start symbol
                builder.tokenType == GolrTokenTypes.KEYWORD_CONTROL && builder.tokenText == "@start" ->
                    parseStartDeclaration(builder)

                // @precedence { @left : SYM ; ... }
                builder.tokenType == GolrTokenTypes.KEYWORD_CONTROL && builder.tokenText == "@precedence" ->
                    parsePrecedenceBlock(builder)

                // NAME : body ;  — a parser rule
                builder.tokenType == GolrTokenTypes.IDENTIFIER ->
                    parseParserRule(builder)

                else -> builder.advanceLexer()          // error recovery
            }
        }

        advanceIf(builder, GolrTokenTypes.RBRACE)      // consume }
    }

    // @start : NAME ;
    // The NAME after the colon is the grammar's start nonterminal. It is a reference,
    // not a definition, so we wrap it as SYMBOL_REFERENCE.
    private fun parseStartDeclaration(builder: PsiBuilder) {
        advance(builder)                                // consume @start
        advanceIf(builder, GolrTokenTypes.COLON)

        if (builder.tokenType == GolrTokenTypes.IDENTIFIER) {
            val refMarker = builder.mark()
            builder.advanceLexer()
            refMarker.done(GolrElementTypes.SYMBOL_REFERENCE)
        }

        advanceIf(builder, GolrTokenTypes.SEMICOLON)
    }

    // @precedence { lines }
    // The outer @precedence block is just a structural container; each line inside it
    // is a PRECEDENCE_DECLARATION produced by parsePrecedenceLine().
    private fun parsePrecedenceBlock(builder: PsiBuilder) {
        advance(builder)                                // consume @precedence
        advanceIf(builder, GolrTokenTypes.LBRACE)      // consume {

        while (!builder.eof() && builder.tokenType != GolrTokenTypes.RBRACE) {
            when (builder.tokenType) {
                GolrTokenTypes.KEYWORD_CONTROL -> parsePrecedenceLine(builder)
                else -> builder.advanceLexer()          // error recovery
            }
        }

        advanceIf(builder, GolrTokenTypes.RBRACE)      // consume }
    }

    // @left : SYMBOL1 SYMBOL2 ;   (also @right, @none, @precedence used as associativity)
    // All identifiers after the colon are references to terminals.
    private fun parsePrecedenceLine(builder: PsiBuilder) {
        val declMarker = builder.mark()

        advance(builder)                                // consume @left / @right / @none
        advanceIf(builder, GolrTokenTypes.COLON)

        while (!builder.eof() && builder.tokenType != GolrTokenTypes.SEMICOLON) {
            if (builder.tokenType == GolrTokenTypes.IDENTIFIER) {
                val refMarker = builder.mark()
                builder.advanceLexer()
                refMarker.done(GolrElementTypes.SYMBOL_REFERENCE)
            } else {
                builder.advanceLexer()                  // consume string literals and other tokens
            }
        }

        advanceIf(builder, GolrTokenTypes.SEMICOLON)
        declMarker.done(GolrElementTypes.PRECEDENCE_DECLARATION)
    }

    // NAME : alternative | alternative ;
    // The NAME before the colon defines the nonterminal; every IDENTIFIER in the body
    // is a reference to a terminal or nonterminal.
    private fun parseParserRule(builder: PsiBuilder) {
        val ruleMarker = builder.mark()

        // Wrap the defining identifier as NAME_ELEMENT.
        val nameMarker = builder.mark()
        builder.advanceLexer()
        nameMarker.done(GolrElementTypes.NAME_ELEMENT)

        advanceIf(builder, GolrTokenTypes.COLON)

        // Parse the production body. We stop at SEMICOLON (end of this rule) or
        // RBRACE (end of the @parser block, for malformed files).
        parseRuleBody(builder)

        advanceIf(builder, GolrTokenTypes.SEMICOLON)
        ruleMarker.done(GolrElementTypes.SYMBOL_DEFINITION)
    }

    // Production body: a sequence of alternatives separated by "|".
    // Each symbol in an alternative is either an IDENTIFIER (reference to a named
    // symbol) or a STRING (reference to an inline terminal like "+" or "(").
    // We only produce SYMBOL_REFERENCE nodes for IDENTIFIER tokens; STRING tokens are
    // left as plain leaves for now.
    //
    // Special case: @precedence(NAME) is an inline annotation that binds a production
    // to a precedence level. The NAME inside the parentheses is a reference.
    private fun parseRuleBody(builder: PsiBuilder) {
        while (!builder.eof()
            && builder.tokenType != GolrTokenTypes.SEMICOLON
            && builder.tokenType != GolrTokenTypes.RBRACE
        ) {
            when {
                builder.tokenType == GolrTokenTypes.IDENTIFIER -> {
                    val refMarker = builder.mark()
                    builder.advanceLexer()
                    refMarker.done(GolrElementTypes.SYMBOL_REFERENCE)
                }

                // @precedence(SYMBOL) — the symbol inside the parens is a reference to a terminal
                builder.tokenType == GolrTokenTypes.KEYWORD_CONTROL && builder.tokenText == "@precedence" -> {
                    advance(builder)                        // consume @precedence
                    advanceIf(builder, GolrTokenTypes.LPAREN)
                    if (builder.tokenType == GolrTokenTypes.IDENTIFIER) {
                        val refMarker = builder.mark()
                        builder.advanceLexer()
                        refMarker.done(GolrElementTypes.SYMBOL_REFERENCE)
                    }
                    advanceIf(builder, GolrTokenTypes.RPAREN)
                }

                // @empty, "|", string literals, and anything else — consume without wrapping
                else -> builder.advanceLexer()
            }
        }
    }

    // ── helpers ──────────────────────────────────────────────────────────────────────────

    // Unconditional advance — named for readability at call sites.
    private fun advance(builder: PsiBuilder) = builder.advanceLexer()

    // Advance only when the current token matches `type`; silently skip if it does not.
    // (A production parser would mark an error here; we keep it simple.)
    private fun advanceIf(builder: PsiBuilder, type: IElementType) {
        if (builder.tokenType == type) builder.advanceLexer()
    }
}
