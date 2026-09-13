package com.backbone81.golr

import java.nio.charset.StandardCharsets

// Pure, platform-independent reformatter for .golr files. It turns arbitrary (possibly messy)
// GoLR source into the canonical layout demonstrated by examples/golang/spec/golang.golr:
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
// The actual algorithm lives in Formatter, a structural port of internal/fmt/formatter.go. This
// object plays the role internal/fmt/golr.go's GoLRString plays for the CLI: encode to bytes, run
// the formatter, decode back to a string. It is kept independent of IntelliJ's block-based
// formatting engine so it can be unit-tested directly. GolrPreFormatProcessor wires it into the
// IDE's "Reformat Code" action.
object GolrFormatter {
    fun format(text: String): String {
        val formatted = Formatter().format(text.toByteArray(StandardCharsets.UTF_8), "in-memory")
        return String(formatted, StandardCharsets.UTF_8)
    }
}
