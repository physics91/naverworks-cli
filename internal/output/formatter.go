package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Formatter struct {
	format  string
	writer  io.Writer
	columns []string
	dataKey string
}

func NewFormatter(format string, writer io.Writer) *Formatter {
	return &Formatter{format: format, writer: writer}
}

func (f *Formatter) WithTable(columns []string, dataKey string) *Formatter {
	f.columns = columns
	f.dataKey = dataKey
	return f
}

func (f *Formatter) Print(data interface{}) {
	encoded, _ := json.MarshalIndent(data, "", "  ")
	fmt.Fprintln(f.writer, string(encoded))
}

func (f *Formatter) PrintRaw(data []byte) {
	if f.format == "table" && len(f.columns) > 0 {
		f.printAsTable(data)
		return
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err == nil {
		buf.WriteByte('\n')
		buf.WriteTo(f.writer)
	} else {
		fmt.Fprintln(f.writer, string(data))
	}
}

func (f *Formatter) printAsTable(data []byte) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Fprintln(f.writer, string(data))
		return
	}
	arrayData, ok := raw[f.dataKey]
	if !ok {
		fmt.Fprintln(f.writer, string(data))
		return
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(arrayData, &items); err != nil {
		fmt.Fprintln(f.writer, string(data))
		return
	}

	rows := make([][]string, 0, len(items))
	for _, item := range items {
		row := make([]string, len(f.columns))
		for i, col := range f.columns {
			if v, ok := item[col]; ok {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		rows = append(rows, row)
	}
	f.PrintTable(f.columns, rows)
}

func (f *Formatter) PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	dashes := make([]string, len(headers))
	for i, h := range headers {
		dashes[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(w, strings.Join(dashes, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}
