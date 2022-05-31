package handler

import (
	"fmt"

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
}

func NewShadowHandler(config *config.Config) *ShadowHandler {
	return &ShadowHandler{
		cfg: config,
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

// HandleQuery overwrites the original HandleQuery.
func (h *ShadowHandler) HandleQuery(query string) (*mysql.Result, error) {
	fmt.Println("------")

	myResult, err1 := h.mysqlConn.Execute(query)
	tiResult, err2 := h.tidbConn.Execute(query)

	errEq := diffError(query, err1, err2)
	resEq := errEq && diffResult(query, myResult, tiResult)

	if errEq && resEq {
		color.Green("QUERY >\t %s", query)
	}

	return myResult, err1
}
