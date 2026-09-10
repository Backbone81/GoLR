// The TypeScript runner for the backend test corpus. It reads an input file, runs the generated scanner and the
// generated parser over it, and writes the canonical scanner trace and parser trace the harness diffs against.
//
// This file has no dependencies, and it must not grow any. What little of node it needs is declared in node.d.ts. It
// runs in the image with no network, which is what proves that generated GoLR code needs nothing but the bare language.

import { readFileSync, writeFileSync } from "node:fs";

import { Scanner, Token, TokenSkipper, tokenToString } from "./scanner.js";
import { Parser } from "./parser.js";

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

// appendScannerTrace scans the whole input and appends one line per event: the position the token or the failed match
// starts at, a keyword, and for a token its rule and lexeme, for a failed match the bytes it could not match.
function appendScannerTrace(lines: string[], source: Uint8Array, inputPath: string): void {
    // The plain Scanner and not the TokenSkipper: a skipped rule matched like any other, and the position of the
    // tokens around it is only checkable when it is in the trace.
    const scanner = new Scanner(source, inputPath);

    while (scanner.next()) {
        const location = `${scanner.line()}:${scanner.column()}`.padEnd(7);
        const lexeme = escapeLexeme(scanner.lexeme());

        if (scanner.token() === Token.InvalidToken) {
            lines.push(`${location} ${"ERROR".padEnd(7)} "${lexeme}"`);
            continue;
        }
        lines.push(`${location} ${"TOKEN".padEnd(7)} ${tokenToString(scanner.token())} "${lexeme}"`);
    }

    // The position after the scanner ran out of input, which is one past the last byte only when it consumed
    // everything.
    const endLocation = `${scanner.line()}:${scanner.column()}`.padEnd(7);
    lines.push(`${endLocation} EOF`);
}

// appendParserTrace parses the whole input and appends the line the parser's trace hook emits for every action. The
// hook reports the error recovery steps too, which the returned tree does not, so the tree and the error are ignored.
function appendParserTrace(lines: string[], source: Uint8Array, inputPath: string): void {
    const parser = new Parser();
    parser.trace = (line) => lines.push(line);

    // The TokenSkipper here, because a skipped rule never reaches the parser.
    parser.parse(new TokenSkipper(new Scanner(source, inputPath)));
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
