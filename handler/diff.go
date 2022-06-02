package handler

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/fatih/color"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func diffResult(myErr error, tiErr error, myResult, tiResult *mysql.Result) (mysqlContent, tidbContent string) {
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

	mysqlContent, tidbContent = mysqlResult.PrettyText(), tidbResult.PrettyText()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	patch := diffmatchpatch.New()
	diff := patch.DiffMain(mysqlContent, tidbContent, false)
	var newMySQLContent, newTiDBContent bytes.Buffer
	for _, d := range diff {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			newMySQLContent.WriteString(d.Text)
			newTiDBContent.WriteString(d.Text)
		case diffmatchpatch.DiffDelete:
			newMySQLContent.WriteString(red(d.Text))
		case diffmatchpatch.DiffInsert:
			newTiDBContent.WriteString(green(d.Text))
		}
	}
	mysqlContent = newMySQLContent.String()
	tidbContent = newTiDBContent.String()

	return
}
