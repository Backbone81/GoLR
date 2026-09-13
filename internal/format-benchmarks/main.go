// Command format-benchmarks rewrites go test benchmark tables embedded in fenced code blocks of markdown files so
// that columns line up with a fixed 3-space gap, numbers get thousands separators, and multi-column numbers align on
// the decimal point.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var numberRe = regexp.MustCompile(`^[0-9,]+(\.[0-9]+)?$`)

// benchmarkRow is one "Benchmark..." line, tokenized into whitespace-separated fields.
type benchmarkRow struct {
	lineIdx int
	fields  []string
}

func main() {
	dir := flag.String("dir", "docs", "directory containing the markdown files to reformat")
	flag.Parse()

	if err := run(*dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(dir string) error {
	paths, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return fmt.Errorf("globbing %s: %w", dir, err)
	}

	for _, path := range paths {
		changed, err := processFile(path)
		if err != nil {
			return fmt.Errorf("processing %s: %w", path, err)
		}
		if changed {
			fmt.Println("formatted", path)
		}
	}
	return nil
}

func processFile(path string) (bool, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	formatted := formatMarkdown(string(original))
	if formatted == string(original) {
		return false, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(formatted), info.Mode()); err != nil {
		return false, err
	}
	return true, nil
}

// formatMarkdown reformats every benchmark table found inside fenced code blocks, leaving everything else untouched.
func formatMarkdown(content string) string {
	trailingNewline := strings.HasSuffix(content, "\n")
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")

	var (
		result      []string
		blockStart  = -1
		insideFence = false
	)
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if insideFence {
				result = append(result, formatBenchmarkBlock(lines[blockStart:i])...)
				result = append(result, line)
				insideFence = false
			} else {
				result = append(result, line)
				blockStart = i + 1
				insideFence = true
			}
			continue
		}
		if !insideFence {
			result = append(result, line)
		}
	}
	// An unterminated fence (shouldn't happen in well-formed markdown) just keeps its lines unformatted.
	if insideFence {
		result = append(result, lines[blockStart:]...)
	}

	out := strings.Join(result, "\n")
	if trailingNewline {
		out += "\n"
	}
	return out
}

// formatBenchmarkBlock reformats the go test benchmark lines (those starting with "Benchmark") inside a fenced code
// block. Column widths and numeric alignment are computed across all such lines in the block, even when other lines
// (goos/cpu headers, PASS/FAIL, ok summaries) are interspersed; those other lines are returned unchanged.
func formatBenchmarkBlock(lines []string) []string {
	var rows []benchmarkRow
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "Benchmark") {
			rows = append(rows, benchmarkRow{lineIdx: i, fields: strings.Fields(line)})
		}
	}
	if len(rows) == 0 {
		return lines
	}

	maxCols := 0
	for _, r := range rows {
		if len(r.fields) > maxCols {
			maxCols = len(r.fields)
		}
	}

	// formattedCols[col][rowPos] holds the aligned cell for rows[rowPos], for the rows that have that column.
	formattedCols := make([][]string, maxCols)
	for col := 0; col < maxCols; col++ {
		formattedCols[col] = formatColumn(rows, col)
	}

	result := make([]string, len(lines))
	copy(result, lines)
	for rowPos, r := range rows {
		var b strings.Builder
		for col := range r.fields {
			if col > 0 {
				if isUnitColumn(col) {
					b.WriteByte(' ')
				} else {
					b.WriteString("   ")
				}
			}
			b.WriteString(formattedCols[col][rowPos])
		}
		result[r.lineIdx] = strings.TrimRight(b.String(), " ")
	}
	return result
}

// isUnitColumn reports whether col holds a unit (ns/op, B/op, ...) that sticks to the value column right before it
// with a single space. Benchmark lines are shaped as: name, iterations, then repeated (value, unit) pairs starting
// at index 2, so units sit at the odd indices from 3 on.
func isUnitColumn(col int) bool {
	return col >= 3 && col%2 == 1
}

// formatColumn aligns a single column across the given rows: numeric columns are right-aligned with thousands
// separators and aligned decimal points, other columns keep their original left alignment.
func formatColumn(rows []benchmarkRow, col int) []string {
	cells := make([]string, len(rows))

	numeric := true
	for _, r := range rows {
		if col >= len(r.fields) {
			continue
		}
		if !numberRe.MatchString(r.fields[col]) {
			numeric = false
			break
		}
	}

	if !numeric {
		maxWidth := 0
		for _, r := range rows {
			if col < len(r.fields) && len(r.fields[col]) > maxWidth {
				maxWidth = len(r.fields[col])
			}
		}
		for i, r := range rows {
			if col < len(r.fields) {
				cells[i] = fmt.Sprintf("%-*s", maxWidth, r.fields[col])
			}
		}
		return cells
	}

	intParts := make([]string, len(rows))
	fracParts := make([]string, len(rows))
	maxInt, maxFrac := 0, 0
	for i, r := range rows {
		if col >= len(r.fields) {
			continue
		}
		intPart, fracPart, _ := strings.Cut(r.fields[col], ".")
		intParts[i] = addThousandsSeparators(strings.ReplaceAll(intPart, ",", ""))
		fracParts[i] = fracPart
		if len(intParts[i]) > maxInt {
			maxInt = len(intParts[i])
		}
		if len(fracPart) > maxFrac {
			maxFrac = len(fracPart)
		}
	}

	for i, r := range rows {
		if col >= len(r.fields) {
			continue
		}
		cell := fmt.Sprintf("%*s", maxInt, intParts[i])
		if maxFrac > 0 {
			if fracParts[i] != "" {
				cell += "." + fracParts[i] + strings.Repeat(" ", maxFrac-len(fracParts[i]))
			} else {
				cell += strings.Repeat(" ", maxFrac+1)
			}
		}
		cells[i] = cell
	}
	return cells
}

func addThousandsSeparators(digits string) string {
	var b strings.Builder
	n := len(digits)
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(digits[i])
	}
	return b.String()
}
