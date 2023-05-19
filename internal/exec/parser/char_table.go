package parser

import (
	"fmt"
	"strings"

	"github.com/sanity-io/litter"
	"k8s.io/utils/strings/slices"
)

type TableChar struct {
	char string
}

func NewTableChar() *TableChar {
	return &TableChar{
		char: "|",
	}
}

// TableSeparated takes a string input and returns a slice of slices containing the separated values in each row
// and a slice of the original input lines.
// TODO: change the output to a JSON or YAML format to allow standardized parser interface.
func (t *TableChar) TableSeparated(in string) TableOutput {
	fmt.Println("input!!")
	fmt.Println(in)
	var out TableOutput

	lines := strings.Split(in, "\n")

	lines = slices.Filter(nil, lines, func(s string) bool {
		return strings.TrimSpace(s) != ""
	})
	litter.Dump(lines)
	out.Table.Headers = t.parseHeaders(lines[0])
	out.Lines = append(out.Lines, lines[0])
	for i := 2; i < len(lines); i++ {
		out.Lines = append(out.Lines, lines[i])
		record := t.parseRecord(lines[i], out.Table.Headers)

		if len(record) == 0 {
			continue
		}
		if record[0] != "" {
			out.Table.Rows = append(out.Table.Rows, record)
			continue
		}
		lastItem := len(out.Table.Rows) - 1
		for idx, val := range record {
			if val == "" {
				continue
			}
			out.Table.Rows[lastItem][idx] = fmt.Sprintf("%s %s", out.Table.Rows[lastItem][idx], val)
		}

	}

	litter.Dump(out)
	return out
}

func (t *TableChar) parseHeaders(headerLine string) []string {
	var headers []string
	cols := strings.Split(headerLine, t.char)
	fmt.Println("cols")
	fmt.Println(cols)
	for _, col := range cols {
		header := strings.TrimSpace(col)
		header = strings.ReplaceAll(header, " ", "_")
		if header != "" {
			headers = append(headers, header)
		}
	}

	return headers
}

func (t *TableChar) parseRecord(recordLine string, headers []string) []string {
	var record []string
	cols := strings.Split(recordLine, t.char)
	for _, col := range cols {
		//if i >= len(headers) {
		//	break
		//}
		record = append(record, strings.TrimSpace(col))
	}

	return record
}
