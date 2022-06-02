package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-mysql-org/go-mysql/client"
	_ "github.com/go-mysql-org/go-mysql/driver"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/server"
	"github.com/hashicorp/go-multierror"
	"github.com/lonng/ticomp/config"
)

type ShadowHandler struct {
	server.EmptyHandler
	cfg       *config.Config
	mysqlConn *client.Conn
	tidbConn  *client.Conn
	connIdent string
}

func NewShadowHandler(config *config.Config, connIdent string) *ShadowHandler {
	return &ShadowHandler{
		cfg:       config,
		connIdent: connIdent,
	}
}

func (h *ShadowHandler) Initialize() error {
	mycfg := h.cfg.MySQL
	mycon, err := client.Connect(fmt.Sprintf("%s:%d", mycfg.Host, mycfg.Port), mycfg.User, mycfg.Pass, mycfg.Name)
	if err != nil {
		return err
	}
	if err := mycon.Ping(); err != nil {
		return err
	}
	h.mysqlConn = mycon

	ticfg := h.cfg.TiDB
	ticon, err := client.Connect(fmt.Sprintf("%s:%d", ticfg.Host, ticfg.Port), ticfg.User, ticfg.Pass, ticfg.Name)
	if err != nil {
		return err
	}
	if err := ticon.Ping(); err != nil {
		return err
	}
	h.tidbConn = ticon

	return nil
}

func (h *ShadowHandler) Finalize() error {
	var result error
	if h.mysqlConn != nil {
		err := h.mysqlConn.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	if h.tidbConn != nil {
		err := h.tidbConn.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (h *ShadowHandler) UseDB(dbName string) error {
	var result error
	if err := h.mysqlConn.UseDB(dbName); err != nil {
		result = multierror.Append(result, err)
	}
	if err := h.tidbConn.UseDB(dbName); err != nil {
		result = multierror.Append(result, err)
	}
	return result
}

// HandleQuery overwrites the original HandleQuery.
func (h *ShadowHandler) HandleQuery(query string) (*mysql.Result, error) {
	start := time.Now()
	myResult, err1 := h.mysqlConn.Execute(query)
	myTime := time.Now().Sub(start).String()

	start = time.Now()
	tiResult, err2 := h.tidbConn.Execute(query)
	tiTime := time.Now().Sub(start).String()

	c1, c2 := diffResult(err1, err2, myResult, tiResult)
	if c1 == c2 {
		color.Green("%s [MySQL %s, TiDB %s] ==> %s", h.connIdent, myTime, tiTime, query)
	} else {
		color.Red("%s [MySQL %s, TiDB %s] ==> %s", h.connIdent, myTime, tiTime, query)
		color.Yellow("%s MySQL >\n%s", h.connIdent, c1)
		color.Yellow("%s TiDB  >\n%s", h.connIdent, c2)
	}

	return myResult, err1
}

func (h *ShadowHandler) HandleFieldList(table string, fieldWildcard string) ([]*mysql.Field, error) {
	myFields, err1 := h.mysqlConn.FieldList(table, fieldWildcard)

	// TODO(lonng): implement diff result for field list.
	_, _ = h.tidbConn.FieldList(table, fieldWildcard)

	return myFields, err1
}

func (h *ShadowHandler) HandleStmtPrepare(query string) (int, int, interface{}, error) {
	mystmt, err := h.mysqlConn.Prepare(query)

	// TODO(lonng): implement diff result for preparing.
	_, _ = h.tidbConn.Prepare(query)

	return mystmt.ParamNum(), mystmt.ColumnNum(), nil, err
}

func (h *ShadowHandler) HandleStmtExecute(context interface{}, query string, args []interface{}) (*mysql.Result, error) {
	start := time.Now()
	myResult, err1 := h.mysqlConn.Execute(query, args...)
	myTime := time.Now().Sub(start).String()

	start = time.Now()
	tiResult, err2 := h.tidbConn.Execute(query, args...)
	tiTime := time.Now().Sub(start).String()

	c1, c2 := diffResult(err1, err2, myResult, tiResult)
	argStr := strings.Join(FormatArgs(args), ", ")
	if c1 == c2 {
		color.Green("%s [MySQL %s, TiDB %s] ==> %s (%s)", h.connIdent, myTime, tiTime, query, argStr)
	} else {
		color.Red("%s [MySQL %s, TiDB %s] ==> %s (%s)", h.connIdent, myTime, tiTime, query, argStr)
		color.Yellow("%s MySQL >\n%s", h.connIdent, c1)
		color.Yellow("%s TiDB  >\n%s", h.connIdent, c2)
	}

	return myResult, err1
}
