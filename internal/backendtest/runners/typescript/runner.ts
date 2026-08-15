// The TypeScript runner for the backend test corpus. It reads an input file, runs the generated scanner and the
// generated parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. What little of node it needs is declared in node.d.ts. It
// runs in the image with no network, which is what proves that generated GoLR code needs nothing but the bare language.

import { readFileSync, writeFileSync } from "node:fs";

import { Scanner, Token, TokenSkipper, tokenToString } from "./scanner.js";
import { Nonterminal, Parser, ParseSymbol, nonterminalToString } from "./parser.js";
import type { ParseNode } from "./parser.js";

const scannerTraceFileName = "scanner.actual";
const parserTraceFileName = "parser.actual";

// The bytes a trace line carries as they are. Everything outside of it is escaped.
const printableLow = 0x20;
const printableHigh = 0x7e;

// escapeLexeme escapes the bytes of a lexeme. The caller writes the quotes around the result.
//
// The lexeme is a Uint8Array and is escaped byte by byte, never rune by rune, so a multi byte UTF-8 sequence becomes one
// \xHH escape per byte. Decoding it first would report UTF-16 code unit offsets and disagree with every other backend.
// JSON.stringify is no alternative: it escapes a different set and would need a decoded string to begin with.
function escapeLexeme(lexeme: Uint8Array): string {
    let result = "";
    for (const value of lexeme) {
        switch (value) {
            case 0x5c: // backslash
                result += "\\\\";
                continue;
            case 0x22: // quote
                result += "\\\"";
                continue;
            case 0x0a:
                result += "\\n";
                continue;
            case 0x0d:
                result += "\\r";
                continue;
            case 0x09:
                result += "\\t";
                continue;
        }

        if (printableLow <= value && value <= printableHigh) {
            result += String.fromCharCode(value);
            continue;
        }

        // toString(16) is lower case, and the trace asks for a zero padded pair of lower case hex digits.
        result += "\\x" + value.toString(16).padStart(2, "0");
    }
    return result;
}

// appendScannerTrace scans the whole input and appends one line per event.
function appendScannerTrace(lines: string[], source: Uint8Array, inputPath: string): void {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the offsets of the tokens
    // around it are only checkable when it is in the trace.
    const scanner = new Scanner(source, inputPath);

    while (scanner.next()) {
        const token = scanner.token();

        // For a failed match this is the start of the attempt, not the byte which could not be consumed.
        const start = scanner.byteOffset();

        if (token === Token.InvalidToken) {
            lines.push(`ERROR ${start}`);
            continue;
        }

        // The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a lexeme
        // which disagrees with the offsets cannot pass unnoticed.
        const lexeme = scanner.lexeme();
        lines.push(`TOKEN ${tokenToString(token)} ${start} ${start + lexeme.length} "${escapeLexeme(lexeme)}"`);
    }

    // The offset the scanner reports after it ran out of input, not source.length. The two agree only when the scanner
    // consumed everything.
    lines.push(`EOF ${scanner.byteOffset()}`);
}

// appendNodeTrace appends the post-order walk of the subtree, which is the shift and reduce sequence of the parse.
function appendNodeTrace(lines: string[], node: ParseNode): void {
    for (const child of node.children) {
        appendNodeTrace(lines, child);
    }

    if (ParseSymbol.isNonterminal(node.symbol)) {
        const nonterminal = ParseSymbol.value(node.symbol) as Nonterminal;
        lines.push(`REDUCE ${nonterminalToString(nonterminal)} ${node.children.length}`);
        return;
    }

    const token = ParseSymbol.value(node.symbol) as Token;
    if (token === Token.ErrorToken) {
        // The leaf the recovery pushed where it resumed. It stands for the dropped input and names no token.
        lines.push("RESYNC");
        return;
    }
    lines.push(`SHIFT ${tokenToString(token)}`);
}

// appendParserTrace parses the whole input and appends its errors, the walk of the tree and the accept event.
//
// The errors come first because a shift carries no offset to interleave them by. A parse which was given up returns no
// tree, so its trace is the errors alone and the missing accept event is what says the two outcomes apart.
function appendParserTrace(lines: string[], source: Uint8Array, inputPath: string): void {
    // The TokenSkipper here, because a skipped rule never reaches the parser.
    const { tree, errors } = new Parser().parse(new TokenSkipper(new Scanner(source, inputPath)));

    for (const error of errors) {
        lines.push(`ERROR ${error.byteOffset}`);
    }

    if (tree === null) {
        return;
    }
    appendNodeTrace(lines, tree);
    lines.push("ACCEPT");
}

// writeTrace produces one trace and writes it to its file. Whatever was produced before a throw is written all the
// same, so a runner which breaks half way still says how far it got, and the other trace is still produced.
function writeTrace(fileName: string, produce: (lines: string[]) => void): void {
    const lines: string[] = [];
    try {
        produce(lines);
    } catch (error) {
        const detail = error instanceof Error ? (error.stack ?? error.message) : String(error);
        process.stderr.write(`producing ${fileName} failed: ${detail}\n`);
    }

    // Every line is terminated, and an empty trace is an empty file rather than a bare newline.
    writeFileSync(fileName, lines.map((line) => line + "\n").join(""));
}

const inputPath = process.argv[2];
if (inputPath === undefined) {
    throw new Error("the runner takes the input file as its only argument");
}

// readFileSync without an encoding returns the bytes, which is what the generated scanner wants. Handing it a string
// would be the classic mistake this harness exists to catch.
const source = readFileSync(inputPath);

writeTrace(scannerTraceFileName, (lines) => appendScannerTrace(lines, source, inputPath));
writeTrace(parserTraceFileName, (lines) => appendParserTrace(lines, source, inputPath));
