package com.backbone81.golr

// Structural port of internal/utils/heredoc.go. Lets GolrFormatterTest's fixtures be written as
// Kotlin raw strings indented to match the surrounding code, the same way the Go tests write a
// Go raw string literal wrapped in utils.HereDoc, instead of as concatenated string literals.
// Deliberately not Kotlin's own trimIndent(): that also drops the first and last line if they are
// blank, which would silently swallow the trailing newline a fixture needs when its closing
// """ sits on its own line - exactly the shape hereDoc (like HereDoc) is meant to preserve.
fun hereDoc(s: String): String {
    val trimmed = s.removePrefix("\n")
    val lines = trimmed.split("\n").toMutableList()

    val indent = commonIndent(lines)
    for (i in lines.indices) {
        val line = lines[i]
        lines[i] = if (line.isBlank()) {
            // Force blank lines to be fully empty instead of relying on removePrefix below: a
            // blank line may be shorter than indent (e.g. an editor stripped its trailing
            // whitespace), in which case removePrefix would leave its leftover whitespace in
            // place rather than stripping it.
            ""
        } else {
            line.removePrefix(indent)
        }
    }
    return lines.joinToString("\n")
}

// Returns the longest leading whitespace prefix shared by every non-blank line.
private fun commonIndent(lines: List<String>): String {
    var indent = ""
    var found = false
    for (line in lines) {
        if (line.isBlank()) {
            // Blank lines carry no indentation information and must not narrow the common
            // prefix: a short or empty blank line would otherwise collapse indent to less than
            // what the actual content agrees on.
            continue
        }
        val lineIndent = line.takeWhile { it == ' ' || it == '\t' }
        if (!found) {
            // Seed indent with the first non-blank line's own indentation. Without this, indent
            // would start as "" and commonPrefix("", lineIndent) is always "", so no line could
            // ever establish an initial value.
            indent = lineIndent
            found = true
            continue
        }
        indent = commonPrefix(indent, lineIndent)
    }
    return indent
}

// Returns the longest prefix shared by a and b.
private fun commonPrefix(a: String, b: String): String {
    var i = 0
    while (i < a.length && i < b.length && a[i] == b[i]) {
        i++
    }
    return a.substring(0, i)
}
