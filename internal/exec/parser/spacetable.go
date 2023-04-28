package parser

import (
	"bufio"
	"strings"
	"unicode"
)

type Table struct {
	Headers []string
	Rows    [][]string
}

type TableSpaceSeparatedOutput struct {
	Table Table
	Lines []string
}

// TableSpaceSeparated takes a string input and returns a slice of slices containing the separated values in each row
// and a slice of the original input lines.
func TableSpaceSeparated(in string) TableSpaceSeparatedOutput {
	var out TableSpaceSeparatedOutput
	in = replaceTabsWithSpaces(in)
	scanner := bufio.NewScanner(strings.NewReader(in))

	// Parse the headers
	var separators []int
	if scanner.Scan() {
		line := scanner.Text()
		separators = getSeparators(line)
		out.Lines = append(out.Lines, line)
		out.Table.Headers = splitIntoCells(line, separators)
	}
	// Parse the rows
	for scanner.Scan() {
		line := scanner.Text()
		out.Lines = append(out.Lines, line)

		row := splitIntoCells(line, separators)
		out.Table.Rows = append(out.Table.Rows, row)
	}
	return out
}

func replaceTabsWithSpaces(in string) string {
	return strings.ReplaceAll(in, "\t", "  ")
}

// function takes a line and returns a list of separators (positions of left edges of the cells)
func getSeparators(line string) []int {
	var separators []int
	for idx, ch := range line {
		isCurrentCharSpace := unicode.IsSpace(ch)
		if !isCurrentCharSpace { // not separator
			continue
		}

		var (
			previousIdx = decreaseWithMin(idx, 0)
			nextIdx     = increaseWithMax(idx, len(line))

			isNextSpace  = unicode.IsSpace(rune(line[nextIdx]))
			wasPrevSpace = unicode.IsSpace(rune(line[previousIdx]))
		)

		if isCurrentCharSpace && isNextSpace {
			continue
		}

		if isCurrentCharSpace && !wasPrevSpace && !isNextSpace { // check for multi world colum name like "APP VERSION"
			continue
		}
		separators = append(separators, idx)
	}
	return separators
}

func increaseWithMax(in, max int) int {
	in++
	if in > max {
		return max
	}
	return in
}

func decreaseWithMin(in, min int) int {
	in--
	if in < min {
		return min
	}
	return in
}

// function takes a line and a list of separators and returns a list of cells (the line divided by the separators)
func splitIntoCells(line string, separators []int) []string {
	var (
		res   []string
		start = 0
	)

	separators = append(separators, len(line)) // to add the final "end", otherwise the last 'cell' won't be extracted
	for _, end := range separators {
		if end > len(line) {
			end = len(line)
		}
		cell := strings.TrimSpace(line[start:end])
		start = end
		res = append(res, cell)
	}

	return res
}
