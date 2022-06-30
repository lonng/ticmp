package render

import (
	"bytes"
	sqldriver "database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type ConsoleRender struct{}

func (c ConsoleRender) Push(frame *Frame) {
	c1, c2 := c.diffResult(frame.MySQL.Error, frame.TiDB.Error, frame.MySQL.Result, frame.TiDB.Result)
	var argStr string
	if len(frame.Args) > 0 {
		argStr = strings.Join(FormatArgs(frame.Args), ", ")
		argStr = fmt.Sprintf("(%s)", argStr)
	}
	if c1 == c2 {
		color.Green("%s [MySQL %s, TiDB %s] ==> %s (%s)", frame.Ident, frame.MySQL.Duration, frame.TiDB.Duration, frame.Query, argStr)
	} else {
		color.Red("%s [MySQL %s, TiDB %s] ==> %s (%s)", frame.Ident, frame.MySQL.Duration, frame.TiDB.Duration, frame.Query, argStr)
		fmt.Printf("%s MySQL >\n%s\n", frame.Ident, c1)
		fmt.Printf("%s TiDB  >\n%s\n", frame.Ident, c2)
	}
}

type colorFunc func(a ...interface{}) string

func genDiffResult(mysqlContent string, tidbContent string,
	diffDelete colorFunc, diffInsert colorFunc) (string, string) {
	patch := diffmatchpatch.New()
	diff := patch.DiffMain(mysqlContent, tidbContent, false)
	var newMySQLContent, newTiDBContent bytes.Buffer
	for _, d := range diff {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			newMySQLContent.WriteString(d.Text)
			newTiDBContent.WriteString(d.Text)
		case diffmatchpatch.DiffDelete:
			newMySQLContent.WriteString(diffDelete(d.Text))
		case diffmatchpatch.DiffInsert:
			newTiDBContent.WriteString(diffInsert(d.Text))
		}
	}
	mysqlContent = newMySQLContent.String()
	tidbContent = newTiDBContent.String()
	return mysqlContent, tidbContent
}

func formatError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%s", err.Error())
}

func (c ConsoleRender) diffResult(myErr error, tiErr error, myResult, tiResult *mysql.Result) (mysqlContent, tidbContent string) {
	if myErr != tiErr {
		mysqlContent = formatError(myErr)
		tidbContent = formatError(tiErr)
		return
	}

	if reflect.DeepEqual(myResult.Resultset, tiResult.Resultset) {
		return "", ""
	}

	mysqlResult, _ := newRows(myResult.Resultset)
	tidbResult, _ := newRows(tiResult.Resultset)
	defer mysqlResult.Close()
	defer tidbResult.Close()

	mysqlContent, tidbContent = prettyText(mysqlResult), prettyText(tidbResult)
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	mysqlContent, tidbContent = genDiffResult(mysqlContent, tidbContent, red, yellow)

	return
}

func prettyText(r *rows) string {
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
	push(strings.ToLower(header))
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
