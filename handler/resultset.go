package handler

import (
	sqldriver "database/sql/driver"
	"fmt"
	"io"
	"strings"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/siddontang/go/hack"
)

type rows struct {
	*mysql.Resultset

	columns []string
	step    int
}

func newRows(r *mysql.Resultset) (*rows, error) {
	if r == nil {
		return nil, fmt.Errorf("invalid mysql query, no correct result")
	}

	rs := new(rows)
	rs.Resultset = r

	rs.columns = make([]string, len(r.Fields))

	for i, f := range r.Fields {
		rs.columns[i] = hack.String(f.Name)
	}
	rs.step = 0

	return rs, nil
}

func (r *rows) Columns() []string {
	return r.columns
}

func (r *rows) Close() error {
	r.step = -1
	return nil
}

func (r *rows) Next(dest []sqldriver.Value) error {
	if r.step >= r.Resultset.RowNumber() {
		return io.EOF
	} else if r.step == -1 {
		return io.ErrUnexpectedEOF
	}

	for i := 0; i < r.Resultset.ColumnNumber(); i++ {
		value, err := r.Resultset.GetValue(r.step, i)
		if err != nil {
			return err
		}

		dest[i] = sqldriver.Value(value)
	}

	r.step++

	return nil
}

func (r *rows) PrettyText() string {
	cols := r.columns
	var allRows [][]string
	for {
		dest := make([]sqldriver.Value, len(cols))
		if err := r.Next(dest); err == io.EOF {
			break
		}
		var row []string
		for _, c := range dest {
			if c == nil {
				row = append(row, "NULL")
			} else {
				// Ref: https://github.com/go-mysql-org/go-mysql/blob/33ea963610607f7b5505fd39d0955b78039ef783/mysql/field.go#L199
				// Only four types need to be asserted.
				switch x := c.(type) {
				case uint64, int64:
					row = append(row, fmt.Sprintf("%d", x))
				case float64:
					row = append(row, fmt.Sprintf("%f", x))
				case string:
					row = append(row, x)
				default:
					row = append(row, fmt.Sprintf("%s", c))
				}
			}
		}
		allRows = append(allRows, row)
	}

	// Calculate the max column length
	var colLength []int
	for _, c := range cols {
		colLength = append(colLength, len(c))
	}
	for _, row := range allRows {
		for n, col := range row {
			if l := len(col); colLength[n] < l {
				colLength[n] = l
			}
		}
	}
	// The total length
	var total = len(cols) - 1
	for index := range colLength {
		colLength[index] += 2 // Value will wrap with space
		total += colLength[index]
	}

	var lines []string
	var push = func(line string) {
		lines = append(lines, line)
	}

	// Write table header
	var header string
	for index, col := range cols {
		length := colLength[index]
		padding := length - 1 - len(col)
		if index == 0 {
			header += "|"
		}
		header += " " + col + strings.Repeat(" ", padding) + "|"
	}
	splitLine := "+" + strings.Repeat("-", total) + "+"
	push(splitLine)
	push(header)
	push(splitLine)

	// Write rows data
	for _, row := range allRows {
		var line string
		for index, col := range row {
			length := colLength[index]
			padding := length - 1 - len(col)
			if index == 0 {
				line += "|"
			}
			line += " " + string(col) + strings.Repeat(" ", padding) + "|"
		}
		push(line)
	}
	push(splitLine)
	return strings.Join(lines, "\n")
}
