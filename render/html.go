package render

import (
	sqldriver "database/sql/driver"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/go-mysql-org/go-mysql/mysql"
)

type HTMLRender struct {
	writer io.WriteCloser
}

func (c *HTMLRender) Open(file string) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	c.writer = f

	c.output(`
<style>
* {
	font-family: monospace;
}

table, th, td {
  border: 1px solid black;
  border-collapse: collapse;
}

th, td {
	padding: 5px;
	font-size: 16px;
}

.result-table {
	
}
</style>
`)

	return nil
}

func (c *HTMLRender) Close() error {
	if c.writer != nil {
		return c.writer.Close()
	}
	return nil
}

func (c *HTMLRender) output(format string, args ...interface{}) {
	_, err := fmt.Fprintf(c.writer, format, args...)
	if err != nil {
		fmt.Println("Write HTML file failed", err)
	}
}

func (c *HTMLRender) Push(frame *Frame) {
	c1, c2 := c.diffResult(frame.MySQL.Error, frame.TiDB.Error, frame.MySQL.Result, frame.TiDB.Result)
	var argStr string
	if len(frame.Args) > 0 {
		argStr = strings.Join(FormatArgs(frame.Args), ", ")
		argStr = fmt.Sprintf(" (%s)", argStr)
	}
	if c1 == c2 {
		c.output("<h2 style='color:green'>%s [MySQL %s, TiDB %s]<h2>", frame.Ident, frame.MySQL.Duration, frame.TiDB.Duration)
	} else {
		c.output("<h2 style='color:red'>%s [MySQL %s, TiDB %s]<h2>", frame.Ident, frame.MySQL.Duration, frame.TiDB.Duration)
		c.output("<h3>%s MySQL > %s%s </h3>\n%s", frame.Ident, frame.Query, argStr, c1)
		c.output("<h3>%s TiDB > %s%s </h3>\n%s", frame.Ident, frame.Query, argStr, c2)
	}

	c.output("<br><hr>")
}

func (c *HTMLRender) diffResult(myErr error, tiErr error, myResult, tiResult *mysql.Result) (mysqlContent, tidbContent string) {
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

	mysqlContent, tidbContent = c.prettyHTML(mysqlResult), c.prettyHTML(tidbResult)
	green := func(a ...interface{}) string {
		return fmt.Sprintf("<span style='color:green; font-weight: bolder'>%s</span>", a[0])
	}
	red := func(a ...interface{}) string {
		return fmt.Sprintf("<span style='color:red; font-weight: bolder'>%s</span>", a[0])
	}

	mysqlContent, tidbContent = genDiffResult(mysqlContent, tidbContent, red, green)

	return
}

func (HTMLRender) prettyHTML(r *rows) string {
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

	var lines []string
	var push = func(line string) {
		lines = append(lines, "<tr>"+line+"</tr>\n")
	}

	// Write table header
	var header string
	for _, col := range cols {
		header += "<th>" + col + "</th>\n"
	}
	push(strings.ToLower(header))

	// Write rows data
	for _, row := range allRows {
		var line string
		for _, col := range row {
			line += "<td>" + string(col) + "</td>\n"
		}
		push(line)
	}
	return "<table class='result-table'>" + strings.Join(lines, "\n") + "</table>"
}
