// The JavaScript runner for the scanner half of the corpus. It reads an input file, runs the generated scanner over it
// and prints the canonical scanner trace, which the harness diffs against the trace the Go reference interpreter
// produced for the same case.
//
// It sees nothing but the generated module and the input.
//
// This file has no dependencies, and it must not grow any. It runs in the official node image with no network, which is
// what proves that generated GoLR code needs nothing but the bare language.

import { readFileSync } from "node:fs";

import { Scanner, Token, tokenToString } from "./scanner.js";

// The range of bytes a trace line carries as they are. Everything outside of it is escaped, which keeps a trace free of
// control characters and of anything a target language could interpret as an encoding of its own.
const printableLow = 0x20;
const printableHigh = 0x7e;

// escapeLexeme escapes the bytes of a lexeme for a trace line. The caller writes the quotes around the result, so that
// the template literal of an event shows the whole shape of the line.
//
// The lexeme is a Uint8Array and is escaped byte by byte, never rune by rune, so a multi byte UTF-8 sequence becomes one
// \xHH escape per byte. That is the whole point of the exercise on this language: a JavaScript string is a sequence of
// UTF-16 code units, so decoding the lexeme first would report code unit offsets and would disagree with every other
// backend.
//
// JSON.stringify is not an alternative, and neither is any other built in quoter. It escapes a different set, writes
// \u001b where the trace asks for \x1b, and it would have to be handed a decoded string in the first place.
function escapeLexeme(lexeme) {
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

// writeLine prints one event of the trace, terminated by the LF the canonical form asks for. Every event goes out as it
// is produced rather than into a buffer, so a scanner which throws half way still leaves behind how far it got.
//
// process.stdout.write and not console.log: console.log substitutes %s and friends, which would rewrite a trace line
// carrying a lexeme that happens to contain one.
function writeLine(line) {
    process.stdout.write(line + "\n");
}

const inputPath = process.argv[2];

// readFileSync without an encoding returns a Buffer, which is a Uint8Array and therefore exactly what the generated
// scanner wants. Handing it a string would be the classic mistake this whole harness exists to catch.
const source = readFileSync(inputPath);

// The plain Scanner and not the TokenSkipper: a skipped rule matched like any other and belongs in the trace, because
// the byte offsets of the tokens around it are only checkable when it is there.
const scanner = new Scanner(source, inputPath);

while (scanner.next()) {
    const token = scanner.token();

    // The start of the lexeme, which for a failed match is the start of the attempt and not the byte which could not be
    // consumed.
    const start = scanner.byteOffset();

    if (token === Token.InvalidToken) {
        writeLine(`ERROR ${start}`);
        continue;
    }

    // The end comes from the length of the lexeme rather than from an offset the scanner reports, so that a lexeme
    // which disagrees with the offsets cannot pass unnoticed.
    const lexeme = scanner.lexeme();
    writeLine(`TOKEN ${tokenToString(token)} ${start} ${start + lexeme.length} "${escapeLexeme(lexeme)}"`);
}

// The offset the scanner reports after it ran out of input, not source.length. The two agree only when the scanner
// consumed everything, which is what this line is here to check.
writeLine(`EOF ${scanner.byteOffset()}`);
