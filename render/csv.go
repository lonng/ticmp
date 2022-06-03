package render

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/go-mysql-org/go-mysql/mysql"
)

type CSVRender struct {
	csvWriter *csv.Writer
	w         io.WriteCloser
}

func (c *CSVRender) Open(file string) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	c.w = f
	c.csvWriter = csv.NewWriter(f)

	c.csvWriter.Write([]string{"IsDiff", "Query", "Args", "MySQL Time", "TiDB Time", "MySQL", "TiDB"})
	c.csvWriter.Flush()
	return nil
}

func (r *CSVRender) Close() error {
	r.csvWriter.Flush()
	return r.w.Close()
}

func (c *CSVRender) Push(frame *Frame) {
	c1, c2 := c.diffResult(frame.MySQL.Error, frame.TiDB.Error, frame.MySQL.Result, frame.TiDB.Result)

	var argStr string
	if len(frame.Args) > 0 {
		argStr = strings.Join(FormatArgs(frame.Args), ", ")
	}

	records := make([]string, 7)
	if c1 == c2 {
		records[0] = "NO"
	} else {
		records[0] = "YES"
		records[5] = c1
		records[6] = c2
	}
	records[1] = frame.Query
	records[2] = argStr
	records[3] = fmt.Sprintf("%v", frame.MySQL.Duration.String())
	records[4] = fmt.Sprintf("%v", frame.TiDB.Duration.String())

	c.csvWriter.Write(records)
	c.csvWriter.Flush()

}

func (c *CSVRender) diffResult(myErr error, tiErr error, myResult, tiResult *mysql.Result) (mysqlContent, tidbContent string) {
	if myErr != tiErr {
		mysqlContent = fmt.Sprintf("%s", myErr)
		tidbContent = fmt.Sprintf("%s", tiErr)
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
	noColor := func(a ...interface{}) string {
		return a[0].(string)
	}

	mysqlContent, tidbContent = genDiffResult(mysqlContent, tidbContent, noColor, noColor)

	return
}
