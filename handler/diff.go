package handler

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/fatih/color"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func diffError(query string, myErr, tiErr error) bool {
	if myErr == tiErr {
		return true
	}
	color.Red("QUERY >\t %s", query)
	color.Yellow("TiDB  >\t %v", myErr)
	color.Yellow("MySQL >\t %v", tiErr)
	return false
}

func diffResult(query string, myResult, tiResult *mysql.Result) bool {
	eq := reflect.DeepEqual(myResult.Resultset, tiResult.Resultset)
	if eq {
		return true
	}

	mysqlResult, _ := newRows(myResult.Resultset)
	tidbResult, _ := newRows(tiResult.Resultset)
	defer mysqlResult.Close()
	defer tidbResult.Close()

	mysqlContent, tidbContent := mysqlResult.PrettyText(), tidbResult.PrettyText()
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

	if mysqlContent == tidbContent {
		return true
	}

	color.Red("QUERY >\t %s", query)
	color.Yellow("TiDB  >")
	fmt.Println(tidbContent)
	color.Yellow("MySQL >")
	fmt.Println(mysqlContent)

	return false
}
