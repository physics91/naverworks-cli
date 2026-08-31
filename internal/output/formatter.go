package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
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

// PrintRaw writes a raw JSON response. Every command prints through here, so
// this is where presigned upload URLs are stripped — placing the check at
// individual call sites left the list and pagination paths unprotected.
func (f *Formatter) PrintRaw(data []byte) {
	data = SanitizeJSON(data)

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
		for i, column := range f.columns {
			_, path := tableColumn(column)
			if value, ok := tableValue(item, path); ok && value != nil {
				switch value.(type) {
				case map[string]interface{}, []interface{}:
					if encoded, err := json.Marshal(value); err == nil {
						row[i] = string(encoded)
					}
				default:
					row[i] = fmt.Sprintf("%v", value)
				}
			}
		}
		rows = append(rows, row)
	}
	headers := make([]string, len(f.columns))
	for i, column := range f.columns {
		headers[i], _ = tableColumn(column)
	}
	f.PrintTable(headers, rows)
}

func tableColumn(column string) (header, path string) {
	header, path, found := strings.Cut(column, "=")
	if !found || header == "" || path == "" {
		return column, column
	}
	return header, path
}

func tableValue(item map[string]interface{}, path string) (interface{}, bool) {
	var current interface{} = item
	for _, segment := range strings.Split(path, ".") {
		switch value := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = value[segment]
			if !ok {
				return nil, false
			}
		case []interface{}:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(value) {
				return nil, false
			}
			current = value[index]
		default:
			return nil, false
		}
	}
	return current, true
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
